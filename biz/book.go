package biz

import (
	"context"
	"errors"

	"SolarNotesAdminAPI/model"
	reqBook "SolarNotesAdminAPI/param/req/v1/book"
	"SolarNotesAdminAPI/snowflake"
	"SolarNotesAdminAPI/sys_error"
)

// Book 承载书籍创建业务.
type Book struct {
	ID            int64   `json:"id"`             // ID 书籍ID
	Title         string  `json:"title"`          // Title 书籍标题
	Sort          int32   `json:"sort"`           // Sort 书籍在所属天体下的排序值
	DisplayStatus string  `json:"display_status"` // DisplayStatus 书籍的展示状态
	Planet        *Planet `json:"planet"`         // Planet 书籍所属天体
}

// Create 本函数用于创建书籍
func (b *Book) Create(ctx context.Context, req *reqBook.CreateReq) error {
	if b == nil {
		return &sys_error.InternalError{Cause: errors.New("nil book entity")}
	}

	if ctx == nil || req == nil || req.Book == nil || req.Book.Planet == nil || req.Book.Sort == nil {
		return &sys_error.InternalError{Cause: errors.New("missing book creation arguments")}
	}

	input := req.Book
	record := &model.Book{
		ID:            snowflake.Next(),
		PlanetID:      input.Planet.ID,
		Title:         string(input.Title),
		Sort:          *input.Sort,
		DisplayStatus: input.DisplayStatus,
	}

	err := record.Insert(ctx)
	if err != nil {
		switch {
		case errors.Is(err, model.PlanetNotFoundErr):
			return sys_error.BookPlanetNotFoundErr
		case errors.Is(err, model.BookSortOccupiedErr):
			return sys_error.BookSortOccupiedErr
		default:
			return &sys_error.DatabaseError{Cause: err}
		}
	}

	b.fillByORM(ctx, record)
	return nil
}

// fillByORM 本函数用于将书籍记录及其已加载的所属天体填入业务实体
func (b *Book) fillByORM(ctx context.Context, record *model.Book) {
	planet := &Planet{Exist: true}
	planet.fillByORM(ctx, record.Planet)

	*b = Book{
		ID:            record.ID,
		Title:         record.Title,
		Sort:          record.Sort,
		DisplayStatus: record.DisplayStatus,
		Planet:        planet,
	}
}
