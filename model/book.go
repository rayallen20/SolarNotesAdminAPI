package model

import (
	"SolarNotesAdminAPI/db"
	"context"
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BookSortOccupiedErr 本错误表示同一天体下的书籍排序值已被其他记录占用
var BookSortOccupiedErr = errors.New("book sort is occupied")

// Book 本结构体用于映射 book 表,只包含本次创建书籍所需的字段.
type Book struct {
	ID            int64   `gorm:"column:id;primaryKey;autoIncrement:false"` // ID 书籍ID
	PlanetID      int64   `gorm:"column:planet_id"`                         // PlanetID 书籍所属天体的 ID
	Title         string  `gorm:"column:title"`                             // Title 书籍标题
	Sort          int32   `gorm:"column:sort"`                              // Sort 书籍在所属天体下的排序值
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
