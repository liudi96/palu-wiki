package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	t.Run("正常GET请求 - CORS头设置", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()
		
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "*", recorder.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", recorder.Header().Get("Access-Control-Allow-Credentials"))
		assert.Contains(t, recorder.Header().Get("Access-Control-Allow-Headers"), "Authorization")
		assert.Contains(t, recorder.Header().Get("Access-Control-Allow-Methods"), "POST")
	})

	t.Run("OPTIONS预检请求处理", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		recorder := httptest.NewRecorder()
		
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusNoContent, recorder.Code)
		assert.Equal(t, "*", recorder.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("CORS头完整性检查", func(t *testing.T) {
		expectedHeaders := map[string]string{
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
			"Access-Control-Allow-Methods":     "POST, OPTIONS, GET, PUT, DELETE",
		}
		
		req := httptest.NewRequest("GET", "/test", nil)
		recorder := httptest.NewRecorder()
		
		router.ServeHTTP(recorder, req)
		
		for header, expected := range expectedHeaders {
			assert.Equal(t, expected, recorder.Header().Get(header), "CORS header %s mismatch", header)
		}
		
		// 检查允许的请求头包含重要字段
		allowedHeaders := recorder.Header().Get("Access-Control-Allow-Headers")
		requiredHeaders := []string{"Content-Type", "Authorization", "X-Requested-With"}
		
		for _, required := range requiredHeaders {
			assert.Contains(t, allowedHeaders, required, "Missing required header: %s", required)
		}
	})
}