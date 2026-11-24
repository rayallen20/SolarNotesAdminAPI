package model

import (
	"SolarNotesAdminAPI/db"
	"SolarNotesAdminAPI/paginate"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BookOrderInvalidErr 表示已有书序不连续或下一序号已超出列的范围.
var BookOrderInvalidErr = errors.New("cannot append to book order")

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

// Insert 在事务中将新书追加到所属天体的书单末尾.
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

	// 在副本上分配序号并加载天体, 事务提交成功后再回填接收者.
	record := *b
	opts := &sql.TxOptions{Isolation: sql.LevelReadCommitted}
	if err := db.DB.WithContext(ctx).Transaction(record.insertWithPlanet, opts); err != nil {
		return err
	}
	*b = record
	return nil
}

// insertWithPlanet 先排他锁定天体, 使同一天体的并发创建在空书单上也能依次分配序号.
func (b *Book) insertWithPlanet(tx *gorm.DB) error {
	if b == nil {
		return errors.New("nil book model")
	}

	if tx == nil {
		return errors.New("nil book insert transaction")
	}

	planet := &Planet{}

	err := tx.Select("id", "name", "title", "intro").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", b.PlanetID).
		Take(planet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return PlanetNotFoundErr
	}

	if err != nil {
		return err
	}

	var order struct {
		BookCount uint64
		MinSort   uint64
		MaxSort   uint64
	}
	err = tx.Model(&Book{}).
		Select("COUNT(*) AS book_count, COALESCE(MIN(sort), 0) AS min_sort, COALESCE(MAX(sort), 0) AS max_sort").
		Where("planet_id = ?", b.PlanetID).
		Scan(&order).Error
	if err != nil {
		return err
	}

	// (planet_id, sort) 唯一, 因此 N 本书占据 1..N 时必然连续. 隐藏书籍也占一个位置.
	if order.BookCount > 0 && (order.MinSort != 1 || order.MaxSort != order.BookCount) {
		return fmt.Errorf("%w: planet %d has non-contiguous positions (count=%d, min=%d, max=%d)",
			BookOrderInvalidErr, b.PlanetID, order.BookCount, order.MinSort, order.MaxSort)
	}
	if order.BookCount >= math.MaxUint32 {
		return fmt.Errorf("%w: planet %d has exhausted its positions (count=%d, min=%d, max=%d)",
			BookOrderInvalidErr, b.PlanetID, order.BookCount, order.MinSort, order.MaxSort)
	}

	b.Sort = uint32(order.BookCount + 1)
	result := tx.Create(b)
	if result.Error != nil {
		return result.Error
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
