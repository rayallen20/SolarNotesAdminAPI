package catalogue

// TreeReq 定义读取整本书目录树的请求
type TreeReq struct {
	Book *TreeBookReq `json:"book" binding:"required"` // Book 标识当前选择的书籍
}

// TreeBookReq 请求体中的书籍部分
type TreeBookReq struct {
	ID int64 `json:"id,string" binding:"gt=0"` // ID 书籍 ID
}
