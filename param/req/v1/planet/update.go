package planet

import (
	"encoding/json"

	"SolarNotesAdminAPI/jstext"
)

// UpdateReq 定义天体编辑请求.
type UpdateReq struct {
	Planet *UpdatePlanetReq `json:"planet" binding:"required"` // Planet 包含记录 ID, 原文和目标值.
}

// UpdatePlanetReq 用 ID 定位天体, 同时提交两个可编辑字段的原值和目标值.
type UpdatePlanetReq struct {
	ID       int64                 `json:"id,string" binding:"gt=0"`    // ID 接收十进制字符串.
	Expected *ExpectedPlanetFields `json:"expected" binding:"required"` // Expected 保存抽屉打开时的原文.
	Target   *TargetPlanetFields   `json:"target" binding:"required"`   // Target 保存本次提交的完整内容.
}

// ExpectedPlanetFields 使用指针区分历史空串和缺失字段.
type ExpectedPlanetFields struct {
	Title *string `json:"title" binding:"required"` // Title 保留原始标题及首尾空白.
	Intro *string `json:"intro" binding:"required"` // Intro 保留原始简介及首尾空白.
}

// TargetPlanetFields 在 JSON 绑定时修剪文本, 然后校验长度.
type TargetPlanetFields struct {
	Title EditText `json:"title" binding:"required,max=20"`  // Title 接受 1 至 20 个码点.
	Intro EditText `json:"intro" binding:"required,max=150"` // Intro 接受 1 至 150 个码点.
}

// EditText 表示天体编辑请求中已修剪的目标文本.
type EditText string

// UnmarshalJSON 使用与页面 String.prototype.trim 相同的首尾空白规则.
func (t *EditText) UnmarshalJSON(data []byte) error {
	var value string
	err := json.Unmarshal(data, &value)
	if err != nil {
		return err
	}
	*t = EditText(jstext.TrimSpace(value))
	return nil
}
