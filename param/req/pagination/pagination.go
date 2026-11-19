package pagination

// Pagination 本结构体用于定义列表请求中的分页器
type Pagination struct {
	CurrentPage *int `json:"currentPage" binding:"required,min=1"` // CurrentPage 当前页码
	ItemPerPage *int `json:"itemPerPage" binding:"required,min=1"` // ItemPerPage 每页显示条数
}
