package admin

// ChangePasswordReq 定义当前管理员的改密请求.
type ChangePasswordReq struct {
	Admin *ChangePasswordAdminReq `json:"admin" binding:"required"` // Admin 包含原密码和新密码.
}

// ChangePasswordAdminReq 保留两项输入的原始字符串.
type ChangePasswordAdminReq struct {
	CurrentPassword string `json:"current_password" binding:"required,notblank"`          // CurrentPassword 用原值核验, 兼容历史长度.
	NewPassword     string `json:"new_password" binding:"required,min=6,max=20,notblank"` // NewPassword 按 Unicode 码点限制长度.
}
