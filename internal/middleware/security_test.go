package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.Use(SecurityHeaders())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	t.Run("基础安全头设置", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()
		
		router.ServeHTTP(recorder, req)
		
		expectedHeaders := map[string]string{
			"X-XSS-Protection":       "1; mode=block",
			"X-Content-Type-Options": "nosniff",
			"X-Frame-Options":        "DENY",
			"Referrer-Policy":        "strict-origin-when-cross-origin",
			"Server":                 "Palu-Wiki",
		}
		
		for header, expected := range expectedHeaders {
			assert.Equal(t, expected, recorder.Header().Get(header), "Security header %s mismatch", header)
		}
	})

	t.Run("内容安全策略CSP设置", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()
		
		router.ServeHTTP(recorder, req)
		
		csp := recorder.Header().Get("Content-Security-Policy")
		
		// 检查CSP包含重要指令
		expectedDirectives := []string{
			"default-src 'self'",
			"script-src 'self'",
			"style-src 'self'",
			"img-src 'self'",
			"object-src 'none'",
			"frame-ancestors 'none'",
		}
		
		for _, directive := range expectedDirectives {
			assert.Contains(t, csp, directive, "CSP should contain directive: %s", directive)
		}
	})

	t.Run("权限策略设置", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()
		
		router.ServeHTTP(recorder, req)
		
		permissions := recorder.Header().Get("Permissions-Policy")
		
		expectedPolicies := []string{
			"geolocation=()",
			"microphone=()",
			"camera=()",
		}
		
		for _, policy := range expectedPolicies {
			assert.Contains(t, permissions, policy, "Permissions-Policy should contain: %s", policy)
		}
	})

	t.Run("HTTPS环境下的HSTS头", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-Proto", "https") // 模拟HTTPS环境
		recorder := httptest.NewRecorder()
		
		router.ServeHTTP(recorder, req)
		
		hsts := recorder.Header().Get("Strict-Transport-Security")
		assert.Contains(t, hsts, "max-age=31536000", "HSTS should be set for HTTPS")
		assert.Contains(t, hsts, "includeSubDomains", "HSTS should include subdomains")
	})
}

func TestSecureCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.Use(SecureCORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	t.Run("允许的源设置正确的CORS头", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3001")
		recorder := httptest.NewRecorder()
		
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, "http://localhost:3001", recorder.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", recorder.Header().Get("Access-Control-Allow-Credentials"))
		assert.Contains(t, recorder.Header().Get("Access-Control-Allow-Methods"), "GET")
		assert.Contains(t, recorder.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	})

	t.Run("不允许的源不设置CORS头", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://malicious.example.com")
		recorder := httptest.NewRecorder()
		
		router.ServeHTTP(recorder, req)
		
		// 不应该设置允许的源
		assert.NotEqual(t, "https://malicious.example.com", recorder.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("OPTIONS预检请求处理", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		recorder := httptest.NewRecorder()
		
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusNoContent, recorder.Code)
		assert.Equal(t, "86400", recorder.Header().Get("Access-Control-Max-Age"))
	})

	t.Run("环境变量CORS源配置", func(t *testing.T) {
		// 临时设置环境变量
		originalCorsOrigins := os.Getenv("CORS_ORIGINS")
		os.Setenv("CORS_ORIGINS", "https://prod.example.com,https://app.example.com")
		defer func() {
			if originalCorsOrigins == "" {
				os.Unsetenv("CORS_ORIGINS")
			} else {
				os.Setenv("CORS_ORIGINS", originalCorsOrigins)
			}
		}()
		
		// 创建新的路由器来测试环境变量配置
		testRouter := gin.New()
		testRouter.Use(SecureCORS())
		testRouter.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://prod.example.com")
		recorder := httptest.NewRecorder()
		
		testRouter.ServeHTTP(recorder, req)
		
		assert.Equal(t, "https://prod.example.com", recorder.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestInputValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.Use(InputValidation())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	t.Run("正常请求通过验证", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", strings.NewReader("normal data"))
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible)")
		recorder := httptest.NewRecorder()
		
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("缺少User-Agent被拒绝", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", strings.NewReader("data"))
		// 不设置User-Agent
		recorder := httptest.NewRecorder()
		
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "缺少User-Agent")
		assert.Contains(t, recorder.Body.String(), "MISSING_USER_AGENT")
	})

	t.Run("路径遍历攻击被阻止", func(t *testing.T) {
		maliciousPaths := []string{
			"/test/../../../etc/passwd",
			"/test/~/secret",
			"/test/..%2F..%2Fetc%2Fpasswd",
		}
		
		for _, path := range maliciousPaths {
			req := httptest.NewRequest("GET", path, nil)
			req.Header.Set("User-Agent", "Mozilla/5.0")
			recorder := httptest.NewRecorder()
			
			router.ServeHTTP(recorder, req)
			
			assert.Equal(t, http.StatusBadRequest, recorder.Code, "Path %s should be blocked", path)
			assert.Contains(t, recorder.Body.String(), "无效的路径")
		}
	})

	t.Run("OPTIONS请求允许空User-Agent", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		// OPTIONS请求不需要User-Agent
		recorder := httptest.NewRecorder()
		
		router.ServeHTTP(recorder, req)
		
		// OPTIONS请求应该通过验证（状态码取决于后续处理）
		assert.NotEqual(t, http.StatusBadRequest, recorder.Code)
	})
}

func TestAntiBot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.Use(AntiBot())
	router.GET("/test", func(c *gin.Context) {
		isBot, _ := c.Get("is_bot")
		c.JSON(200, gin.H{"is_bot": isBot})
	})

	t.Run("正常用户不被标记为机器人", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		recorder := httptest.NewRecorder()
		
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Body.String(), `"is_bot":null`) // 或者 false，取决于实现
	})

	t.Run("检测常见爬虫User-Agent", func(t *testing.T) {
		botUserAgents := []string{
			"Googlebot/2.1",
			"Mozilla/5.0 (compatible; bingbot/2.0)",
			"python-requests/2.25.1",
			"wget/1.20.3",
			"curl/7.68.0",
			"Apache-HttpClient/4.5.1",
		}
		
		for _, userAgent := range botUserAgents {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("User-Agent", userAgent)
			recorder := httptest.NewRecorder()
			
			router.ServeHTTP(recorder, req)
			
			assert.Equal(t, http.StatusOK, recorder.Code)
			assert.Contains(t, recorder.Body.String(), `"is_bot":true`, "User-Agent %s should be detected as bot", userAgent)
		}
	})

	t.Run("大小写不敏感检测", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "MyApp-CRAWLER/1.0")
		recorder := httptest.NewRecorder()
		
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Body.String(), `"is_bot":true`)
	})
}