package pagination

// Pagination 本结构体用于定义分页器响应体结构
type Pagination struct {
	CurrentPage int   `json:"currentPage"` // CurrentPage 本次请求的页码
	ItemPerPage int   `json:"itemPerPage"` // ItemPerPage 每页请求条数
	TotalPage   int64 `json:"totalPage"`   // TotalPage 是总页数
	TotalItem   int64 `json:"totalItem"`   // TotalItem 数据总条数
}
