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

// Book 本结构体用于表示书籍业务实体
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

	if ctx == nil || req == nil || req.Book == nil || req.Book.Planet == nil {
		return &sys_error.InternalError{Cause: errors.New("missing book creation arguments")}
	}

	input := req.Book
	record := &model.Book{
		ID:            snowflake.Next(),
		PlanetID:      input.Planet.ID,
		Title:         string(input.Title),
		DisplayStatus: input.DisplayStatus,
	}

	err := record.Insert(ctx)
	if err != nil {
		switch {
		case errors.Is(err, model.PlanetNotFoundErr):
			return sys_error.BookPlanetNotFoundErr
		case errors.Is(err, model.BookOrderInvalidErr):
			return &sys_error.InternalError{Cause: err}
		default:
			return &sys_error.DatabaseError{Cause: err}
		}
	}

	b.fillByORM(ctx, record)
	return nil
}

// Update 编辑书籍并在提交成功后将书籍摘要填入接收者.
func (b *Book) Update(ctx context.Context, req *reqBook.UpdateReq) error {
	if b == nil {
		return &sys_error.InternalError{Cause: errors.New("biz.Book.Update: nil receiver")}
	}
	if ctx == nil {
		return &sys_error.InternalError{Cause: errors.New("biz.Book.Update: nil context")}
	}
	if req == nil || req.Book == nil {
		return &sys_error.InternalError{Cause: errors.New("biz.Book.Update: missing book request")}
	}
	input := req.Book
	if input.Expected == nil || input.Expected.Title == nil {
		return &sys_error.InternalError{Cause: errors.New("biz.Book.Update: missing expected book values")}
	}
	if input.Target == nil {
		return &sys_error.InternalError{Cause: errors.New("biz.Book.Update: missing target book values")}
	}

	record := &model.Book{
		ID:            input.ID,
		Title:         string(input.Target.Title),
		DisplayStatus: input.Target.DisplayStatus,
	}
	err := record.Update(ctx, *input.Expected.Title, input.Expected.DisplayStatus)
	if err != nil {
		switch {
		case errors.Is(err, model.BookNotFoundErr):
			return sys_error.BookNotFoundErr
		case errors.Is(err, model.BookEditChangedErr):
			return sys_error.BookEditChangedErr
		default:
			return &sys_error.DatabaseError{Cause: err}
		}
	}

	*b = Book{
		ID:            record.ID,
		Title:         record.Title,
		Sort:          record.Sort,
		DisplayStatus: record.DisplayStatus,
	}
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

// Reorder 接收完整请求并返回提交后的书籍摘要.
func (b *Book) Reorder(ctx context.Context, req *reqBook.ReorderReq) ([]Book, error) {
	if b == nil || ctx == nil {
		return nil, &sys_error.InternalError{Cause: errors.New("biz.Book.Reorder: missing receiver or context")}
	}
	if req == nil || req.Planet == nil {
		return nil, &sys_error.InternalError{Cause: errors.New("biz.Book.Reorder: missing planet request")}
	}

	input := req.Planet
	orm := &model.Book{}
	rows, err := orm.Reorder(ctx, input.ID, input.Expected.IDs(), input.Target.IDs())
	if err != nil {
		switch {
		case errors.Is(err, model.BookOrderInputErr):
			return nil, sys_error.BookOrderInputErr
		case errors.Is(err, model.PlanetNotFoundErr):
			return nil, sys_error.BookPlanetNotFoundErr
		case errors.Is(err, model.BookOrderChangedErr):
			return nil, sys_error.BookOrderChangedErr
		default:
			return nil, &sys_error.DatabaseError{Cause: err}
		}
	}

	return bookSummaries(rows), nil
}

// Options 本函数用于展示给定天体下的书籍选项,用于前端下拉选择
func (b *Book) Options(ctx context.Context, req *reqBook.OptionsReq) ([]Book, error) {
	if b == nil {
		err := &sys_error.InternalError{Cause: errors.New("biz.Book.Options: receiver b (*biz.Book) is nil")}
		return nil, err
	}

	if ctx == nil {
		err := &sys_error.InternalError{Cause: errors.New("biz.Book.Options: request context (ctx) is nil")}
		return nil, err
	}
	if req == nil {
		err := &sys_error.InternalError{Cause: errors.New("biz.Book.Options: options request (req) is nil")}
		return nil, err
	}
	if req.Planet == nil {
		err := &sys_error.InternalError{Cause: errors.New("biz.Book.Options: planet reference (req.Planet) is nil")}
		return nil, err
	}

	orm := &model.Book{}
	records, err := orm.OptionsByPlanetID(ctx, req.Planet.ID)
	if err != nil {
		if errors.Is(err, model.PlanetNotFoundErr) {
			return nil, sys_error.BookPlanetNotFoundErr
		}

		err = &sys_error.DatabaseError{Cause: err}
		return nil, err
	}

	return bookSummaries(records), nil
}

// bookSummaries 按模型查询或调序后的顺序填入书籍摘要.
func bookSummaries(records []model.Book) []Book {
	books := make([]Book, len(records))
	for i, record := range records {
		books[i] = Book{
			ID:            record.ID,
			Title:         record.Title,
			Sort:          record.Sort,
			DisplayStatus: record.DisplayStatus,
		}
	}
	return books
}
