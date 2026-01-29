package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"oj-system/internal/model"
)

// 简单的内存限流器
type rateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
	limit    int           // 时间窗口内最大请求数
	window   time.Duration // 时间窗口
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (r *rateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-r.window)

	// 清理过期记录
	var valid []time.Time
	for _, t := range r.requests[key] {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}
	r.requests[key] = valid

	// 检查是否超限
	if len(r.requests[key]) >= r.limit {
		return false
	}

	// 记录本次请求
	r.requests[key] = append(r.requests[key], now)
	return true
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := newRateLimiter(limit, window)
	return func(c *gin.Context) {
		key := c.ClientIP()
		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, model.Error(429, "请求过于频繁，请稍后再试"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// SubmitRateLimitMiddleware 提交限流（更严格）
func SubmitRateLimitMiddleware() gin.HandlerFunc {
	limiter := newRateLimiter(3, 10*time.Second) // 每 10 秒最多 3 次提交
	return func(c *gin.Context) {
		key := c.ClientIP()
		if userID := GetUserID(c); userID > 0 {
			key = fmt.Sprintf("user:%d", userID)
		} else {
			key = "ip:" + c.ClientIP()
		}
		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, model.Error(429, "请求过于频繁，请稍后再试"))
			c.Abort()
			return
		}
		c.Next()
	}
}
