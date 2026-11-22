package book

// OptionsReq 定义按天体读取书籍选项的请求
type OptionsReq struct {
	Planet *OptionsPlanetReq `json:"planet" binding:"required"` // Planet 当前选中的天体
}

// OptionsPlanetReq 定义书籍选项查询需要的天体标识
type OptionsPlanetReq struct {
	ID int64 `json:"id,string" binding:"gt=0"` // ID 天体 ID
}
