package article

// PreviewReq 指定管理员在目录中点击的文件节点.
type PreviewReq struct {
	Catalogue *PreviewCatalogueReq `json:"catalogue" binding:"required"` // Catalogue 为必填节点引用.
}

// PreviewCatalogueReq 接收十进制字符串形式的雪花 ID.
type PreviewCatalogueReq struct {
	ID int64 `json:"id,string" binding:"gt=0"` // ID 来自目录树中的文件节点.
}
