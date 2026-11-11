package admin

// LoginReq 本结构体用于定义请求登录接口时的请求体
type LoginReq struct {
	Admin *LoginAdminReq `json:"admin" binding:"required"`
}

// LoginAdminReq 本结构体用于定义请求登录接口时的管理员信息
type LoginAdminReq struct {
	Username string `json:"username" binding:"required,min=5,max=20,notblank"` // Username 管理员用户名,要求长度在5-20之间
	Password string `json:"password" binding:"required,min=6,max=20,notblank"` // Password 管理员密码,要求长度在6-20之间
}
