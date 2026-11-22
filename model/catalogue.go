package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"SolarNotesAdminAPI/db"
	"SolarNotesAdminAPI/enum"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	// CatalogueNotFoundErr 表示父目录不存在或已删除.
	CatalogueNotFoundErr = errors.New("catalogue not found")
	// CatalogueParentNotFolderErr 表示请求指定的父节点是文件.
	CatalogueParentNotFolderErr = errors.New("catalogue parent must be a folder")
	// CatalogueOrderChangedErr 表示父目录归属或同级排序快照已变化.
	CatalogueOrderChangedErr = errors.New("catalogue order changed")
	// CatalogueOwnerChangedErr 表示预查的所属书籍与锁定节点的当前归属不同.
	CatalogueOwnerChangedErr = errors.New("catalogue ownership changed during display update")
	// CatalogueCycleErr 表示沿父子关系遍历时再次遇到了已访问的节点.
	CatalogueCycleErr = errors.New("cycle in catalogue subtree")
)

// Catalogue 映射 catalogue 表
type Catalogue struct {
	ID            int64   `gorm:"column:id;primaryKey;autoIncrement:false"` // ID 目录节点主键
	BookID        int64   `gorm:"column:book_id"`                           // BookID 所属书籍
	ParentID      int64   `gorm:"column:parent_id"`                         // ParentID 直接父节点 ID, 根节点为 0
	Name          string  `gorm:"column:name"`                              // Name 节点名称
	Intro         string  `gorm:"column:intro"`                             // Intro 节点介绍
	Type          string  `gorm:"column:type"`                              // Type 节点类型
	Sort          *uint32 `gorm:"column:sort"`                              // Sort 节点排序(nil 表示待排序)
	DisplayStatus string  `gorm:"column:display_status"`                    // DisplayStatus 节点显隐状态
}

// ReadBookTree 在同一个只读快照内,读取给定书籍和该书全部目录行
func (c *Catalogue) ReadBookTree(ctx context.Context, bookID int64) (*Book, []Catalogue, error) {
	if c == nil {
		err := errors.New("model.Catalogue.ReadBookTree: nil receiver")
		return nil, nil, err
	}

	if ctx == nil {
		err := errors.New("model.Catalogue.ReadBookTree: nil context")
		return nil, nil, err
	}

	if bookID <= 0 {
		err := errors.New("model.Catalogue.ReadBookTree: book ID must be positive")
		return nil, nil, err
	}

	if db.DB == nil {
		err := errors.New("model.Catalogue.ReadBookTree: database is not initialized")
		return nil, nil, err
	}

	tx := db.DB.WithContext(ctx).Begin(&sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	defer tx.Rollback()

	var book Book
	err := tx.Select("id", "title", "sort", "display_status").
		Where("id = ?", bookID).Take(&book).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, BookNotFoundErr
		}

		return nil, nil, err
	}

	var rows []Catalogue
	err = tx.Model(&Catalogue{}).
		Select("id", "book_id", "parent_id", "name", "intro", "type", "sort", "display_status").
		Where("book_id = ?", bookID).Find(&rows).Error
	if err != nil {
		return nil, nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, nil, err
	}

	return &book, rows, nil
}

// ReorderChildren 锁定直接子节点, 由 plan 按 ID 返回有变更的目标状态, 再原子提交排序.
func (c *Catalogue) ReorderChildren(ctx context.Context, parentID int64, plan func([]Catalogue) (map[int64]Catalogue, error)) error {
	if c == nil {
		return errors.New("model.Catalogue.ReorderChildren: nil receiver")
	}
	if ctx == nil {
		return errors.New("model.Catalogue.ReorderChildren: nil context")
	}
	if parentID <= 0 {
		return errors.New("model.Catalogue.ReorderChildren: invalid parent ID")
	}
	if plan == nil {
		return errors.New("model.Catalogue.ReorderChildren: nil plan")
	}
	if db.DB == nil {
		return errors.New("database is not initialized")
	}

	conn := db.DB.WithContext(ctx)
	var parent Catalogue
	// 此次普通读只定位要锁的书. 所有决策都依据事务内的当前读.
	err := conn.Select("id", "book_id").Where("id = ?", parentID).Take(&parent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return CatalogueNotFoundErr
	}
	if err != nil {
		return err
	}

	tx := conn.Begin(&sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	rows, err := parent.lockChildren(tx)
	if err != nil {
		return err
	}
	changes, err := plan(rows)
	if err != nil {
		return err
	}
	err = parent.writeChildOrder(tx, rows, changes)
	if err != nil {
		return err
	}
	return tx.Commit().Error
}

// lockChildren 依次锁定书籍, 父节点和该书中它的全部直接子节点.
func (c *Catalogue) lockChildren(tx *gorm.DB) ([]Catalogue, error) {
	lock := clause.Locking{Strength: "UPDATE"}
	var book Book
	err := tx.Clauses(lock).Select("id").Where("id = ?", c.BookID).Take(&book).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, BookNotFoundErr
	}
	if err != nil {
		return nil, err
	}

	var parent Catalogue
	err = tx.Clauses(lock).Select("id", "book_id", "type").
		Where("id = ?", c.ID).Take(&parent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, CatalogueNotFoundErr
	}
	if err != nil {
		return nil, err
	}
	if parent.BookID != c.BookID {
		return nil, CatalogueOrderChangedErr
	}
	if parent.Type != enum.TypeFolder {
		return nil, CatalogueParentNotFolderErr
	}

	var rows []Catalogue
	err = tx.Clauses(lock).Select("id", "book_id", "parent_id", "sort", "display_status").
		Where("book_id = ? AND parent_id = ?", c.BookID, c.ID).
		Order("sort ASC").Order("id ASC").Find(&rows).Error
	return rows, err
}

// writeChildOrder 先释放变更项占用的旧序号, 再分配最终序号, 避免同层唯一键冲突.
func (c *Catalogue) writeChildOrder(tx *gorm.DB, rows []Catalogue, changes map[int64]Catalogue) error {
	if len(changes) == 0 {
		return nil
	}
	for _, row := range rows {
		if _, changed := changes[row.ID]; !changed || row.Sort == nil {
			continue
		}
		if err := c.updateChildOrder(tx, row.ID, nil, enum.DisplayHidden); err != nil {
			return err
		}
	}
	for _, row := range rows {
		target, changed := changes[row.ID]
		if !changed {
			continue
		}
		if row.Sort != nil && target.Sort == nil {
			// 该节点已在释放旧序号时写为 NULL / hidden, 按锁定时的状态决定是否隐藏后代.
			if row.DisplayStatus == enum.DisplayVisible {
				if err := row.hideDescendants(tx); err != nil {
					return err
				}
			}
			continue
		}
		if err := c.updateChildOrder(tx, row.ID, target.Sort, target.DisplayStatus); err != nil {
			return err
		}
	}
	return nil
}

// updateChildOrder 更新一行已锁定且实际变更的子节点, nil 会写入 SQL NULL.
func (c *Catalogue) updateChildOrder(tx *gorm.DB, childID int64, rank *uint32, status string) error {
	result := tx.Model(&Catalogue{}).
		Where("book_id = ? AND parent_id = ? AND id = ?", c.BookID, c.ID, childID).
		Updates(map[string]any{"sort": rank, "display_status": status})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("catalogue %d order update affected %d rows", childID, result.RowsAffected)
	}
	return nil
}

// SetDisplayStatus 在一个事务中设置节点显隐, 从显示切为隐藏时同步隐藏全部后代.
func (c *Catalogue) SetDisplayStatus(ctx context.Context, check func(Catalogue, string) error) error {
	if c == nil {
		return errors.New("model.Catalogue.SetDisplayStatus: nil receiver")
	}
	if ctx == nil {
		return errors.New("model.Catalogue.SetDisplayStatus: nil context")
	}
	if check == nil {
		return errors.New("model.Catalogue.SetDisplayStatus: missing display rule")
	}
	if db.DB == nil {
		return errors.New("database is not initialized")
	}
	conn := db.DB.WithContext(ctx)
	var hint Catalogue
	err := conn.Select("book_id").Where("id = ?", c.ID).Take(&hint).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return CatalogueNotFoundErr
	}
	if err != nil {
		return err
	}

	tx := conn.Begin(&sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	lock := clause.Locking{Strength: "UPDATE"}
	var book Book
	bookErr := tx.Clauses(lock).Select("id").Where("id = ?", hint.BookID).Take(&book).Error
	if bookErr != nil && !errors.Is(bookErr, gorm.ErrRecordNotFound) {
		return bookErr
	}

	// 即使书已消失, 也要区分节点一起被删除和节点仍在但引用已失效.
	var node Catalogue
	err = tx.Clauses(lock).Select("id", "book_id", "parent_id", "type", "sort", "display_status").
		Where("id = ?", c.ID).Take(&node).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return CatalogueNotFoundErr
	}
	if err != nil {
		return err
	}
	if node.BookID != hint.BookID {
		return CatalogueOwnerChangedErr
	}
	if bookErr != nil {
		return BookNotFoundErr
	}

	err = check(node, c.DisplayStatus)
	if err != nil {
		return err
	}
	if node.DisplayStatus == c.DisplayStatus {
		return tx.Commit().Error
	}

	result := tx.Model(&Catalogue{}).Where("id = ? AND book_id = ?", node.ID, node.BookID).
		UpdateColumn("display_status", c.DisplayStatus)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("catalogue %d display update affected %d rows", node.ID, result.RowsAffected)
	}
	if node.DisplayStatus == enum.DisplayVisible && c.DisplayStatus == enum.DisplayHidden {
		err = node.hideDescendants(tx)
		if err != nil {
			return err
		}
	}
	return tx.Commit().Error
}

// hideDescendants 使用调用方的事务逐批隐藏同书后代. 接收者是已锁定的目标行.
func (c *Catalogue) hideDescendants(tx *gorm.DB) error {
	const batchSize = 256
	pending := []int64{c.ID}
	seen := map[int64]struct{}{c.ID: {}}
	lock := clause.Locking{Strength: "UPDATE"}

	for len(pending) > 0 {
		count := min(len(pending), batchSize)
		parentIDs := pending[:count]
		pending = pending[count:]

		// 查出所有直接子节点, 已隐藏的中间目录也可能有可见后代.
		var childIDs []int64
		err := tx.Model(&Catalogue{}).Clauses(lock).
			Where("book_id = ? AND parent_id IN ?", c.BookID, parentIDs).
			Order("id ASC").Pluck("id", &childIDs).Error
		if err != nil {
			return err
		}
		if len(childIDs) == 0 {
			continue
		}
		for _, id := range childIDs {
			if _, exists := seen[id]; exists {
				return fmt.Errorf("%w: book %d, node %d", CatalogueCycleErr, c.BookID, id)
			}
			seen[id] = struct{}{}
		}

		// 本批子节点可能都已隐藏, 更新零行也成功.
		err = tx.Model(&Catalogue{}).
			Where("book_id = ? AND parent_id IN ?", c.BookID, parentIDs).
			Where("display_status <> ?", enum.DisplayHidden).
			UpdateColumn("display_status", enum.DisplayHidden).Error
		if err != nil {
			return err
		}
		pending = append(pending, childIDs...)
	}
	return nil
}
