package catalogue

// ReorderReq 是保存同级阅读顺序的请求体.
type ReorderReq struct {
	Catalogue *ReorderCatalogueReq `json:"catalogue" binding:"required"` // Catalogue 描述所选父目录和两份分组.
}

// ReorderCatalogueReq 标识一次保存操作.
type ReorderCatalogueReq struct {
	ParentID int64        `json:"parent_id,string" binding:"gt=0"` // ParentID 为父文件夹本身的 ID.
	Expected *OrderGroups `json:"expected" binding:"required"`     // Expected 是打开排序抽屉时的分组.
	Target   *OrderGroups `json:"target" binding:"required"`       // Target 是保存后的分组.
}

// OrderGroups 要求两个数组都存在, 非 nil 的空数组合法.
type OrderGroups struct {
	SortedChildren   []ChildRef `json:"sorted_children" binding:"required,dive"`   // SortedChildren 的位置决定阅读顺序.
	UnsortedChildren []ChildRef `json:"unsorted_children" binding:"required,dive"` // UnsortedChildren 只表达成员集合.
}

// ChildRef 引用父文件夹的一个直接子节点.
type ChildRef struct {
	ID int64 `json:"id,string" binding:"gt=0"` // ID 是以十进制字符串传入的节点 ID.
}
