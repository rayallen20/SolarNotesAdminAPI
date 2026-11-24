package middleware

import (
	"SolarNotesAdminAPI/auth"
	"SolarNotesAdminAPI/biz"
	"SolarNotesAdminAPI/resp_body"
	"SolarNotesAdminAPI/sys_error"
	"crypto/sha256"
	"errors"
	"net/http"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

// AdminAuth 本结构体用于对免鉴权路由集合之外的请求执行鉴权
type AdminAuth struct {
	exemptRoutes map[string][]string // exemptRoutes 免鉴权路由集合.其中 key 为 HTTP 方法,value 为该方法下免鉴权路由集合
}

// NewAdminAuth 本函数用于校验并复制免鉴权规则(nil 或空集合表示所有请求都需要鉴权)
// map 及其中每个路径切片均复制;构造完成后只读取副本
func NewAdminAuth(exemptRoutes map[string][]string) (*AdminAuth, error) {
	routes := make(map[string][]string, len(exemptRoutes))

	for method, paths := range exemptRoutes {
		if method == "" || strings.IndexFunc(method, unicode.IsSpace) >= 0 {
			return nil, errors.New("exempt route method must not be empty or contain whitespace")
		}

		if len(paths) == 0 {
			return nil, errors.New("exempt route method must specify at least one path")
		}

		// 为每个路径列表分配新的底层数组,避免与调用方共享切片数据
		copiedPaths := make([]string, len(paths))
		for index, path := range paths {
			if !strings.HasPrefix(path, "/") {
				return nil, errors.New("exempt route path must start with /")
			}

			if strings.ContainsAny(path, "?#") {
				return nil, errors.New("exempt route path must not contain query (?) or fragment (#) delimiters")
			}

			if strings.IndexFunc(path, unicode.IsSpace) >= 0 {
				return nil, errors.New("exempt route path must not contain whitespace")
			}

			copiedPaths[index] = path
		}

		routes[method] = copiedPaths
	}

	return &AdminAuth{exemptRoutes: routes}, nil
}

// Handle 本方法用于读取 Bearer 令牌,调用业务层鉴权,成功后将管理员 ID 放入本次请求上下文
// (只有方法与完整路由模板同时命中免鉴权集合时才跳过校验)
func (a *AdminAuth) Handle(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	if a == nil {
		respBody := &resp_body.RespBody{}
		err := &sys_error.InternalError{
			Cause: errors.New("nil authentication middleware"),
		}
		respBody.InternalError(err)
		c.AbortWithStatusJSON(http.StatusOK, respBody)
		return
	}

	if a.isExempt(c) {
		c.Next()
		return
	}

	rawToken, err := a.bearerToken(c)
	if err != nil {
		a.reject(c, err)
		return
	}

	administrator := &biz.Admin{}
	err = administrator.Authenticate(c.Request.Context(), rawToken)
	if err != nil {
		a.reject(c, err)
		return
	}

	// 保存已验证的管理员 ID 和本次请求使用的令牌摘要.
	c.Set(adminIDContextKey, administrator.ID)
	c.Set(adminTokenDigestContextKey, sha256.Sum256([]byte(rawToken)))
	c.Next()
}

// isExempt 本方法用于判定当前请求的方法与完整路由模板是否为免鉴权路由
func (a *AdminAuth) isExempt(c *gin.Context) bool {
	paths, exists := a.exemptRoutes[c.Request.Method]
	if !exists {
		return false
	}

	fullPath := c.FullPath()
	if fullPath == "" {
		return false
	}

	for _, path := range paths {
		if path == fullPath {
			return true
		}
	}

	return false
}

// bearerToken 本方法用于从唯一的 Authorization 请求头读取 Bearer 令牌
// (缺失/重复请求头/方案不匹配/包含多段凭据时均返回 InvalidTokenErr)
func (a *AdminAuth) bearerToken(c *gin.Context) (string, error) {
	values := c.Request.Header.Values("Authorization")

	if len(values) != 1 {
		return "", auth.InvalidTokenErr
	}

	parts := strings.Fields(values[0])

	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", auth.InvalidTokenErr
	}

	return parts[1], nil
}

// reject 本方法用于按项目业务码返回鉴权错误,同时停止后续中间件和业务处理
// (AbortWithStatusJSON 不会结束当前 Go 函数,调用方仍须立即 return)
func (a *AdminAuth) reject(c *gin.Context, err error) {
	respBody := &resp_body.RespBody{}

	var databaseError *sys_error.DatabaseError
	switch {
	case errors.Is(err, auth.ExpiredTokenErr):
		respBody.TokenExpiredError()
	case errors.Is(err, auth.InvalidTokenErr):
		respBody.InvalidToken()
	case errors.As(err, &databaseError):
		respBody.DBError(err)
	default:
		respBody.InternalError(&sys_error.InternalError{Cause: err})
	}

	c.AbortWithStatusJSON(http.StatusOK, respBody)
}
