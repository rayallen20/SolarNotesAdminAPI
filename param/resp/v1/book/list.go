package book

import (
	respPagination "SolarNotesAdminAPI/param/resp/pagination"
	respPlanet "SolarNotesAdminAPI/param/resp/v1/planet"
)

// ListResp 本结构体用于定义书籍列表接口响应体
type ListResp struct {
	Books      []*ListBookResp            `json:"books"`      // Books 当前页的书籍记录(没有记录时返回空数组)
	Pagination *respPagination.Pagination `json:"pagination"` // Pagination 分页器
}

// ListBookResp 本结构体用于定义书籍列表中的单条记录
type ListBookResp struct {
	ID            string                     `json:"id"`             // ID 书籍 ID 的十进制字符串(避免前端整数精度丢失)
	Title         string                     `json:"title"`          // Title 书籍标题
	Sort          uint32                     `json:"sort"`           // Sort 书籍在所属天体下的排序值
	DisplayStatus string                     `json:"display_status"` // DisplayStatus 书籍的显示状态, 为 visible 或 hidden
	Planet        *respPlanet.ListPlanetResp `json:"planet"`         // Planet 所属天体, 包含字符串 ID, 名称, 标题和简介
}
