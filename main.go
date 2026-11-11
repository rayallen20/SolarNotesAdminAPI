package main

import (
	"SolarNotesAdminAPI/conf"
	"SolarNotesAdminAPI/controller/v1/admin"
	"SolarNotesAdminAPI/db"
	"SolarNotesAdminAPI/middleware"
	"SolarNotesAdminAPI/snowflake"
	"SolarNotesAdminAPI/validation"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	err := conf.Load()
	if err != nil {
		log.Fatal(err)
	}
	configInstance := conf.Get()

	err = validation.Register()
	if err != nil {
		log.Fatal(err)
	}

	err = db.Connect(configInstance.Database)
	if err != nil {
		log.Fatal(err)
	}

	err = snowflake.InitSnowflake(configInstance.Snowflake.NodeID)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	// Tips: 配置可信任代理 IP 或 CIDR 列表,以便在使用 X-Forwarded-For 等头部时正确获取客户端真实 IP 地址
	// 否则获取到的客户端 IP 地址为前置代理的地址,从而导致登录接口访问频率限制失效
	err = router.SetTrustedProxies(configInstance.Service.TrustedProxies)
	if err != nil {
		log.Fatal(err)
	}

	// 为登录路由提供唯一的路径来源,避免限流匹配和注册路径不一致
	v1Prefix := "/v1"
	loginPath := "/admin/login"
	loginLimiter := middleware.NewLoginRouteLimit(v1Prefix + loginPath)

	bodyLimit, err := middleware.NewRequestBodyLimit(configInstance.Service.MaxRequestBodyBytes)
	if err != nil {
		log.Fatal(err)
	}

	// 先设置跨域头和登录限流,再限制所有请求体,最后处理预检
	router.Use(
		middleware.CorsHeaders,
		loginLimiter.Handle,
		bodyLimit.Handle,
		middleware.CorsPreflight,
	)

	v1 := router.Group(v1Prefix)
	{
		v1.POST(loginPath, admin.Login)
	}

	server := &http.Server{
		// 开发环境沿用当前监听地址。
		Addr: "0.0.0.0:" + configInstance.Service.Port,
		// TODO: 使用同机 Nginx 部署时,将上一行改为:
		// Addr: "127.0.0.1:" + configInstance.Service.Port,
		Handler:           router,
		ReadHeaderTimeout: configInstance.Service.ReadHeaderTimeout(), // 读取完整请求头的最长时间
		ReadTimeout:       configInstance.Service.ReadTimeout(),       // 读取整个请求(包括请求头和请求体)的最长时间
		IdleTimeout:       configInstance.Service.IdleTimeout(),       // 已完成请求的连接,在等待下一个请求时的最长空闲时间
	}

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
