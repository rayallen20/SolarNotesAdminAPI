package sys_error

import "errors"

var (
	// CatalogueOrderInvalidErr 表示排序请求自身不合法.
	CatalogueOrderInvalidErr = errors.New("目录排序参数不合法")
	// CatalogueNotFoundErr 表示请求引用的父目录已不存在.
	CatalogueNotFoundErr = errors.New("所选目录不存在,请刷新目录树")
	// CatalogueOrderChangedErr 表示数据库中的本层分组与 expected 不一致.
	CatalogueOrderChangedErr = errors.New("目录状态已变化, 请刷新后确认排序")
	// CatalogueSortRequiredErr 表示普通节点需要先排入阅读顺序才能显示.
	CatalogueSortRequiredErr = errors.New("请先为该节点排序, 再设置为显示")
)
