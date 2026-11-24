package book

import (
	"SolarNotesAdminAPI/jstext"
	"encoding/json"
)

// CreateReq 本结构体用于定义创建书籍请求的外层对象
type CreateReq struct {
	Book *CreateBookReq `json:"book" binding:"required"` // Book 待创建书籍
}

// CreateBookReq 本结构体用于定义表单提交的书籍字段
type CreateBookReq struct {
	Title         Title            `json:"title" binding:"required,max=20"`                        // Title 在绑定时去除首尾空白
	DisplayStatus string           `json:"display_status" binding:"required,oneof=visible hidden"` // DisplayStatus 显示状态
	Planet        *CreatePlanetReq `json:"planet" binding:"required"`                              // Planet 仅引用已创建天体的 ID
}

// CreatePlanetReq 本结构体用于定义书籍所属天体
type CreatePlanetReq struct {
	ID int64 `json:"id,string" binding:"gt=0"` // ID 天体ID, 仅允许已创建天体的 ID
}

// Title 是已去除首尾空白的书名, max 校验针对处理后的值
type Title string

// UnmarshalJSON 本方法用于在 Gin 执行 binding 校验前, 按页面 trim 规则处理书名
func (t *Title) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*t = Title(jstext.TrimSpace(value))
	return nil
}
