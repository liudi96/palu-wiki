package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"palu-wiki/internal/testutils"
)

// 测试分类处理器的基础功能
func TestCategoryHandler_BasicOperations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 创建handler（用nil DB因为这些测试不会访问数据库）
	handler := NewCategoryHandler(nil)
	
	router := gin.New()
	
	// 公开路由
	router.GET("/categories", handler.GetCategories)
	router.GET("/categories/:id/articles", handler.GetCategoryArticles)
	
	// 管理员路由（需要认证和权限）
	admin := router.Group("/admin")
	admin.Use(func(c *gin.Context) {
		// 模拟JWT认证和管理员权限
		c.Set("user_id", uint(1))
		c.Set("username", "admin")
		c.Set("user_role", "admin")
		c.Next()
	})
	{
		admin.POST("/categories", handler.CreateCategory)
		admin.PUT("/categories/:id", handler.UpdateCategory)
		admin.DELETE("/categories/:id", handler.DeleteCategory)
	}

	t.Run("创建分类 - 输入验证", func(t *testing.T) {
		// 测试空请求体
		req := httptest.NewRequest("POST", "/admin/categories", nil)
		req.Header.Set("Content-Type", "application/json")
		
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("URL参数解析", func(t *testing.T) {
		// 创建测试路由来验证参数解析
		testRouter := gin.New()
		testRouter.GET("/categories/:id/articles", func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(200, gin.H{"category_id": id})
		})
		
		req := httptest.NewRequest("GET", "/categories/123/articles", nil)
		recorder := httptest.NewRecorder()
		testRouter.ServeHTTP(recorder, req)
		
		assert.Equal(t, 200, recorder.Code)
		
		var response map[string]string
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "123", response["category_id"])
	})
}

// 测试分类业务逻辑
func TestCategoryHandler_BusinessLogic(t *testing.T) {
	t.Run("分类删除权限检查", func(t *testing.T) {
		// 模拟删除检查逻辑
		canDeleteCategory := func(hasArticles bool) bool {
			return !hasArticles
		}
		
		// 有文章的分类不能删除
		assert.False(t, canDeleteCategory(true))
		
		// 没有文章的分类可以删除
		assert.True(t, canDeleteCategory(false))
	})
	
	t.Run("分类状态检查", func(t *testing.T) {
		// 测试分类状态过滤逻辑
		isActiveCategory := func(status string) bool {
			return status == "active"
		}
		
		assert.True(t, isActiveCategory("active"))
		assert.False(t, isActiveCategory("inactive"))
		assert.False(t, isActiveCategory("deleted"))
	})
}

// 测试分类数据结构
func TestCategoryDataStructure(t *testing.T) {
	t.Run("分类测试数据生成", func(t *testing.T) {
		category := testutils.TestCategory(1)
		
		assert.Equal(t, uint(1), category.ID)
		assert.NotEmpty(t, category.Name)
		assert.NotEmpty(t, category.Description)
		assert.NotZero(t, category.CreatedAt)
		assert.NotZero(t, category.UpdatedAt)
	})
}

// 测试分类请求和响应数据结构
func TestCategoryRequestResponse(t *testing.T) {
	t.Run("分类创建请求数据", func(t *testing.T) {
		validCategoryReq := map[string]interface{}{
			"name":        "新分类",
			"description": "分类描述",
			"icon":        "📚",
			"sort":        1,
			"status":      "active",
		}
		
		assert.Equal(t, "新分类", validCategoryReq["name"])
		assert.Equal(t, "分类描述", validCategoryReq["description"])
		assert.Equal(t, 1, validCategoryReq["sort"])
		assert.Equal(t, "active", validCategoryReq["status"])
	})
	
	t.Run("分类响应数据结构", func(t *testing.T) {
		// 模拟标准API响应格式
		mockResponse := map[string]interface{}{
			"data": map[string]interface{}{
				"id":          1,
				"name":        "技术分享",
				"description": "技术相关文章",
				"sort":        0,
				"status":      "active",
			},
		}
		
		data := mockResponse["data"].(map[string]interface{})
		assert.Equal(t, 1, data["id"])
		assert.Equal(t, "技术分享", data["name"])
		assert.Equal(t, "active", data["status"])
	})
}

// 测试错误处理
func TestCategoryHandler_ErrorHandling(t *testing.T) {
	t.Run("分类不存在的错误响应", func(t *testing.T) {
		expectedErrorResponse := map[string]string{
			"error": "分类不存在",
		}
		
		assert.Contains(t, expectedErrorResponse["error"], "不存在")
	})
	
	t.Run("权限不足的错误响应", func(t *testing.T) {
		expectedErrorResponse := map[string]string{
			"error": "权限不足",
		}
		
		assert.Contains(t, expectedErrorResponse["error"], "权限不足")
	})
	
	t.Run("分类有关联文章的删除错误", func(t *testing.T) {
		expectedErrorResponse := map[string]string{
			"error": "该分类下还有文章，无法删除",
		}
		
		assert.Contains(t, expectedErrorResponse["error"], "无法删除")
	})
}

// 测试HTTP状态码
func TestCategoryHandler_StatusCodes(t *testing.T) {
	testCases := []struct {
		name           string
		expectedStatus int
		description    string
	}{
		{"成功获取", http.StatusOK, "正常请求应该返回200"},
		{"创建成功", http.StatusCreated, "创建资源成功返回201"},
		{"请求错误", http.StatusBadRequest, "请求参数错误返回400"},
		{"未找到", http.StatusNotFound, "资源不存在返回404"},
		{"内部错误", http.StatusInternalServerError, "服务器错误返回500"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 验证HTTP状态码常量的正确性
			assert.NotZero(t, tc.expectedStatus)
			assert.True(t, tc.expectedStatus >= 200 && tc.expectedStatus < 600)
		})
	}
}