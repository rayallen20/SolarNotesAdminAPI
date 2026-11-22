package biz

import (
	"context"
	"errors"

	"SolarNotesAdminAPI/model"
	reqCatalogue "SolarNotesAdminAPI/param/req/v1/catalogue"
	"SolarNotesAdminAPI/sys_error"
)

// Catalogue 本结构体用于表示目录业务实体
type Catalogue struct {
	ID            int64        // ID 目录 ID
	Sort          *uint32      // Sort 排序字段(若节点为根节点,则该字段值为0)
	Type          string       // Type 目录类型 folder: 非叶节点; file: 叶子节点 (文件夹可以为空)
	Name          string       // Name 目录名称
	Intro         string       // Intro 目录介绍
	Parent        *Catalogue   // Parent 父节点(若节点为根节点,则该字段为nil)
	Children      []*Catalogue // Children 子节点列表
	DisplayStatus string       // DisplayStatus 展示状态
}

// Tree 本函数用于根据给定的书籍ID获取该书籍的目录树
func (c *Catalogue) Tree(ctx context.Context, req *reqCatalogue.TreeReq) (*Book, []*Catalogue, error) {
	if c == nil {
		err := &sys_error.InternalError{Cause: errors.New("biz.Catalogue.Tree: nil receiver")}
		return nil, nil, err
	}
	if ctx == nil {
		err := &sys_error.InternalError{Cause: errors.New("biz.Catalogue.Tree: nil context")}
		return nil, nil, err
	}

	if req == nil {
		err := &sys_error.InternalError{Cause: errors.New("biz.Catalogue.Tree: nil tree request")}
		return nil, nil, err
	}
	if req.Book == nil {
		err := &sys_error.InternalError{Cause: errors.New("biz.Catalogue.Tree: nil book reference")}
		return nil, nil, err
	}

	orm := &model.Catalogue{}
	book, rows, err := orm.ReadBookTree(ctx, req.Book.ID)
	if err != nil {
		if errors.Is(err, model.BookNotFoundErr) {
			return nil, nil, sys_error.BookNotFoundErr
		}

		return nil, nil, &sys_error.DatabaseError{Cause: err}
	}

	roots, err := buildCatalogueTree(book.ID, rows)
	if err != nil {
		err = &sys_error.InternalError{Cause: err}
		return nil, nil, err
	}

	summary := &Book{
		ID:            book.ID,
		Title:         book.Title,
		Sort:          book.Sort,
		DisplayStatus: book.DisplayStatus,
	}

	return summary, roots, nil
}
