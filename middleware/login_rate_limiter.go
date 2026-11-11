package middleware

import (
	"SolarNotesAdminAPI/resp_body"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	loginWindow     = time.Minute // loginWindow 本常量用于定义记录请求登录接口次数的时间窗口
	loginMaxRequest = 5           // loginMaxRequest 本常量用于定义在 loginWindow 时间窗口内允许的最大请求次数
	loginMaxClients = 1000        // loginMaxClients 本常量用于定义在 loginWindow 时间窗口内同时跟踪的客户端数量
)

// LoginRateLimiter 本结构体用于在单个进程内按客户端 IP 地址限制请求登录接口的次数,以防止暴力破解攻击
type LoginRateLimiter struct {
	mu          sync.Mutex             // mu 用于保护 requests 和 lastCleanup 字段的并发访问,允许多个请求并发访问
	requests    map[string][]time.Time // requests 本字段用于为每个客户端保存当前窗口内已放行请求的时间戳列表,用于计算请求次数
	lastCleanup time.Time              // lastCleanup 本字段用于记录上一次移除过期客户端记录的时间
	nextExpiry  time.Time              // nextExpiry 最早可能出现过期记录的时间(可能早于实际到期时间)
}

// NewLoginRateLimiter 本函数用于创建一个新的 LoginRateLimiter 实例,并初始化其 requests 字段
func NewLoginRateLimiter() *LoginRateLimiter {
	return &LoginRateLimiter{
		requests: make(map[string][]time.Time),
	}
}

// Handle 本方法用于在请求登录接口时检查客户端 IP 地址的请求频率
// 若超过限制则返回访问频率限制响应体
// 否则继续处理请求
func (l *LoginRateLimiter) Handle(c *gin.Context) {
	// 控制浏览器/代理服务器,不要缓存本次响应的内容
	c.Header("Cache-Control", "no-store")

	if !l.allow(c.ClientIP()) {
		respBody := &resp_body.RespBody{}
		respBody.LoginRateLimitError()
		c.AbortWithStatusJSON(200, respBody)
		return
	}

	c.Next()
}

func (l *LoginRateLimiter) allow(clientIP string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-loginWindow)

	// 定期清理,避免未满容量时长期保留过期客户端
	if l.lastCleanup.IsZero() || now.Sub(l.lastCleanup) >= loginWindow {
		l.cleanupExpired(now)
	}

	history, exists := l.requests[clientIP]

	// 移除过期的请求记录,只保留窗口内的请求记录
	first := 0
	for first < len(history) && !history[first].After(cutoff) {
		first++
	}
	history = history[first:]

	if len(history) >= loginMaxRequest {
		return false
	}

	// 容量已满且有记录可能到期时,先回收过期客户端
	if !exists && len(l.requests) >= loginMaxClients {
		if l.nextExpiry.IsZero() || !now.Before(l.nextExpiry) {
			l.cleanupExpired(now)
		}
		if len(l.requests) >= loginMaxClients {
			return false
		}
	}

	// 记录当前请求的时间戳
	l.requests[clientIP] = append(history, now)
	newExpiry := now.Add(loginWindow)
	if l.nextExpiry.IsZero() || newExpiry.Before(l.nextExpiry) {
		l.nextExpiry = newExpiry
	}
	return true
}

// cleanupExpired 本方法用于清除窗口内没有请求的客户端,并更新下次可能过期的时间
// 调用前必须持有 l.mu
// 此方法不再加锁,避免重复加锁
func (l *LoginRateLimiter) cleanupExpired(now time.Time) {
	cutoff := now.Add(-loginWindow)
	var nextExpiry time.Time

	for ip, history := range l.requests {
		if len(history) == 0 {
			delete(l.requests, ip)
			continue
		}

		lastRequestTime := history[len(history)-1]
		if !lastRequestTime.After(cutoff) {
			delete(l.requests, ip)
			continue
		}

		expiresAt := lastRequestTime.Add(loginWindow)
		if nextExpiry.IsZero() || expiresAt.Before(nextExpiry) {
			nextExpiry = expiresAt
		}
	}

	l.nextExpiry = nextExpiry
	l.lastCleanup = now
}
