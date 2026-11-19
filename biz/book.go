package biz

import (
	"SolarNotesAdminAPI/paginate"
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
	Sort          uint32  `json:"sort"`           // Sort 书籍在所属天体下的排序值
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

// List 本方法用于分页查询书籍列表,返回当前页的书籍记录,总记录数和总页数
func (b *Book) List(ctx context.Context, req *reqBook.ListReq) ([]*Book, int64, int64, error) {
	if b == nil || ctx == nil {
		return nil, 0, 0, &sys_error.InternalError{Cause: errors.New("nil book entity or context")}
	}

	if req == nil || req.Pagination == nil || req.Pagination.CurrentPage == nil || req.Pagination.ItemPerPage == nil {
		return nil, 0, 0, paginate.ErrInvalidPage
	}

	input := req.Pagination
	page, err := paginate.NewPage(*input.CurrentPage, *input.ItemPerPage)
	if err != nil {
		return nil, 0, 0, err
	}

	orm := &model.Book{}
	records, total, err := orm.ListPage(ctx, page)
	if err != nil {
		return nil, 0, 0, &sys_error.DatabaseError{Cause: err}
	}

	books := make([]*Book, 0, len(records))
	for _, record := range records {
		item := &Book{}
		item.fillByORM(ctx, record)

		books = append(books, item)
	}

	return books, total, page.TotalPages(total), nil
}
