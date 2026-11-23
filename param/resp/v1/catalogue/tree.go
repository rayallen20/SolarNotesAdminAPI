package catalogue

// TreeBookResp 定义响应体中的书籍部分
type TreeBookResp struct {
	ID            string          `json:"id"`             // ID 书籍 ID
	Title         string          `json:"title"`          // Title 书名
	Sort          uint32          `json:"sort"`           // Sort 书籍排序值
	DisplayStatus string          `json:"display_status"` // DisplayStatus 书籍显隐状态
	Catalogues    []*TreeNodeResp `json:"catalogues"`     // Catalogues 书籍目录树, 为空或根目录被过滤时返回 []
}

// TreeNodeResp 书籍部分中的目录节点部分
type TreeNodeResp struct {
	ID            string          `json:"id"`             // ID 目录节点 ID(文件预览也使用此 ID)
	ParentID      string          `json:"parent_id"`      // ParentID 节点的直接父节点 ID, 根节点为 "0"
	Name          string          `json:"name"`           // Name 节点名称
	Type          string          `json:"type"`           // Type 节点类型
	Sort          *uint32         `json:"sort"`           // Sort 节点排序(排序为 nil 时编码为 JSON null)
	DisplayStatus string          `json:"display_status"` // DisplayStatus 节点显隐状态
	Children      []*TreeNodeResp `json:"children"`       // Children 排除 img 文件夹后的直接子节点, 无返回项时为 []
}
