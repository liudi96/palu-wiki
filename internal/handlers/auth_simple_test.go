package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"palu-wiki/internal/config"
	"palu-wiki/internal/testutils"
)

// 测试输入验证 - 不需要数据库
func TestAuthHandler_InputValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 创建测试配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: testutils.TestJWTSecret,
		},
	}
	
	// 创建handler（暂时用nil DB，因为这个测试不会用到）
	handler := NewAuthHandler(nil, cfg)
	
	// 设置路由
	router := gin.New()
	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)
	
	t.Run("注册请求验证 - 用户名太短", func(t *testing.T) {
		reqBody := map[string]string{
			"username": "ab", // 太短
			"email":    "test@example.com",
			"password": "123456",
		}
		
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		
		var response map[string]string
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "min")
	})
	
	t.Run("登录请求验证 - 缺少用户名", func(t *testing.T) {
		reqBody := map[string]string{
			"password": "123456",
		}
		
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		
		var response map[string]string
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "required")
	})
	
	t.Run("RefreshToken - 缺少Authorization头", func(t *testing.T) {
		router.POST("/refresh", handler.RefreshToken)
		
		req := httptest.NewRequest("POST", "/refresh", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		
		var response map[string]string
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "缺少token")
	})
}

// 测试JWT工具函数
func TestJWT_Functions(t *testing.T) {
	require.NoError(t, testutils.InitTestTokens())
	
	t.Run("生成和解析有效token", func(t *testing.T) {
		token, err := testutils.GenerateTestToken(1, "testuser", "user")
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		
		claims, err := testutils.ParseTestToken(token)
		require.NoError(t, err)
		assert.Equal(t, uint(1), claims.UserID)
		assert.Equal(t, "testuser", claims.Username)
		assert.Equal(t, "user", claims.Role)
	})
	
	t.Run("解析无效token", func(t *testing.T) {
		invalidToken := "invalid.jwt.token"
		
		_, err := testutils.ParseTestToken(invalidToken)
		assert.Error(t, err)
	})
	
	t.Run("生成过期token", func(t *testing.T) {
		expiredToken, err := testutils.GenerateExpiredToken(1, "testuser", "user")
		require.NoError(t, err)
		
		_, err = testutils.ParseTestToken(expiredToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})
}

// 测试HTTP辅助工具
func TestHTTP_Helper(t *testing.T) {
	helper := testutils.NewHTTPTestHelper()
	helper.Router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	
	resp := helper.GET("/test")
	assert.Equal(t, 200, resp.Code)
	assert.Contains(t, resp.Body.String(), "success")
}