package biz

import (
	"context"
	"errors"
	"fmt"

	"SolarNotesAdminAPI/enum"
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

// Reorder 校验编排请求, 使用锁定后的同级目录行比较快照并保存新顺序.
func (c *Catalogue) Reorder(ctx context.Context, req *reqCatalogue.ReorderReq) error {
	if c == nil {
		return &sys_error.InternalError{Cause: errors.New("biz.Catalogue.Reorder: nil receiver")}
	}
	if ctx == nil {
		return &sys_error.InternalError{Cause: errors.New("biz.Catalogue.Reorder: nil context")}
	}
	if req == nil {
		return &sys_error.InternalError{Cause: errors.New("biz.Catalogue.Reorder: nil request")}
	}
	plan, err := newReorderPlan(req.Catalogue)
	if err != nil {
		return err
	}
	orm := &model.Catalogue{}
	err = orm.ReorderChildren(ctx, req.Catalogue.ParentID, plan.changes)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, model.CatalogueNotFoundErr):
		return sys_error.CatalogueNotFoundErr
	case errors.Is(err, model.BookNotFoundErr):
		return sys_error.BookNotFoundErr
	case errors.Is(err, model.CatalogueParentNotFolderErr):
		return fmt.Errorf("%w: parent_id must reference a folder", sys_error.CatalogueOrderInvalidErr)
	case errors.Is(err, model.CatalogueOrderChangedErr), errors.Is(err, sys_error.CatalogueOrderChangedErr):
		return sys_error.CatalogueOrderChangedErr
	case errors.Is(err, model.CatalogueCycleErr):
		return &sys_error.InternalError{Cause: err}
	default:
		return &sys_error.DatabaseError{Cause: err}
	}
}

// Display 在完整请求的基础上设置节点状态, 由控制器组装成功响应.
func (c *Catalogue) Display(ctx context.Context, req *reqCatalogue.DisplayReq) error {
	if c == nil {
		return &sys_error.InternalError{Cause: errors.New("biz.Catalogue.Display: nil receiver")}
	}
	if ctx == nil {
		return &sys_error.InternalError{Cause: errors.New("biz.Catalogue.Display: nil context")}
	}
	if req == nil {
		return &sys_error.InternalError{Cause: errors.New("biz.Catalogue.Display: nil request")}
	}
	if req.Catalogue == nil {
		return &sys_error.InternalError{Cause: errors.New("biz.Catalogue.Display: nil catalogue")}
	}

	orm := &model.Catalogue{ID: req.Catalogue.ID, DisplayStatus: req.Catalogue.DisplayStatus}
	err := orm.SetDisplayStatus(ctx, c.checkDisplayTarget)
	var internalErr *sys_error.InternalError
	switch {
	case err == nil:
		return nil
	case errors.Is(err, model.CatalogueNotFoundErr):
		return sys_error.CatalogueNotFoundErr
	case errors.Is(err, model.BookNotFoundErr):
		return sys_error.BookNotFoundErr
	case errors.Is(err, sys_error.CatalogueSortRequiredErr), errors.As(err, &internalErr):
		return err
	case errors.Is(err, model.CatalogueOwnerChangedErr), errors.Is(err, model.CatalogueCycleErr):
		return &sys_error.InternalError{Cause: err}
	default:
		return &sys_error.DatabaseError{Cause: err}
	}
}

// checkDisplayTarget 用锁定行的副本校验目标状态, 保留 DAO 中用于幂等判断的原值.
func (c *Catalogue) checkDisplayTarget(node model.Catalogue, target string) error {
	if node.ParentID != 0 && node.Sort == nil && target == enum.DisplayVisible {
		return sys_error.CatalogueSortRequiredErr
	}
	node.DisplayStatus = target
	err := validateCatalogueRow(&node)
	if err != nil {
		return &sys_error.InternalError{Cause: err}
	}
	return nil
}
