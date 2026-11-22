package main

import (
	"SolarNotesAdminAPI/conf"
	"SolarNotesAdminAPI/controller/v1/admin"
	"SolarNotesAdminAPI/controller/v1/article"
	"SolarNotesAdminAPI/controller/v1/book"
	"SolarNotesAdminAPI/controller/v1/catalogue"
	"SolarNotesAdminAPI/controller/v1/planet"
	"SolarNotesAdminAPI/db"
	"SolarNotesAdminAPI/middleware"
	"SolarNotesAdminAPI/snowflake"
	"SolarNotesAdminAPI/validation"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// main 本函数用于串联启动流程,并处理启动和运行错误
func main() {
	configInstance, err := initApp()
	if err != nil {
		log.Fatal(err)
	}

	router, err := newRouter(configInstance.Service)
	if err != nil {
		log.Fatal(err)
	}

	server := newHTTPServer(configInstance.Service, router)
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

// initApp 本函数用于加载配置,并按顺序初始化:
// - 校验器
// - 数据库
// - 雪花 ID 生成器
func initApp() (*conf.Config, error) {
	err := conf.Load()
	if err != nil {
		return nil, err
	}
	configInstance := conf.Get()

	err = validation.Register()
	if err != nil {
		return nil, err
	}

	err = db.Connect(configInstance.Database)
	if err != nil {
		return nil, err
	}

	err = snowflake.InitSnowflake(configInstance.Snowflake.NodeID)
	if err != nil {
		return nil, err
	}

	return configInstance, nil
}

// newRouter 本函数用于配置可信代理,并按既定顺序注册中间件和业务路由
// 中间件顺序:
// 1. 设置跨域头
// 2. 登录限流
// 3. 限制请求体大小
// 4. 处理预检请求
// 5. 鉴权
func newRouter(serviceConf conf.Service) (*gin.Engine, error) {
	router := gin.Default()
	// Tips: 配置可信任代理 IP 或 CIDR 列表,以便在使用 X-Forwarded-For 等头部时正确获取客户端真实 IP 地址
	// 否则获取到的客户端 IP 地址为前置代理的地址,从而导致登录接口访问频率限制失效
	err := router.SetTrustedProxies(serviceConf.TrustedProxies)
	if err != nil {
		return nil, err
	}

	// 为登录路由提供唯一的路径来源,避免限流匹配和注册路径不一致
	v1Prefix := "/v1"
	loginPath := "/admin/login"
	loginFullPath := v1Prefix + loginPath
	loginLimiter := middleware.NewLoginRouteLimit(loginFullPath)

	// 声明免鉴权规则
	// key 为 HTTP 方法,value 为该方法下允许免鉴权的完整路由模板列表
	exemptRoutes := map[string][]string{
		http.MethodPost: {loginFullPath},
	}
	adminAuth, err := middleware.NewAdminAuth(exemptRoutes)
	if err != nil {
		return nil, err
	}

	bodyLimit, err := middleware.NewRequestBodyLimit(serviceConf.MaxRequestBodyBytes)
	if err != nil {
		return nil, err
	}

	router.Use(
		// 设置跨域头
		middleware.CorsHeaders,
		// 设置登录限流
		loginLimiter.Handle,
		// 限制请求体大小
		bodyLimit.Handle,
		// 处理预检请求
		middleware.CorsPreflight,
		// 鉴权
		adminAuth.Handle,
	)

	v1 := router.Group(v1Prefix)
	{
		v1.POST(loginPath, admin.Login)
		v1.GET("/planet/status", planet.ShowStatus)
		v1.POST("/planet/create", planet.Create)
		v1.GET("/planet/list", planet.List)
		v1.POST("/book/create", book.Create)
		v1.POST("/book/list", book.List)
		v1.POST("/book/options", book.Options)
		v1.POST("/catalogue/tree", catalogue.Tree)
		v1.POST("/catalogue/reorder", catalogue.Reorder)
		v1.POST("/catalogue/display", catalogue.Display)
		v1.POST("/article/preview", article.Preview)
	}

	return router, nil
}

// newHTTPServer 本函数用于根据服务配置构建 HTTP 服务,由调用方启动监听
func newHTTPServer(serviceConf conf.Service, handler http.Handler) *http.Server {
	return &http.Server{
		// 开发环境沿用当前监听地址
		Addr: "0.0.0.0:" + serviceConf.Port,
		// TODO: 使用同机 Nginx 部署时,将上一行改为:
		// Addr: "127.0.0.1:" + serviceConf.Port,
		Handler:           handler,
		ReadHeaderTimeout: serviceConf.ReadHeaderTimeout(), // 读取完整请求头的最长时间
		ReadTimeout:       serviceConf.ReadTimeout(),       // 读取整个请求(包括请求头和请求体)的最长时间
		IdleTimeout:       serviceConf.IdleTimeout(),       // 已完成请求的连接,在等待下一个请求时的最长空闲时间
	}
}
