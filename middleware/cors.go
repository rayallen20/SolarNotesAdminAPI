package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CorsHeaders 本函数用于在处理请求前写入跨域响应头
func CorsHeaders(c *gin.Context) {
	origin := c.GetHeader("Origin")
	if origin != "" {
		// 开发期不拒绝任何源,生产环境发版之前再收紧
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Vary", "Origin")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", c.GetHeader("Access-Control-Request-Headers"))
		c.Header("Access-Control-Max-Age", "600")
	}
	c.Next()
}

// CorsPreflight 本函数用于在请求体大小校验通过后应答 OPTIONS 预检请求
func CorsPreflight(c *gin.Context) {
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	c.Next()
}
