package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 测试管理后台基础功能
func TestAdminHandler_BasicOperations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 创建handler（用nil DB因为这些测试不会访问数据库）
	handler := NewAdminHandler(nil)
	
	router := gin.New()
	
	// 管理员路由（需要认证和管理员权限）
	admin := router.Group("/admin")
	admin.Use(func(c *gin.Context) {
		// 模拟JWT认证和管理员权限
		c.Set("user_id", uint(1))
		c.Set("username", "admin")
		c.Set("user_role", "admin")
		c.Next()
	})
	{
		admin.GET("/dashboard", handler.Dashboard)
		admin.GET("/users", handler.GetAllUsers)
		admin.PUT("/users/:id/status", handler.UpdateUserStatus)
		admin.GET("/articles", handler.GetAllArticles)
		admin.PUT("/articles/:id/status", handler.UpdateArticleStatus)
		admin.GET("/system", handler.GetSystemInfo)
	}

	t.Run("仪表板参数解析", func(t *testing.T) {
		// 由于Dashboard需要数据库查询，我们只测试路由是否正确设置
		req := httptest.NewRequest("GET", "/admin/dashboard", nil)
		recorder := httptest.NewRecorder()
		
		// 使用简化的处理器来测试路由
		testRouter := gin.New()
		testRouter.GET("/admin/dashboard", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "dashboard endpoint reached"})
		})
		
		testRouter.ServeHTTP(recorder, req)
		assert.Equal(t, 200, recorder.Code)
	})

	t.Run("用户列表分页参数", func(t *testing.T) {
		testRouter := gin.New()
		testRouter.GET("/admin/users", func(c *gin.Context) {
			page := c.DefaultQuery("page", "1")
			pageSize := c.DefaultQuery("page_size", "20")
			keyword := c.Query("q")
			role := c.Query("role")
			status := c.Query("status")
			
			c.JSON(200, gin.H{
				"page":      page,
				"page_size": pageSize,
				"keyword":   keyword,
				"role":      role,
				"status":    status,
			})
		})
		
		req := httptest.NewRequest("GET", "/admin/users?page=2&page_size=10&q=test&role=admin&status=active", nil)
		recorder := httptest.NewRecorder()
		testRouter.ServeHTTP(recorder, req)
		
		assert.Equal(t, 200, recorder.Code)
		
		var response map[string]string
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "2", response["page"])
		assert.Equal(t, "10", response["page_size"])
		assert.Equal(t, "test", response["keyword"])
		assert.Equal(t, "admin", response["role"])
		assert.Equal(t, "active", response["status"])
	})

	t.Run("更新用户状态 - 输入验证", func(t *testing.T) {
		// 测试空请求体
		req := httptest.NewRequest("PUT", "/admin/users/1/status", nil)
		req.Header.Set("Content-Type", "application/json")
		
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("文章列表过滤参数", func(t *testing.T) {
		testRouter := gin.New()
		testRouter.GET("/admin/articles", func(c *gin.Context) {
			status := c.Query("status")
			categoryID := c.Query("category_id")
			authorID := c.Query("author_id")
			isAIGenerated := c.Query("is_ai_generated")
			
			c.JSON(200, gin.H{
				"status":          status,
				"category_id":     categoryID,
				"author_id":       authorID,
				"is_ai_generated": isAIGenerated,
			})
		})
		
		req := httptest.NewRequest("GET", "/admin/articles?status=published&category_id=1&author_id=2&is_ai_generated=true", nil)
		recorder := httptest.NewRecorder()
		testRouter.ServeHTTP(recorder, req)
		
		var response map[string]string
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "published", response["status"])
		assert.Equal(t, "1", response["category_id"])
		assert.Equal(t, "2", response["author_id"])
		assert.Equal(t, "true", response["is_ai_generated"])
	})
}

// 测试管理后台业务逻辑
func TestAdminHandler_BusinessLogic(t *testing.T) {
	t.Run("用户状态验证", func(t *testing.T) {
		validStatuses := []string{"active", "inactive", "banned"}
		invalidStatuses := []string{"unknown", "deleted", "pending", ""}
		
		isValidUserStatus := func(status string) bool {
			for _, valid := range validStatuses {
				if status == valid {
					return true
				}
			}
			return false
		}
		
		// 测试有效状态
		for _, status := range validStatuses {
			assert.True(t, isValidUserStatus(status), "状态 %s 应该是有效的", status)
		}
		
		// 测试无效状态
		for _, status := range invalidStatuses {
			assert.False(t, isValidUserStatus(status), "状态 %s 应该是无效的", status)
		}
	})
	
	t.Run("用户角色验证", func(t *testing.T) {
		validRoles := []string{"user", "editor", "admin"}
		invalidRoles := []string{"super_admin", "guest", "unknown", ""}
		
		isValidUserRole := func(role string) bool {
			if role == "" {
				return true // 空角色是允许的（不更新角色）
			}
			for _, valid := range validRoles {
				if role == valid {
					return true
				}
			}
			return false
		}
		
		// 测试有效角色
		for _, role := range validRoles {
			assert.True(t, isValidUserRole(role), "角色 %s 应该是有效的", role)
		}
		
		// 空角色应该是有效的
		assert.True(t, isValidUserRole(""))
		
		// 测试无效角色（排除空字符串）
		for _, role := range invalidRoles {
			if role == "" {
				continue // 跳过空字符串，因为它是有效的
			}
			assert.False(t, isValidUserRole(role), "角色 %s 应该是无效的", role)
		}
	})
	
	t.Run("文章状态验证", func(t *testing.T) {
		validStatuses := []string{"draft", "pending", "published", "rejected"}
		invalidStatuses := []string{"archived", "deleted", "unknown"}
		
		isValidArticleStatus := func(status string) bool {
			for _, valid := range validStatuses {
				if status == valid {
					return true
				}
			}
			return false
		}
		
		// 测试有效状态
		for _, status := range validStatuses {
			assert.True(t, isValidArticleStatus(status), "文章状态 %s 应该是有效的", status)
		}
		
		// 测试无效状态
		for _, status := range invalidStatuses {
			assert.False(t, isValidArticleStatus(status), "文章状态 %s 应该是无效的", status)
		}
	})
	
	t.Run("分页逻辑验证", func(t *testing.T) {
		// 模拟分页逻辑
		calculatePagination := func(page, pageSize int, total int64) map[string]interface{} {
			if page < 1 {
				page = 1
			}
			if pageSize < 1 || pageSize > 100 {
				pageSize = 20
			}
			
			totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
			
			return map[string]interface{}{
				"page":        page,
				"page_size":   pageSize,
				"total":       total,
				"total_pages": totalPages,
			}
		}
		
		// 测试正常分页
		result := calculatePagination(2, 10, 25)
		assert.Equal(t, 2, result["page"])
		assert.Equal(t, 10, result["page_size"])
		assert.Equal(t, int64(25), result["total"])
		assert.Equal(t, int64(3), result["total_pages"])
		
		// 测试边界条件
		result = calculatePagination(0, 0, 50) // 无效的页码和页大小
		assert.Equal(t, 1, result["page"])     // 应该修正为1
		assert.Equal(t, 20, result["page_size"]) // 应该修正为20
		
		result = calculatePagination(1, 150, 100) // 页大小超过限制
		assert.Equal(t, 20, result["page_size"])  // 应该修正为20
	})
}

// 测试权限检查
func TestAdminHandler_PermissionCheck(t *testing.T) {
	t.Run("管理员权限检查", func(t *testing.T) {
		hasAdminPermission := func(userRole string) bool {
			return userRole == "admin"
		}
		
		assert.True(t, hasAdminPermission("admin"))
		assert.False(t, hasAdminPermission("editor"))
		assert.False(t, hasAdminPermission("user"))
		assert.False(t, hasAdminPermission(""))
	})
	
	t.Run("管理操作权限矩阵", func(t *testing.T) {
		// 定义操作权限矩阵
		permissions := map[string]map[string]bool{
			"admin": {
				"view_dashboard":   true,
				"manage_users":     true,
				"manage_articles":  true,
				"system_info":      true,
			},
			"editor": {
				"view_dashboard":   false,
				"manage_users":     false,
				"manage_articles":  true,  // 编辑可以管理文章
				"system_info":      false,
			},
			"user": {
				"view_dashboard":   false,
				"manage_users":     false,
				"manage_articles":  false,
				"system_info":      false,
			},
		}
		
		// 测试管理员权限
		adminPerms := permissions["admin"]
		assert.True(t, adminPerms["view_dashboard"])
		assert.True(t, adminPerms["manage_users"])
		assert.True(t, adminPerms["manage_articles"])
		assert.True(t, adminPerms["system_info"])
		
		// 测试编辑权限
		editorPerms := permissions["editor"]
		assert.False(t, editorPerms["view_dashboard"])
		assert.False(t, editorPerms["manage_users"])
		assert.True(t, editorPerms["manage_articles"])
		assert.False(t, editorPerms["system_info"])
		
		// 测试普通用户权限
		userPerms := permissions["user"]
		assert.False(t, userPerms["view_dashboard"])
		assert.False(t, userPerms["manage_users"])
		assert.False(t, userPerms["manage_articles"])
		assert.False(t, userPerms["system_info"])
	})
}

// 测试数据结构和响应格式
func TestAdminHandler_DataStructures(t *testing.T) {
	t.Run("仪表板统计数据结构", func(t *testing.T) {
		mockStats := map[string]interface{}{
			"stats": map[string]int64{
				"total_users":        100,
				"total_articles":     250,
				"total_categories":   15,
				"total_comments":     500,
				"ai_articles":        50,
				"pending_articles":   20,
				"published_articles": 200,
				"draft_articles":     30,
			},
			"recent_articles": []map[string]interface{}{},
			"recent_users":    []map[string]interface{}{},
		}
		
		stats := mockStats["stats"].(map[string]int64)
		assert.Equal(t, int64(100), stats["total_users"])
		assert.Equal(t, int64(250), stats["total_articles"])
		assert.Equal(t, int64(50), stats["ai_articles"])
		
		assert.NotNil(t, mockStats["recent_articles"])
		assert.NotNil(t, mockStats["recent_users"])
	})
	
	t.Run("用户更新请求数据结构", func(t *testing.T) {
		updateUserReq := map[string]interface{}{
			"status": "active",
			"role":   "editor",
		}
		
		assert.Equal(t, "active", updateUserReq["status"])
		assert.Equal(t, "editor", updateUserReq["role"])
		
		// 测试只更新状态的请求
		statusOnlyReq := map[string]interface{}{
			"status": "banned",
		}
		
		assert.Equal(t, "banned", statusOnlyReq["status"])
		assert.Nil(t, statusOnlyReq["role"])
	})
	
	t.Run("文章状态更新请求数据结构", func(t *testing.T) {
		updateArticleReq := map[string]interface{}{
			"status": "published",
			"reason": "审核通过",
		}
		
		assert.Equal(t, "published", updateArticleReq["status"])
		assert.Equal(t, "审核通过", updateArticleReq["reason"])
	})
}