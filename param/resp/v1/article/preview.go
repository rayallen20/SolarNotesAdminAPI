package article

// PreviewResp 是 data.article 的内容.
type PreviewResp struct {
	ID      string `json:"id"`      // ID 是文章主键的十进制字符串.
	Title   string `json:"title"`   // Title 是目录树中显示的文件名.
	Content string `json:"content"` // Content 是已改写图片地址的 UTF-8 Markdown.
}
