package middleware

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("默认日志格式", func(t *testing.T) {
		// 确保没有设置JSON格式
		originalLogFormat := os.Getenv("LOG_FORMAT")
		os.Unsetenv("LOG_FORMAT")
		defer func() {
			if originalLogFormat != "" {
				os.Setenv("LOG_FORMAT", originalLogFormat)
			}
		}()

		var logBuffer bytes.Buffer
		gin.DefaultWriter = &logBuffer

		router := gin.New()
		router.Use(Logger())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "TestAgent/1.0")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		logOutput := logBuffer.String()
		assert.Contains(t, logOutput, "GET")
		assert.Contains(t, logOutput, "/test")
		assert.Contains(t, logOutput, "200")
		assert.Contains(t, logOutput, "TestAgent/1.0")
	})

	t.Run("JSON格式日志", func(t *testing.T) {
		// 设置JSON日志格式
		originalLogFormat := os.Getenv("LOG_FORMAT")
		os.Setenv("LOG_FORMAT", "json")
		defer func() {
			if originalLogFormat == "" {
				os.Unsetenv("LOG_FORMAT")
			} else {
				os.Setenv("LOG_FORMAT", originalLogFormat)
			}
		}()

		var logBuffer bytes.Buffer
		gin.DefaultWriter = &logBuffer

		router := gin.New()
		router.Use(Logger())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "TestAgent/1.0")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		logOutput := strings.TrimSpace(logBuffer.String())
		
		// 验证是否为有效JSON
		var logData map[string]interface{}
		err := json.Unmarshal([]byte(logOutput), &logData)
		assert.NoError(t, err, "Log output should be valid JSON")

		// 验证JSON日志包含重要字段
		expectedFields := []string{"timestamp", "level", "method", "path", "status", "client_ip", "user_agent"}
		for _, field := range expectedFields {
			assert.Contains(t, logData, field, "JSON log should contain field: %s", field)
		}

		assert.Equal(t, "GET", logData["method"])
		assert.Equal(t, "/test", logData["path"])
		assert.Equal(t, float64(200), logData["status"]) // JSON数字解析为float64
		assert.Equal(t, "INFO", logData["level"])
	})
}

func TestJSONLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("JSON格式输出验证", func(t *testing.T) {
		var logBuffer bytes.Buffer
		gin.DefaultWriter = &logBuffer

		router := gin.New()
		router.Use(JSONLogger())
		router.GET("/api/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"result": "success"})
		})

		req := httptest.NewRequest("GET", "/api/test?param=value", nil)
		req.Header.Set("User-Agent", "TestClient/2.0")
		req.Header.Set("Referer", "https://example.com")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		logOutput := strings.TrimSpace(logBuffer.String())
		
		var logData map[string]interface{}
		err := json.Unmarshal([]byte(logOutput), &logData)
		assert.NoError(t, err)

		// 验证详细字段
		assert.Equal(t, "GET", logData["method"])
		// 路径可能包含查询参数，检查开头部分
		path, ok := logData["path"].(string)
		assert.True(t, ok)
		assert.True(t, strings.HasPrefix(path, "/api/test"), "Path should start with /api/test")
		assert.Equal(t, float64(200), logData["status"])
		assert.Equal(t, "INFO", logData["level"])
		assert.Equal(t, "TestClient/2.0", logData["user_agent"])
		assert.Equal(t, "https://example.com", logData["referer"])
		assert.Contains(t, logData, "timestamp")
		assert.Contains(t, logData, "latency_ms")
		assert.Contains(t, logData, "client_ip")
		assert.Contains(t, logData, "proto")
	})

	t.Run("跳过健康检查路径", func(t *testing.T) {
		var logBuffer bytes.Buffer
		gin.DefaultWriter = &logBuffer

		router := gin.New()
		router.Use(JSONLogger())
		router.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest("GET", "/health", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		logOutput := logBuffer.String()
		// 健康检查路径应该被跳过，不产生日志
		assert.Empty(t, strings.TrimSpace(logOutput))
	})
}

func TestErrorLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("错误日志记录", func(t *testing.T) {
		var errorBuffer bytes.Buffer
		gin.DefaultErrorWriter = &errorBuffer

		router := gin.New()
		router.Use(ErrorLogger())
		router.GET("/error", func(c *gin.Context) {
			c.Error(gin.Error{
				Err:  assert.AnError,
				Type: gin.ErrorTypePublic,
				Meta: "test error",
			})
			c.JSON(500, gin.H{"error": "internal error"})
		})

		req := httptest.NewRequest("GET", "/error", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		errorOutput := errorBuffer.String()
		assert.Contains(t, errorOutput, "ERROR")
		assert.Contains(t, errorOutput, "GET")
		assert.Contains(t, errorOutput, "/error")
	})

	t.Run("JSON格式错误日志", func(t *testing.T) {
		// 设置JSON日志格式
		originalLogFormat := os.Getenv("LOG_FORMAT")
		os.Setenv("LOG_FORMAT", "json")
		defer func() {
			if originalLogFormat == "" {
				os.Unsetenv("LOG_FORMAT")
			} else {
				os.Setenv("LOG_FORMAT", originalLogFormat)
			}
		}()

		var errorBuffer bytes.Buffer
		gin.DefaultErrorWriter = &errorBuffer

		router := gin.New()
		router.Use(ErrorLogger())
		router.GET("/json-error", func(c *gin.Context) {
			c.Error(gin.Error{
				Err:  assert.AnError,
				Type: gin.ErrorTypePublic,
			})
			c.JSON(400, gin.H{"error": "bad request"})
		})

		req := httptest.NewRequest("GET", "/json-error", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		errorOutput := strings.TrimSpace(errorBuffer.String())
		
		// 验证JSON格式错误日志
		var errorData map[string]interface{}
		err := json.Unmarshal([]byte(errorOutput), &errorData)
		assert.NoError(t, err, "Error log should be valid JSON")

		assert.Equal(t, "ERROR", errorData["level"])
		assert.Equal(t, "GET", errorData["method"])
		assert.Equal(t, "/json-error", errorData["path"])
		assert.Contains(t, errorData, "timestamp")
		assert.Contains(t, errorData, "client_ip")
		assert.Contains(t, errorData, "errors")
	})

	t.Run("无错误时不记录", func(t *testing.T) {
		var errorBuffer bytes.Buffer
		gin.DefaultErrorWriter = &errorBuffer

		router := gin.New()
		router.Use(ErrorLogger())
		router.GET("/success", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/success", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		errorOutput := errorBuffer.String()
		// 成功请求不应该产生错误日志
		assert.Empty(t, strings.TrimSpace(errorOutput))
	})
}

// 测试日志格式化函数
func TestLogFormatting(t *testing.T) {
	t.Run("日志时间戳格式", func(t *testing.T) {
		var logBuffer bytes.Buffer
		gin.DefaultWriter = &logBuffer

		router := gin.New()
		router.Use(JSONLogger())
		router.GET("/timestamp", func(c *gin.Context) {
			c.JSON(200, gin.H{"time": "now"})
		})

		req := httptest.NewRequest("GET", "/timestamp", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		logOutput := strings.TrimSpace(logBuffer.String())
		var logData map[string]interface{}
		json.Unmarshal([]byte(logOutput), &logData)

		// 验证时间戳格式（RFC3339）
		timestamp, ok := logData["timestamp"].(string)
		assert.True(t, ok, "Timestamp should be string")
		assert.Contains(t, timestamp, "T", "Timestamp should be in RFC3339 format")
		assert.True(t, len(timestamp) > 19, "Timestamp should include timezone")
	})

	t.Run("延迟时间单位", func(t *testing.T) {
		var logBuffer bytes.Buffer
		gin.DefaultWriter = &logBuffer

		router := gin.New()
		router.Use(JSONLogger())
		router.GET("/latency", func(c *gin.Context) {
			c.JSON(200, gin.H{"test": "latency"})
		})

		req := httptest.NewRequest("GET", "/latency", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		logOutput := strings.TrimSpace(logBuffer.String())
		var logData map[string]interface{}
		json.Unmarshal([]byte(logOutput), &logData)

		// 验证延迟时间是毫秒单位
		latency, ok := logData["latency_ms"].(float64)
		assert.True(t, ok, "Latency should be float64")
		assert.True(t, latency >= 0, "Latency should be non-negative")
	})
}