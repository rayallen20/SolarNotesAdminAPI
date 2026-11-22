package book

// OptionResp 定义书籍下拉选项响应体结构
type OptionResp struct {
	ID            string `json:"id"`             // ID 书籍 ID
	Title         string `json:"title"`          // Title 书籍标题
	Sort          uint32 `json:"sort"`           // Sort 书籍排序
	DisplayStatus string `json:"display_status"` // DisplayStatus 书籍显隐状态
}
