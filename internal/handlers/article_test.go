package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"palu-wiki/internal/testutils"
)

type ArticleHandlerTestSuite struct {
	suite.Suite
	handler *ArticleHandler
	testDB  *testutils.TestDB
	router  *gin.Engine
}

func (suite *ArticleHandlerTestSuite) SetupTest() {
	// 初始化测试数据库
	suite.testDB = testutils.SetupTestDB(suite.T())
	
	// 创建handler
	suite.handler = NewArticleHandler(suite.testDB.DB)
	
	// 设置测试路由
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	
	// 公开路由
	suite.router.GET("/articles", suite.handler.GetArticles)
	suite.router.GET("/articles/:id", suite.handler.GetArticle)
	suite.router.GET("/articles/search", suite.handler.SearchArticles)
	
	// 需要认证的路由
	auth := suite.router.Group("/auth")
	auth.Use(func(c *gin.Context) {
		// 模拟JWT中间件设置用户信息
		c.Set("user_id", uint(1))
		c.Set("username", "testuser")
		c.Set("user_role", "user")
		c.Next()
	})
	{
		auth.POST("/articles", suite.handler.CreateArticle)
		auth.PUT("/articles/:id", suite.handler.UpdateArticle)
		auth.DELETE("/articles/:id", suite.handler.DeleteArticle)
	}
	
	// 管理员路由
	admin := suite.router.Group("/admin")
	admin.Use(func(c *gin.Context) {
		// 模拟管理员权限
		c.Set("user_id", uint(3))
		c.Set("username", "admin")
		c.Set("user_role", "admin")
		c.Next()
	})
	{
		admin.DELETE("/articles/:id", suite.handler.DeleteArticle)
	}
}

func (suite *ArticleHandlerTestSuite) TearDownTest() {
	suite.testDB.ExpectationsWereMet(suite.T())
	suite.testDB.Close()
}

func TestArticleHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ArticleHandlerTestSuite))
}

// 测试获取文章列表
func (suite *ArticleHandlerTestSuite) TestGetArticles_Success() {
	// Given
	articles := []*testutils.ArticleRow{
		{
			ID: 1, Title: "文章1", Content: "内容1", Summary: "摘要1",
			AuthorID: 1, CategoryID: 1, Status: "published", ViewCount: 100,
			AuthorUsername: "testuser", AuthorNickname: "测试用户",
			CategoryName: "技术分享",
		},
		{
			ID: 2, Title: "文章2", Content: "内容2", Summary: "摘要2", 
			AuthorID: 2, CategoryID: 1, Status: "published", ViewCount: 50,
			AuthorUsername: "user2", AuthorNickname: "用户2",
			CategoryName: "技术分享",
		},
	}
	
	// Mock数据库查询
	rows := sqlmock.NewRows([]string{
		"id", "title", "content", "summary", "author_id", "category_id", 
		"status", "view_count", "created_at", "updated_at",
		"Author__id", "Author__username", "Author__nickname", "Author__role",
		"Category__id", "Category__name",
	})
	
	for _, article := range articles {
		rows.AddRow(
			article.ID, article.Title, article.Content, article.Summary,
			article.AuthorID, article.CategoryID, article.Status, article.ViewCount,
			sqlmock.AnyArg(), sqlmock.AnyArg(), // created_at, updated_at
			article.AuthorID, article.AuthorUsername, article.AuthorNickname, "user",
			article.CategoryID, article.CategoryName,
		)
	}
	
	suite.testDB.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(rows)

	// When
	req := httptest.NewRequest("GET", "/articles?page=1&limit=10", nil)
	recorder := httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)

	// Then
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	assert.NotNil(suite.T(), response["data"])
	assert.Equal(suite.T(), float64(1), response["page"])
	assert.Equal(suite.T(), float64(10), response["limit"])
}

// 测试获取文章详情 - 成功
func (suite *ArticleHandlerTestSuite) TestGetArticle_Success() {
	// Given
	article := testutils.TestArticle(1, 1)
	
	// Mock查询文章
	articleRows := sqlmock.NewRows([]string{
		"id", "title", "content", "summary", "author_id", "category_id",
		"status", "view_count", "created_at", "updated_at",
		"Author__id", "Author__username", "Author__nickname",
		"Category__id", "Category__name",
	}).AddRow(
		article.ID, article.Title, article.Content, article.Summary,
		article.AuthorID, article.CategoryID, article.Status, article.ViewCount,
		article.CreatedAt, article.UpdatedAt,
		1, "testuser", "测试用户",
		1, "技术分享",
	)
	
	suite.testDB.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WithArgs(1).
		WillReturnRows(articleRows)
	
	// Mock增加浏览量
	suite.testDB.Mock.ExpectBegin()
	suite.testDB.Mock.ExpectExec(regexp.QuoteMeta(`UPDATE "articles" SET "view_count"`)).
		WithArgs(article.ViewCount+1, article.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	suite.testDB.Mock.ExpectCommit()

	// When
	req := httptest.NewRequest("GET", "/articles/1", nil)
	recorder := httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)

	// Then
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), article.Title, data["title"])
}

// 测试获取文章详情 - 文章不存在
func (suite *ArticleHandlerTestSuite) TestGetArticle_NotFound() {
	// Given
	suite.testDB.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WithArgs(999).
		WillReturnError(gorm.ErrRecordNotFound)

	// When
	req := httptest.NewRequest("GET", "/articles/999", nil)
	recorder := httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)

	// Then
	assert.Equal(suite.T(), http.StatusNotFound, recorder.Code)
	
	var response map[string]string
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "文章不存在")
}

// 测试创建文章 - 成功
func (suite *ArticleHandlerTestSuite) TestCreateArticle_Success() {
	// Given
	reqData := testutils.ValidArticleRequest()
	
	// Mock创建文章
	suite.testDB.Mock.ExpectBegin()
	suite.testDB.Mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "articles"`)).
		WithArgs(
			reqData.Title, reqData.Content, reqData.Summary,
			uint(1), // 从JWT中间件设置的user_id
			reqData.CategoryID, reqData.Status, 0, // view_count starts at 0
			sqlmock.AnyArg(), sqlmock.AnyArg(), // created_at, updated_at
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	suite.testDB.Mock.ExpectCommit()

	// When
	req := testutils.NewHTTPTestHelper().POST("/auth/articles", reqData)

	// Then
	assert.Equal(suite.T(), http.StatusCreated, req.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(req.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	assert.True(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "创建成功")
	assert.NotNil(suite.T(), response["data"])
}

// 测试创建文章 - 无效数据
func (suite *ArticleHandlerTestSuite) TestCreateArticle_InvalidData() {
	// Given
	invalidData := testutils.InvalidArticleRequest()

	// When
	helper := testutils.NewHTTPTestHelper()
	helper.Router = suite.router
	resp := helper.POST("/auth/articles", invalidData)

	// Then
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
	
	var response map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), response["error"])
}

// 测试更新文章 - 作者成功更新
func (suite *ArticleHandlerTestSuite) TestUpdateArticle_AuthorSuccess() {
	// Given
	articleID := 1
	updateData := testutils.ValidArticleRequest()
	updateData.Title = "更新后的标题"
	
	// Mock查找文章
	suite.testDB.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "articles"`)).
		WithArgs(articleID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "content", "summary", "author_id", "category_id",
			"status", "view_count", "created_at", "updated_at",
		}).AddRow(
			1, "原标题", "原内容", "原摘要", 1, 1,
			"published", 100, sqlmock.AnyArg(), sqlmock.AnyArg(),
		))
	
	// Mock更新文章
	suite.testDB.Mock.ExpectBegin()
	suite.testDB.Mock.ExpectExec(regexp.QuoteMeta(`UPDATE "articles" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	suite.testDB.Mock.ExpectCommit()

	// When
	helper := testutils.NewHTTPTestHelper()
	helper.Router = suite.router
	resp := helper.PUT(fmt.Sprintf("/auth/articles/%d", articleID), updateData)

	// Then
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["data"])
}

// 测试更新文章 - 权限不足
func (suite *ArticleHandlerTestSuite) TestUpdateArticle_Forbidden() {
	// Given - 文章属于其他用户
	articleID := 1
	updateData := testutils.ValidArticleRequest()
	
	// Mock查找文章（author_id = 2，但当前用户是1）
	suite.testDB.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "articles"`)).
		WithArgs(articleID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "content", "summary", "author_id", "category_id",
			"status", "view_count", "created_at", "updated_at",
		}).AddRow(
			1, "他人文章", "他人内容", "他人摘要", 2, 1, // author_id = 2
			"published", 100, sqlmock.AnyArg(), sqlmock.AnyArg(),
		))

	// When
	helper := testutils.NewHTTPTestHelper()
	helper.Router = suite.router
	resp := helper.PUT(fmt.Sprintf("/auth/articles/%d", articleID), updateData)

	// Then
	assert.Equal(suite.T(), http.StatusForbidden, resp.Code)
	
	var response map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "无权限修改")
}

// 测试删除文章 - 管理员成功删除
func (suite *ArticleHandlerTestSuite) TestDeleteArticle_AdminSuccess() {
	// Given
	articleID := 1
	
	// Mock查找文章
	suite.testDB.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "articles"`)).
		WithArgs(articleID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "content", "summary", "author_id", "category_id",
			"status", "view_count", "created_at", "updated_at",
		}).AddRow(
			1, "待删除文章", "内容", "摘要", 2, 1, // 不是管理员的文章
			"published", 100, sqlmock.AnyArg(), sqlmock.AnyArg(),
		))
	
	// Mock软删除
	suite.testDB.Mock.ExpectBegin()
	suite.testDB.Mock.ExpectExec(regexp.QuoteMeta(`UPDATE "articles" SET "deleted_at"`)).
		WithArgs(sqlmock.AnyArg(), articleID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	suite.testDB.Mock.ExpectCommit()

	// When - 使用管理员路由
	helper := testutils.NewHTTPTestHelper()
	helper.Router = suite.router
	resp := helper.DELETE(fmt.Sprintf("/admin/articles/%d", articleID))

	// Then
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	
	var response map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["message"], "删除成功")
}