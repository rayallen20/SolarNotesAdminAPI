package admin

// LoginResp 本结构体用于定义登录接口的响应体
type LoginResp struct {
	Admin *LoginAdminResp `json:"admin"` // Admin 登录成功后返回的管理员信息
}

type LoginAdminResp struct {
	ID          string `json:"id"`           // ID 管理员ID
	Username    string `json:"username"`     // Username 管理员用户名
	AccessToken string `json:"access_token"` // AccessToken 登录成功后返回的访问令牌
}
