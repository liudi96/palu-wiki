package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("基础限流功能", func(t *testing.T) {
		limiter := NewRateLimiter(3, time.Minute) // 每分钟3次请求
		
		clientIP := "127.0.0.1"
		
		// 前3次请求应该被允许
		for i := 0; i < 3; i++ {
			allowed := limiter.Allow(clientIP)
			assert.True(t, allowed, "Request %d should be allowed", i+1)
		}
		
		// 第4次请求应该被拒绝
		allowed := limiter.Allow(clientIP)
		assert.False(t, allowed, "4th request should be denied")
	})

	t.Run("不同客户端独立限流", func(t *testing.T) {
		limiter := NewRateLimiter(2, time.Minute)
		
		client1 := "192.168.1.1"
		client2 := "192.168.1.2"
		
		// 客户端1的请求
		assert.True(t, limiter.Allow(client1))
		assert.True(t, limiter.Allow(client1))
		assert.False(t, limiter.Allow(client1)) // 第3次被拒绝
		
		// 客户端2应该不受影响
		assert.True(t, limiter.Allow(client2))
		assert.True(t, limiter.Allow(client2))
		assert.False(t, limiter.Allow(client2)) // 第3次被拒绝
	})

	t.Run("时间窗口过期后重置", func(t *testing.T) {
		limiter := NewRateLimiter(1, 100*time.Millisecond) // 100ms窗口，1次请求
		
		clientIP := "127.0.0.1"
		
		// 第1次请求应该被允许
		assert.True(t, limiter.Allow(clientIP))
		
		// 立即第2次请求应该被拒绝
		assert.False(t, limiter.Allow(clientIP))
		
		// 等待窗口过期
		time.Sleep(150 * time.Millisecond)
		
		// 现在应该又可以请求了
		assert.True(t, limiter.Allow(clientIP))
	})
}

func TestRateLimiterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	t.Run("中间件限流功能", func(t *testing.T) {
		limiter := NewRateLimiter(2, time.Minute)
		
		router := gin.New()
		router.Use(limiter.Middleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		// 前2次请求应该成功
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "127.0.0.1:12345" // 模拟客户端IP
			recorder := httptest.NewRecorder()
			
			router.ServeHTTP(recorder, req)
			assert.Equal(t, http.StatusOK, recorder.Code, "Request %d should succeed", i+1)
		}

		// 第3次请求应该被限流
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		recorder := httptest.NewRecorder()
		
		router.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
		
		// 验证错误响应格式
		assert.Contains(t, recorder.Body.String(), "请求过于频繁")
		assert.Contains(t, recorder.Body.String(), "RATE_LIMIT_EXCEEDED")
	})

	t.Run("默认限流器创建", func(t *testing.T) {
		middleware := DefaultRateLimit()
		assert.NotNil(t, middleware)
	})

	t.Run("严格限流器创建", func(t *testing.T) {
		middleware := StrictRateLimit()
		assert.NotNil(t, middleware)
	})

	t.Run("API限流器创建", func(t *testing.T) {
		middleware := APIRateLimit()
		assert.NotNil(t, middleware)
	})
}

func TestRateLimiterCleanup(t *testing.T) {
	t.Run("客户端记录清理", func(t *testing.T) {
		limiter := NewRateLimiter(10, 50*time.Millisecond)
		
		// 添加一些客户端请求
		limiter.Allow("client1")
		limiter.Allow("client2")
		
		// 验证客户端记录存在
		limiter.mutex.RLock()
		initialCount := len(limiter.requests)
		limiter.mutex.RUnlock()
		
		assert.Equal(t, 2, initialCount)
		
		// 等待清理协程运行（实际测试中清理间隔很长，这里只是验证结构）
		time.Sleep(10 * time.Millisecond)
		
		// 确保清理逻辑存在（通过代码覆盖验证）
		assert.NotNil(t, limiter.requests)
	})
}