package model

import (
	"SolarNotesAdminAPI/db"
	"SolarNotesAdminAPI/paginate"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BookSortOccupiedErr 本错误表示同一天体下的书籍排序值已被其他记录占用
var BookSortOccupiedErr = errors.New("book sort is occupied")

// BookNotFoundErr 本错误表示表示给定的书籍不存在
var BookNotFoundErr = errors.New("book not found")

// Book 本结构体用于映射 book 表,只包含本次创建书籍所需的字段.
type Book struct {
	ID            int64   `gorm:"column:id;primaryKey;autoIncrement:false"` // ID 书籍ID
	PlanetID      int64   `gorm:"column:planet_id"`                         // PlanetID 书籍所属天体的 ID
	Title         string  `gorm:"column:title"`                             // Title 书籍标题
	Sort          uint32  `gorm:"column:sort"`                              // Sort 书籍在所属天体下的排序值
	DisplayStatus string  `gorm:"column:display_status"`                    // DisplayStatus 书籍的展示状态
	Planet        *Planet `gorm:"-"`                                        // Planet 保存创建成功时读取的所属天体
}

// Insert 本方法用于在事务中确认关联天体存在,并插入接收者表示的书籍记录
// (事务提交成功才返回 nil,查询/插入/提交错误均向调用方传播)
func (b *Book) Insert(ctx context.Context) error {
	if b == nil {
		return errors.New("nil book model")
	}

	if ctx == nil {
		return errors.New("nil book insert context")
	}

	if db.DB == nil {
		return errors.New("database is not initialized")
	}

	if b.ID <= 0 {
		return errors.New("book ID must be positive")
	}

	if b.PlanetID <= 0 {
		return errors.New("book planet ID must be positive")
	}

	// 先在副本上执行事务, 提交成功后再向调用方公开查询结果.
	record := *b
	if err := db.DB.WithContext(ctx).Transaction(record.insertWithPlanet); err != nil {
		return err
	}
	*b = record
	return nil
}

// insertWithPlanet 本函数用于在 Insert 方法中使用的命名事务回调,全部数据库操作共用同一事务
// (关联天体的共享锁持有至事务结束,防止检查通过后天体在插入期间被删除)
func (b *Book) insertWithPlanet(tx *gorm.DB) error {
	if b == nil {
		return errors.New("nil book model")
	}

	if tx == nil {
		return errors.New("nil book insert transaction")
	}

	planet := &Planet{}

	err := tx.Select("id", "name", "title", "intro").
		Clauses(clause.Locking{
			Strength: "SHARE",
		}).
		Where("id = ?", b.PlanetID).
		Take(planet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return PlanetNotFoundErr
	}

	if err != nil {
		return err
	}

	result := tx.Create(b)
	err = result.Error
	if err != nil {
		// 1062 也可能来自主键冲突,不能直接解释为排序值已被占用
		var mysqlErr *mysql.MySQLError
		if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
			return err
		}

		// 在同一事务内执行当前读,确认所属天体与排序值的组合是否已存在
		// 使用新的记录接收结果,避免接收者的 ID 意外参与查询条件
		occupied := &Book{}
		findErr := tx.Select("id").
			Clauses(clause.Locking{
				Strength: "SHARE",
			}).
			Where("planet_id = ? AND sort = ?", b.PlanetID, b.Sort).
			Take(occupied).Error
		if findErr == nil {
			return errors.Join(BookSortOccupiedErr, err)
		}

		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			return err
		}

		// 复查失败时同时保留原始插入错误与查询错误,便于排查.
		return errors.Join(err, findErr)
	}

	if result.RowsAffected != 1 {
		return errors.New("book insert did not affect exactly one row")
	}

	b.Planet = planet
	return nil
}

// ListPage 本方法用于在一个只读事务中统计和读取书籍, 返回当前页以及分页前的总条数
func (b *Book) ListPage(ctx context.Context, page paginate.Page) ([]*Book, int64, error) {
	if b == nil || ctx == nil {
		return nil, 0, errors.New("nil book model or list context")
	}

	if !page.Valid() {
		return nil, 0, errors.New("invalid book list pagination")
	}

	if db.DB == nil {
		return nil, 0, errors.New("database is not initialized")
	}

	tx := db.DB.WithContext(ctx).Begin(&sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	// 提前返回或 panic 时回滚. 成功提交后, 延迟回滚不会改变已提交的结果
	defer tx.Rollback()

	var total int64
	if err := tx.Model(&Book{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	records := make([]*Book, 0)
	if int64(page.Offset()) < total {
		var err error
		records, err = b.readPage(tx, page)
		if err != nil {
			return nil, 0, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// readPage 本方法用于按天体 ID, 排序值和书籍 ID 读取一页, 再补齐所属天体
func (b *Book) readPage(tx *gorm.DB, page paginate.Page) ([]*Book, error) {
	var records []*Book

	err := tx.Model(&Book{}).
		Select("id", "planet_id", "title", "sort", "display_status").
		Order("planet_id ASC").
		Order("sort ASC").
		Order("id ASC").
		Scopes(page.Paginate).
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	err = b.loadPlanets(tx, records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// loadPlanets 本方法用于在同一个事务内批量查询本页涉及的天体, 保留书籍的原顺序
func (b *Book) loadPlanets(tx *gorm.DB, records []*Book) error {
	if len(records) == 0 {
		return nil
	}

	planetIDs := make([]int64, 0, len(records))
	seenIDs := make(map[int64]struct{}, len(records))
	for _, record := range records {
		if record == nil {
			return errors.New("book list data is inconsistent: nil book record")
		}

		if _, exists := seenIDs[record.PlanetID]; exists {
			continue
		}

		seenIDs[record.PlanetID] = struct{}{}
		planetIDs = append(planetIDs, record.PlanetID)
	}

	planets := make([]*Planet, 0, len(planetIDs))
	if err := tx.Model(&Planet{}).
		Select("id", "name", "title", "intro").
		Where("id IN ?", planetIDs).
		Find(&planets).Error; err != nil {
		return err
	}

	planetsByID := make(map[int64]*Planet, len(planets))
	for _, planet := range planets {
		if planet == nil {
			return errors.New("book list data is inconsistent: nil planet record")
		}

		planetsByID[planet.ID] = planet
	}

	for _, record := range records {
		planet, exists := planetsByID[record.PlanetID]
		if !exists {
			return fmt.Errorf("book list data is inconsistent: book %d references missing planet %d", record.ID, record.PlanetID)
		}

		record.Planet = planet
	}

	return nil
}

// OptionsByPlanetID 本函数用于根据给定的天体 ID 查询该天体下的所有书籍选项, 并按排序值升序排列
func (b *Book) OptionsByPlanetID(ctx context.Context, planetID int64) ([]Book, error) {
	if b == nil || ctx == nil {
		return nil, errors.New("nil book model or options context")
	}

	if planetID <= 0 {
		return nil, errors.New("book options planet ID must be positive")
	}

	if db.DB == nil {
		return nil, errors.New("database is not initialized")
	}

	tx := db.DB.WithContext(ctx).Begin(&sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer tx.Rollback()

	// 先确认天体存在, 以区分无效引用和空书单
	var planet Planet
	err := tx.Select("id").Where("id = ?", planetID).Take(&planet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, PlanetNotFoundErr
		}

		return nil, err
	}

	var books []Book
	err = tx.Model(&Book{}).
		Select("id", "title", "sort", "display_status").
		Where("planet_id = ?", planetID).
		Order("sort ASC").
		Order("id ASC").
		Find(&books).Error
	if err != nil {
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return books, nil
}
