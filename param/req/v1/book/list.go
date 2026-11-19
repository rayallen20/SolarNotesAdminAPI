package book

import "SolarNotesAdminAPI/param/req/pagination"

// ListReq 本机构提用于定义书籍列表请求体
type ListReq struct {
	Pagination *pagination.Pagination `json:"pagination" binding:"required"` // Pagination 分页器
}
