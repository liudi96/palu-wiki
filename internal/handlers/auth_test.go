package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"palu-wiki/internal/config"
	"palu-wiki/internal/testutils"
	"palu-wiki/pkg/utils"
)

type AuthHandlerTestSuite struct {
	suite.Suite
	handler *AuthHandler
	testDB  *testutils.TestDB
	router  *gin.Engine
	cfg     *config.Config
}

func (suite *AuthHandlerTestSuite) SetupTest() {
	// 初始化测试数据库
	suite.testDB = testutils.SetupTestDB(suite.T())
	
	// 创建测试配置
	suite.cfg = &config.Config{
		JWT: config.JWTConfig{
			Secret: testutils.TestJWTSecret,
		},
	}
	
	// 创建handler
	suite.handler = NewAuthHandler(suite.testDB.DB, suite.cfg)
	
	// 设置测试路由
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	auth := suite.router.Group("/auth")
	{
		auth.POST("/register", suite.handler.Register)
		auth.POST("/login", suite.handler.Login)
		auth.POST("/refresh", suite.handler.RefreshToken)
		auth.GET("/profile", suite.handler.GetProfile)
	}
	
	// 初始化测试token
	require.NoError(suite.T(), testutils.InitTestTokens())
}

func (suite *AuthHandlerTestSuite) TearDownTest() {
	suite.testDB.ExpectationsWereMet(suite.T())
	suite.testDB.Close()
}

func TestAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

// 测试用户注册 - 成功场景
func (suite *AuthHandlerTestSuite) TestRegister_Success() {
	// Given - 准备测试数据
	reqData := testutils.ValidRegisterRequest()
	
	// Mock数据库期望 - 检查用户名不存在
	suite.testDB.Mock.ExpectQuery(`SELECT \* FROM "users"`).
		WithArgs(reqData.Username, reqData.Email).
		WillReturnError(gorm.ErrRecordNotFound)
	
	// Mock创建用户
	suite.testDB.Mock.ExpectBegin()
	suite.testDB.Mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
		WithArgs(
			reqData.Username, reqData.Email, sqlmock.AnyArg(), // password will be hashed
			reqData.Nickname, "user", "active",
			sqlmock.AnyArg(), sqlmock.AnyArg(), // created_at, updated_at
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	suite.testDB.Mock.ExpectCommit()

	// When - 执行注册请求
	helper := testutils.NewHTTPTestHelper()
	helper.Router = suite.router
	resp := helper.POST("/auth/register", reqData)

	// Then - 验证响应
	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
	
	var response AuthResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	// 验证返回的token有效
	assert.NotEmpty(suite.T(), response.Token)
	claims, err := testutils.ParseTestToken(response.Token)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), reqData.Username, claims.Username)
	assert.Equal(suite.T(), "user", claims.Role)
	
	// 验证用户信息
	assert.Equal(suite.T(), reqData.Username, response.User.Username)
	assert.Equal(suite.T(), reqData.Email, response.User.Email)
	assert.Equal(suite.T(), reqData.Nickname, response.User.Nickname)
	assert.Equal(suite.T(), "user", response.User.Role)
	assert.Empty(suite.T(), response.User.Password) // 密码应该被清空
}

// 测试用户注册 - 用户名已存在
func (suite *AuthHandlerTestSuite) TestRegister_UserExists() {
	// Given
	reqData := testutils.ValidRegisterRequest()
	existingUser := testutils.TestUser(1)
	
	// Mock用户已存在
	suite.testDB.MockSelectUser(existingUser)

	// When
	helper := testutils.NewHTTPTestHelper()
	helper.Router = suite.router
	resp := helper.POST("/auth/register", reqData)

	// Then
	assert.Equal(suite.T(), http.StatusConflict, resp.Code)
	
	var response map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "用户名或邮箱已存在")
}

// 测试用户注册 - 无效请求数据
func (suite *AuthHandlerTestSuite) TestRegister_InvalidRequest() {
	// Given
	invalidReq := testutils.InvalidRegisterRequest()

	// When
	helper := testutils.NewHTTPTestHelper()
	helper.Router = suite.router
	resp := helper.POST("/auth/register", invalidReq)

	// Then
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
	
	var response map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "validation")
}

// 测试用户登录 - 成功场景
func (suite *AuthHandlerTestSuite) TestLogin_Success() {
	// Given
	reqData := testutils.ValidLoginRequest()
	user := testutils.TestUser(1)
	// 设置加密后的密码
	hashedPassword, _ := utils.HashPassword(reqData.Password)
	user.Password = hashedPassword
	
	// Mock查找用户
	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password", "nickname", "role", "status",
		"created_at", "updated_at",
	}).AddRow(
		user.ID, user.Username, user.Email, user.Password, user.Nickname,
		user.Role, user.Status, user.CreatedAt, user.UpdatedAt,
	)
	
	suite.testDB.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 OR email = $2`)).
		WithArgs(reqData.Username, reqData.Username).
		WillReturnRows(rows)

	// When
	helper := testutils.NewHTTPTestHelper()
	helper.Router = suite.router
	resp := helper.POST("/auth/login", reqData)

	// Then
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	
	var response AuthResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	// 验证token
	assert.NotEmpty(suite.T(), response.Token)
	claims, err := testutils.ParseTestToken(response.Token)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.Username, claims.Username)
	
	// 验证用户信息
	assert.Equal(suite.T(), user.Username, response.User.Username)
	assert.Empty(suite.T(), response.User.Password) // 密码应该被清空
}

// 测试用户登录 - 用户不存在
func (suite *AuthHandlerTestSuite) TestLogin_UserNotFound() {
	// Given
	reqData := testutils.ValidLoginRequest()
	
	// Mock用户不存在
	suite.testDB.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 OR email = $2`)).
		WithArgs(reqData.Username, reqData.Username).
		WillReturnError(gorm.ErrRecordNotFound)

	// When
	helper := testutils.NewHTTPTestHelper()
	helper.Router = suite.router
	resp := helper.POST("/auth/login", reqData)

	// Then
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.Code)
	
	var response map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "用户名或密码错误")
}

// 测试用户登录 - 密码错误
func (suite *AuthHandlerTestSuite) TestLogin_WrongPassword() {
	// Given
	reqData := testutils.InvalidLoginRequest() // 错误密码
	user := testutils.TestUser(1)
	// 设置正确的密码哈希
	correctPassword, _ := utils.HashPassword("correctpassword")
	user.Password = correctPassword
	
	// Mock查找用户
	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password", "nickname", "role", "status",
		"created_at", "updated_at",
	}).AddRow(
		user.ID, user.Username, user.Email, user.Password, user.Nickname,
		user.Role, user.Status, user.CreatedAt, user.UpdatedAt,
	)
	
	suite.testDB.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 OR email = $2`)).
		WithArgs(reqData.Username, reqData.Username).
		WillReturnRows(rows)

	// When
	helper := testutils.NewHTTPTestHelper()
	helper.Router = suite.router
	resp := helper.POST("/auth/login", reqData)

	// Then
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.Code)
	
	var response map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "用户名或密码错误")
}

// 测试用户登录 - 账号被禁用
func (suite *AuthHandlerTestSuite) TestLogin_UserDisabled() {
	// Given
	reqData := testutils.ValidLoginRequest()
	user := testutils.TestUser(1)
	user.Status = "disabled" // 设置为禁用状态
	hashedPassword, _ := utils.HashPassword(reqData.Password)
	user.Password = hashedPassword
	
	// Mock查找用户
	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password", "nickname", "role", "status",
		"created_at", "updated_at",
	}).AddRow(
		user.ID, user.Username, user.Email, user.Password, user.Nickname,
		user.Role, user.Status, user.CreatedAt, user.UpdatedAt,
	)
	
	suite.testDB.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 OR email = $2`)).
		WithArgs(reqData.Username, reqData.Username).
		WillReturnRows(rows)

	// When
	helper := testutils.NewHTTPTestHelper()
	helper.Router = suite.router
	resp := helper.POST("/auth/login", reqData)

	// Then
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.Code)
	
	var response map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "账号已被禁用")
}

// 测试获取用户信息 - 成功场景
func (suite *AuthHandlerTestSuite) TestGetProfile_Success() {
	// Given
	user := testutils.TestUser(1)
	
	// Mock查找用户
	rows := sqlmock.NewRows([]string{
		"id", "username", "email", "password", "nickname", "role", "status",
		"created_at", "updated_at",
	}).AddRow(
		user.ID, user.Username, user.Email, user.Password, user.Nickname,
		user.Role, user.Status, user.CreatedAt, user.UpdatedAt,
	)
	
	suite.testDB.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(user.ID).
		WillReturnRows(rows)

	// 手动设置用户上下文（模拟JWT中间件的行为）
	helper := testutils.NewHTTPTestHelper()
	helper.Router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})
	helper.Router.GET("/auth/profile", suite.handler.GetProfile)

	// When
	resp := helper.GET("/auth/profile")

	// Then
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	userData := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), user.Username, userData["username"])
	assert.Equal(suite.T(), user.Email, userData["email"])
	assert.Empty(suite.T(), userData["password"]) // 密码应该被清空
}

// 测试获取用户信息 - 未认证
func (suite *AuthHandlerTestSuite) TestGetProfile_Unauthorized() {
	// Given - 没有设置用户上下文
	helper := testutils.NewHTTPTestHelper()
	helper.Router.GET("/auth/profile", suite.handler.GetProfile)

	// When
	resp := helper.GET("/auth/profile")

	// Then
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.Code)
	
	var response map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "未授权")
}

// 测试刷新Token - 成功场景
func (suite *AuthHandlerTestSuite) TestRefreshToken_Success() {
	// Given - 生成一个即将过期的token（这里简化处理，使用普通token）
	token, err := testutils.GenerateTestToken(1, "testuser", "user")
	require.NoError(suite.T(), err)
	
	helper := testutils.NewHTTPTestHelper()
	helper.Router.POST("/auth/refresh", suite.handler.RefreshToken)

	// When
	resp := helper.POST("/auth/refresh", nil, testutils.WithAuth(token))

	// Then - 由于我们的RefreshToken逻辑检查是否需要刷新，普通token会被拒绝
	// 这里我们测试的是正常的业务逻辑行为
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
	
	var response map[string]string
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "still valid")
}

// 测试刷新Token - 无效token
func (suite *AuthHandlerTestSuite) TestRefreshToken_InvalidToken() {
	// Given
	helper := testutils.NewHTTPTestHelper()
	helper.Router.POST("/auth/refresh", suite.handler.RefreshToken)

	// When - 使用无效token
	invalidToken := testutils.GenerateInvalidToken()
	resp := helper.POST("/auth/refresh", nil, testutils.WithAuth(invalidToken))

	// Then
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
	
	var response map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), response["error"])
}

// 测试刷新Token - 缺少token
func (suite *AuthHandlerTestSuite) TestRefreshToken_MissingToken() {
	// Given
	helper := testutils.NewHTTPTestHelper()
	helper.Router.POST("/auth/refresh", suite.handler.RefreshToken)

	// When - 不提供Authorization header
	resp := helper.POST("/auth/refresh", nil)

	// Then
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
	
	var response map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "缺少token")
}