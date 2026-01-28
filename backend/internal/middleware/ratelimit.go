package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"oj-system/internal/model"
)

// 简单的内存限流器
type rateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int           // 时间窗口内最大请求数
	window   time.Duration // 时间窗口
}

var limiter = &rateLimiter{
	requests: make(map[string][]time.Time),
	limit:    60,           // 每分钟 60 次
	window:   time.Minute,
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		limiter.mu.Lock()
		now := time.Now()
		windowStart := now.Add(-window)
		
		// 清理过期记录
		var valid []time.Time
		for _, t := range limiter.requests[ip] {
			if t.After(windowStart) {
				valid = append(valid, t)
			}
		}
		limiter.requests[ip] = valid
		
		// 检查是否超限
		if len(limiter.requests[ip]) >= limit {
			limiter.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, model.Error(429, "请求过于频繁，请稍后再试"))
			c.Abort()
			return
		}
		
		// 记录本次请求
		limiter.requests[ip] = append(limiter.requests[ip], now)
		limiter.mu.Unlock()
		
		c.Next()
	}
}

// SubmitRateLimitMiddleware 提交限流（更严格）
func SubmitRateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddleware(10, time.Minute) // 每分钟最多 10 次提交
}
