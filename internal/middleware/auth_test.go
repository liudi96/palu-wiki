package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"palu-wiki/internal/config"
	"palu-wiki/internal/testutils"
)

func setupTestRouter(cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// 受保护的路由
	protected := router.Group("/protected")
	protected.Use(JWTAuth(cfg))
	{
		protected.GET("/profile", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			username, _ := c.Get("username")
			role, _ := c.Get("user_role")
			
			c.JSON(200, gin.H{
				"user_id":  userID,
				"username": username,
				"role":     role,
				"message":  "success",
			})
		})
	}
	
	// 需要管理员权限的路由
	admin := router.Group("/admin")
	admin.Use(JWTAuth(cfg), RequireAdmin())
	{
		admin.GET("/dashboard", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "admin dashboard"})
		})
	}
	
	// 需要编辑权限的路由
	editor := router.Group("/editor")
	editor.Use(JWTAuth(cfg), RequireAdminOrEditor())
	{
		editor.GET("/articles", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "editor articles"})
		})
	}
	
	return router
}

func TestJWTAuth(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: testutils.TestJWTSecret,
		},
	}
	
	router := setupTestRouter(cfg)
	require.NoError(t, testutils.InitTestTokens())

	t.Run("有效token访问成功", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected/profile", nil)
		req.Header.Set("Authorization", "Bearer "+testutils.TestTokens.User)
		
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "success")
		assert.Contains(t, recorder.Body.String(), "testuser")
	})
	
	t.Run("缺少Authorization头", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected/profile", nil)
		
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "缺少Authorization header")
	})
	
	t.Run("错误的Authorization格式", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected/profile", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "Authorization header格式错误")
	})
	
	t.Run("无效token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected/profile", nil)
		req.Header.Set("Authorization", "Bearer "+testutils.GenerateInvalidToken())
		
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "无效的token")
	})
	
	t.Run("过期token", func(t *testing.T) {
		expiredToken, err := testutils.GenerateExpiredToken(1, "testuser", "user")
		require.NoError(t, err)
		
		req := httptest.NewRequest("GET", "/protected/profile", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "无效的token")
	})
}

func TestRequireAdmin(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: testutils.TestJWTSecret,
		},
	}
	
	router := setupTestRouter(cfg)
	require.NoError(t, testutils.InitTestTokens())

	t.Run("管理员访问成功", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+testutils.TestTokens.Admin)
		
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "admin dashboard")
	})
	
	t.Run("普通用户访问被拒绝", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+testutils.TestTokens.User)
		
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusForbidden, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "权限不足")
	})
}

func TestRequireAdminOrEditor(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: testutils.TestJWTSecret,
		},
	}
	
	router := setupTestRouter(cfg)
	require.NoError(t, testutils.InitTestTokens())

	t.Run("编辑用户访问成功", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/editor/articles", nil)
		req.Header.Set("Authorization", "Bearer "+testutils.TestTokens.Editor)
		
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "editor articles")
	})
	
	t.Run("管理员也可以访问", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/editor/articles", nil)
		req.Header.Set("Authorization", "Bearer "+testutils.TestTokens.Admin)
		
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "editor articles")
	})
	
	t.Run("普通用户访问被拒绝", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/editor/articles", nil)
		req.Header.Set("Authorization", "Bearer "+testutils.TestTokens.User)
		
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusForbidden, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "权限不足")
	})
}

// 测试权限检查边界情况
func TestRoleValidation(t *testing.T) {
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// 手动设置上下文的路由（模拟JWT认证后的状态）
	router.GET("/test-admin", func(c *gin.Context) {
		c.Set("user_role", "admin")
		c.Next()
	}, RequireAdmin(), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	
	router.GET("/test-invalid-role", func(c *gin.Context) {
		c.Set("user_role", "invalid_role")
		c.Next()
	}, RequireAdmin(), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	
	t.Run("有效的管理员角色", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test-admin", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
	
	t.Run("无效的用户角色", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test-invalid-role", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusForbidden, recorder.Code)
	})
}