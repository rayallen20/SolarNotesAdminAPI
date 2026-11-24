package planet

// ShowStatusResp 定义天体创建状态接口的成功数据.
type ShowStatusResp struct {
	Planets   []*ShowStatusPlanetResp `json:"planets"`   // Planets 供创建表单展示逐项状态.
	Readiness ReadinessResp           `json:"readiness"` // Readiness 供检测卡展示完整性结论.
}

// ShowStatusPlanetResp 定义一个必需天体的状态.
type ShowStatusPlanetResp struct {
	Name  string `json:"name"`  // Name 是前后端约定的天体标识.
	Exist bool   `json:"exist"` // Exist 表示库中存在同名天体.
}

// ReadinessResp 定义本次检测的摘要.
type ReadinessResp struct {
	Ready         bool     `json:"ready"`          // Ready 表示全部必需天体均已创建.
	RequiredCount int      `json:"required_count"` // RequiredCount 是必需天体总数.
	CreatedCount  int      `json:"created_count"`  // CreatedCount 是其中已创建的数量.
	MissingNames  []string `json:"missing_names"`  // MissingNames 保存缺失标识, 齐全时为空数组.
}
