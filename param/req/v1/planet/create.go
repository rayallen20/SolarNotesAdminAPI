package planet

// CreateReq 本结构体用于定义请求创建天体接口时的请求体
type CreateReq struct {
	Planet *CreatePlanetReq `json:"planet" binding:"required"`
}

// CreatePlanetReq 本结构体用于定义请求创建天体接口时的天体信息
type CreatePlanetReq struct {
	Name  string `json:"name" binding:"required,oneof=Sun Mercury Venus Earth Mars Jupiter Saturn Uranus Neptune"` // Name 天体标识(必须精确匹配预定义名称)
	Title string `json:"title" binding:"required,notblank,max=20"`                                                 // Title 简介标题
	Intro string `json:"intro" binding:"required,notblank,max=150"`                                                // Intro 简介内容
}
