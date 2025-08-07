package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 简单的内存限流器
type RateLimiter struct {
	requests map[string]*ClientInfo
	mutex    sync.RWMutex
	maxReq   int
	window   time.Duration
}

type ClientInfo struct {
	requests  []time.Time
	lastReset time.Time
}

func NewRateLimiter(maxReq int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*ClientInfo),
		maxReq:   maxReq,
		window:   window,
	}

	// 启动清理协程
	go rl.cleanup()
	
	return rl
}

// 限流中间件
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		if !rl.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	
	client, exists := rl.requests[clientIP]
	if !exists {
		client = &ClientInfo{
			requests:  []time.Time{now},
			lastReset: now,
		}
		rl.requests[clientIP] = client
		return true
	}
	
	// 清理过期请求
	validRequests := []time.Time{}
	for _, req := range client.requests {
		if now.Sub(req) < rl.window {
			validRequests = append(validRequests, req)
		}
	}
	
	client.requests = validRequests
	
	// 检查是否超过限制
	if len(client.requests) >= rl.maxReq {
		return false
	}
	
	// 添加新请求
	client.requests = append(client.requests, now)
	return true
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		
		for ip, client := range rl.requests {
			// 如果客户端在窗口时间内没有请求，删除记录
			if now.Sub(client.lastReset) > rl.window*2 {
				delete(rl.requests, ip)
			}
		}
		
		rl.mutex.Unlock()
	}
}

// 创建默认限流器
func DefaultRateLimit() gin.HandlerFunc {
	// 从环境变量获取限流配置
	maxReqStr := os.Getenv("RATE_LIMIT_RPM")
	maxReq := 100 // 默认每分钟100个请求
	
	if maxReqStr != "" {
		if parsed, err := strconv.Atoi(maxReqStr); err == nil {
			maxReq = parsed
		}
	}
	
	limiter := NewRateLimiter(maxReq, time.Minute)
	return limiter.Middleware()
}

// 严格限流器（用于登录等敏感操作）
func StrictRateLimit() gin.HandlerFunc {
	limiter := NewRateLimiter(5, time.Minute) // 每分钟5次
	return limiter.Middleware()
}

// API限流器
func APIRateLimit() gin.HandlerFunc {
	limiter := NewRateLimiter(60, time.Minute) // 每分钟60次
	return limiter.Middleware()
}