package catalogue

// DisplayReq 指定单个目录节点的目标显示状态.
type DisplayReq struct {
	Catalogue *DisplayCatalogueReq `json:"catalogue" binding:"required"` // Catalogue 是必填的目标节点.
}

// DisplayCatalogueReq 保存目录主键及要设置的状态.
type DisplayCatalogueReq struct {
	ID            int64  `json:"id,string" binding:"gt=0"`                               // ID 是目录树中节点的十进制字符串 ID.
	DisplayStatus string `json:"display_status" binding:"required,oneof=visible hidden"` // DisplayStatus 是本次请求的目标值.
}
