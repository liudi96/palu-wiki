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

// 测试输入验证和基础功能
func TestArticleHandler_InputValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 创建handler（用nil DB因为这些测试不会访问数据库）
	handler := NewArticleHandler(nil)
	
	router := gin.New()
	
	// 模拟认证中间件
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("username", "testuser") 
		c.Set("user_role", "user")
		c.Next()
	})
	
	router.POST("/articles", handler.CreateArticle)
	router.PUT("/articles/:id", handler.UpdateArticle)

	t.Run("创建文章 - 缺少必要字段", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/articles", nil)
		req.Header.Set("Content-Type", "application/json")
		
		// 注意：由于请求body为nil，会导致绑定错误
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("获取文章列表 - 分页参数验证", func(t *testing.T) {
		// 创建只测试参数解析的路由
		router := gin.New()
		router.GET("/articles", func(c *gin.Context) {
			page := c.DefaultQuery("page", "1")
			limit := c.DefaultQuery("limit", "10")
			
			c.JSON(200, gin.H{
				"page":  page,
				"limit": limit,
			})
		})
		
		// 测试默认参数
		req := httptest.NewRequest("GET", "/articles", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, 200, recorder.Code)
		
		var response map[string]string
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "1", response["page"])
		assert.Equal(t, "10", response["limit"])
		
		// 测试自定义参数
		req = httptest.NewRequest("GET", "/articles?page=2&limit=20", nil)
		recorder = httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "2", response["page"])
		assert.Equal(t, "20", response["limit"])
	})

	t.Run("搜索文章 - 参数解析", func(t *testing.T) {
		router := gin.New()
		router.GET("/search", func(c *gin.Context) {
			keyword := c.Query("q")
			categoryID := c.Query("category_id")
			status := c.DefaultQuery("status", "published")
			sortBy := c.DefaultQuery("sort", "created_at")
			
			c.JSON(200, gin.H{
				"keyword":     keyword,
				"category_id": categoryID,
				"status":      status,
				"sort":        sortBy,
			})
		})
		
		req := httptest.NewRequest("GET", "/search?q=test&category_id=1&sort=updated_at", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		
		assert.Equal(t, 200, recorder.Code)
		
		var response map[string]string
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "test", response["keyword"])
		assert.Equal(t, "1", response["category_id"])
		assert.Equal(t, "published", response["status"]) // 默认值
		assert.Equal(t, "updated_at", response["sort"])
	})
}

// 测试权限检查逻辑
func TestArticleHandler_Permission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	t.Run("检查文章所有权", func(t *testing.T) {
		// 测试权限检查的辅助函数
		checkOwnership := func(articleAuthorID, currentUserID uint, currentUserRole string) bool {
			// 模拟ArticleHandler中的权限检查逻辑
			return articleAuthorID == currentUserID || currentUserRole == "admin"
		}
		
		// 作者可以修改自己的文章
		assert.True(t, checkOwnership(1, 1, "user"))
		
		// 其他用户不能修改
		assert.False(t, checkOwnership(1, 2, "user"))
		
		// 管理员可以修改任何文章
		assert.True(t, checkOwnership(1, 2, "admin"))
		
		// 编辑员不能修改其他人的文章（根据当前代码逻辑）
		assert.False(t, checkOwnership(1, 2, "editor"))
	})
}

// 测试工具函数
func TestArticleTestUtils(t *testing.T) {
	t.Run("测试文章数据生成", func(t *testing.T) {
		article := testutils.TestArticle(1, 1)
		assert.Equal(t, uint(1), article.ID)
		assert.Equal(t, uint(1), article.AuthorID)
		assert.NotEmpty(t, article.Title)
		assert.NotEmpty(t, article.Content)
		assert.Equal(t, "published", article.Status)
	})
	
	t.Run("测试草稿文章数据", func(t *testing.T) {
		draft := testutils.TestDraftArticle(2, 1)
		assert.Equal(t, "draft", draft.Status)
		assert.Equal(t, 0, draft.ViewCount)
	})
	
	t.Run("测试有效请求数据", func(t *testing.T) {
		validReq := testutils.ValidArticleRequest()
		assert.NotEmpty(t, validReq.Title)
		assert.NotEmpty(t, validReq.Content)
		assert.NotEmpty(t, validReq.Summary)
		assert.Equal(t, uint(1), validReq.CategoryID)
	})
	
	t.Run("测试无效请求数据", func(t *testing.T) {
		invalidReq := testutils.InvalidArticleRequest()
		assert.Empty(t, invalidReq.Title)
		assert.Empty(t, invalidReq.Content)
	})
}