package model

import (
	"context"
	"database/sql"
	"errors"

	"SolarNotesAdminAPI/db"

	"gorm.io/gorm"
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
