package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// LoginRouteLimit 本结构体用于对指定登录路由执行已有的 IP 限流
type LoginRouteLimit struct {
	fullPath string            // fullPath 给定路由的完整路径
	limiter  *LoginRateLimiter // limiter 给定路由使用的限流器
}

// NewLoginRouteLimit 本函数用于为给定完整路由创建限流包装器
func NewLoginRouteLimit(fullPath string) *LoginRouteLimit {
	return &LoginRouteLimit{
		fullPath: fullPath,
		limiter:  NewLoginRateLimiter(),
	}
}

// Handle 本方法用于仅对匹配的 POST 请求调用限流中间件
func (l *LoginRouteLimit) Handle(c *gin.Context) {
	if c.Request.Method == http.MethodPost && c.FullPath() == l.fullPath {
		l.limiter.Handle(c)
		return
	}
	c.Next()
}
