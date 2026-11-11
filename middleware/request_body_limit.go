package middleware

import (
	"SolarNotesAdminAPI/resp_body"
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequestBodyLimit 本结构体用于对进入后端 API 的请求体执行统一的字节数限制
type RequestBodyLimit struct {
	maxBytes int64 // maxBytes 本字段用于表示允许的最大请求体字节数
}

// NewRequestBodyLimit 创建请求体限制器；上限必须为正数。
func NewRequestBodyLimit(maxBytes int64) (*RequestBodyLimit, error) {
	if maxBytes <= 0 {
		return nil, errors.New("请求体大小上限必须大于 0")
	}
	return &RequestBodyLimit{maxBytes: maxBytes}, nil
}

// Handle 本方法用于在进入后续处理前获取请求体大小,并恢复供后续处理的请求体
func (l *RequestBodyLimit) Handle(c *gin.Context) {
	// 已知长度超过上限时,不必读取请求体
	if c.Request.ContentLength > l.maxBytes {
		l.reject(c, "请求体超过大小限制")
		return
	}

	limitedBody := http.MaxBytesReader(c.Writer, c.Request.Body, l.maxBytes)
	defer func() {
		err := limitedBody.Close()
		if err != nil {
			log.Printf("关闭请求体失败: error_type=%T", err)
		}
	}()

	body, err := io.ReadAll(limitedBody)
	if err != nil {
		var sizeErr *http.MaxBytesError
		if errors.As(err, &sizeErr) {
			l.reject(c, "请求体超过大小限制")
		} else {
			l.reject(c, "读取请求体失败")
		}
		return
	}

	// 复用读取结果,下游仍可正常读取原始字节
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	c.Next()
}

// reject 本方法用于以项目既有的参数错误格式终止请求
func (l *RequestBodyLimit) reject(c *gin.Context, message string) {
	body := &resp_body.RespBody{}
	body.ValidateError(errors.New(message))
	c.Header("Cache-Control", "no-store")
	c.AbortWithStatusJSON(http.StatusOK, body)
}
