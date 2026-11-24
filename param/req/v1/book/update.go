package book

// UpdateReq 是管理员编辑书籍的请求.
type UpdateReq struct {
	Book *UpdateBookReq `json:"book" binding:"required"` // Book 引用书籍并携带编辑前后的值.
}

// UpdateBookReq 用 ID 定位书籍, 用 Expected 核对编辑时看到的记录.
type UpdateBookReq struct {
	ID       int64               `json:"id,string" binding:"gt=0"`    // ID 是十进制字符串形式的书籍 ID.
	Expected *ExpectedBookFields `json:"expected" binding:"required"` // Expected 是打开抽屉时的原值.
	Target   *TargetBookFields   `json:"target" binding:"required"`   // Target 是要保存的两个字段.
}

// ExpectedBookFields 保留读取时的书名原文.
type ExpectedBookFields struct {
	Title         *string `json:"title" binding:"required"`                               // Title 用指针区分空字符串与缺失.
	DisplayStatus string  `json:"display_status" binding:"required,oneof=visible hidden"` // DisplayStatus 是原显隐状态.
}

// TargetBookFields 采用创建书籍时的书名与显隐规则.
type TargetBookFields struct {
	Title         Title  `json:"title" binding:"required,max=20"`                        // Title 在校验前修整首尾空白.
	DisplayStatus string `json:"display_status" binding:"required,oneof=visible hidden"` // DisplayStatus 是目标显隐状态.
}
