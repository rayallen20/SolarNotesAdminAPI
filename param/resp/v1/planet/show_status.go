package planet

// ShowStatusResp 本结构体用于定义显示行星状态接口的响应体
type ShowStatusResp struct {
	Planets []*ShowStatusPlanetResp `json:"planets"` // Planets 行星列表
}

// ShowStatusPlanetResp 本结构体用于定义显示行星状态接口中每个行星的响应体
type ShowStatusPlanetResp struct {
	Name  string `json:"name"`  // Name 行星名称 (用于前台前端匹配3D资源)
	Exist bool   `json:"exist"` // Exist 行星是否存在 (用于判定能否创建该天体)
}
