package auth

import (
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

type AdminClaims struct {
	// jwt.RegisteredClaims JWT注册声明,包含iss/sub/aud/iat/exp/jti等标准字段
	// 其中:
	// sub为管理员ID的10进制字符串
	// iss为JWT的签发者
	// iat为签发时间
	// exp为过期时间
	// jti为JWT的唯一标识符(签发时随机生成)
	jwt.RegisteredClaims
}

// Validate 本方法用于在 JWT 库完成标准声明检查后,对自定义字段进行额外的验证,确保管理员ID和签发时间有效,若无效则返回InvalidTokenErr错误
func (c *AdminClaims) Validate() error {
	id, err := strconv.ParseInt(c.Subject, 10, 64)
	if err != nil || id <= 0 || c.ID == "" || c.IssuedAt == nil {
		return InvalidTokenErr
	}

	return nil
}
