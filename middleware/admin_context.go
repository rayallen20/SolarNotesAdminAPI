package middleware

import (
	"crypto/sha256"

	"github.com/gin-gonic/gin"
)

// adminIDContextKey 上下文中管理员ID的键名,用于在请求上下文中存储和检索管理员ID
const adminIDContextKey = "solar_notes.auth.admin_id"

// adminTokenDigestContextKey 保存本次已验证的 Bearer 令牌摘要.
const adminTokenDigestContextKey = "solar_notes.auth.token_digest"

// CurrentAdminID 本函数用于读取本次请求通过鉴权的管理员 ID
// (未经过鉴权/值类型不正确/ ID 无效时返回 false,不触发类型断言 panic)
func CurrentAdminID(c *gin.Context) (int64, bool) {
	if c == nil {
		return 0, false
	}

	value, exists := c.Get(adminIDContextKey)
	if !exists {
		return 0, false
	}

	adminID, ok := value.(int64)
	if !ok || adminID <= 0 {
		return 0, false
	}

	return adminID, true
}

// CurrentAdminTokenDigest 读取当前请求的会话摘要, 缺失或类型错误时返回 false.
func CurrentAdminTokenDigest(c *gin.Context) ([sha256.Size]byte, bool) {
	if c == nil {
		return [sha256.Size]byte{}, false
	}
	value, _ := c.Get(adminTokenDigestContextKey)
	digest, ok := value.([sha256.Size]byte)
	return digest, ok
}
