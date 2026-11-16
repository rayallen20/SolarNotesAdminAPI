package planet

// ListResp 本结构体用于定义天体列表成功响应中 data 内的数据结构
type ListResp struct {
	Planets []*ListPlanetResp `json:"planets"` // Planets 已创建的天体列表(没有记录时返回空数组)
}

// ListPlanetResp 本结构体用于定义天体列表中的单条天体信息
type ListPlanetResp struct {
	ID    string `json:"id"`    // ID 雪花 ID 的十进制字符串(避免 JavaScript 整数精度丢失)
	Name  string `json:"name"`  // Name 天体标识(用于匹配前端资源和中文名称)
	Title string `json:"title"` // Title 简介标题
	Intro string `json:"intro"` // Intro 简介内容
}
