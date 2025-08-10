package testutils

import (
	"time"

	"palu-wiki/internal/models"
)

// TestUser 创建测试用户数据
func TestUser(id uint) *models.User {
	return &models.User{
		ID:        id,
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "$2a$10$hashedpassword", // bcrypt哈希后的密码
		Nickname:  "测试用户",
		Role:      "user",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// TestAdminUser 创建管理员测试用户
func TestAdminUser(id uint) *models.User {
	user := TestUser(id)
	user.Username = "admin"
	user.Email = "admin@example.com"
	user.Role = "admin"
	user.Nickname = "管理员"
	return user
}

// TestEditorUser 创建编辑用户
func TestEditorUser(id uint) *models.User {
	user := TestUser(id)
	user.Username = "editor"
	user.Email = "editor@example.com"
	user.Role = "editor"
	user.Nickname = "编辑员"
	return user
}

// TestArticle 创建测试文章数据
func TestArticle(id uint, authorID uint) *models.Article {
	return &models.Article{
		ID:         id,
		Title:      "测试文章标题",
		Content:    "这是一篇测试文章的详细内容...",
		Summary:    "测试文章摘要",
		AuthorID:   authorID,
		CategoryID: 1,
		Status:     "published",
		ViewCount:  100,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// TestDraftArticle 创建草稿文章
func TestDraftArticle(id uint, authorID uint) *models.Article {
	article := TestArticle(id, authorID)
	article.Status = "draft"
	article.ViewCount = 0
	return article
}

// TestCategory 创建测试分类数据
func TestCategory(id uint) *models.Category {
	return &models.Category{
		ID:          id,
		Name:        "技术分享",
		Description: "技术相关的分享文章",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// TestComment 创建测试评论数据
func TestComment(id uint, articleID uint, userID uint) *models.Comment {
	return &models.Comment{
		ID:        id,
		ArticleID: articleID,
		UserID:    userID,
		Content:   "这是一条测试评论",
		Status:    "approved",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// TestFile 创建测试上传文件数据
func TestFile(id uint, userID uint) *models.File {
	return &models.File{
		ID:          id,
		FileName:    "test-image.jpg",
		StoredName:  "test-image-stored.jpg",
		FileSize:    1024000, // 1MB
		FileType:    "jpg",
		MimeType:    "image/jpeg",
		FilePath:    "/uploads/test-image.jpg",
		FileURL:     "/uploads/test-image.jpg",
		Width:       800,
		Height:      600,
		UploaderID:  userID,
		Status:      "active",
		Description: "测试图片",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// LoginRequestData 登录请求测试数据
type LoginRequestData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ValidLoginRequest 有效的登录请求
func ValidLoginRequest() LoginRequestData {
	return LoginRequestData{
		Username: "testuser",
		Password: "123456",
	}
}

// InvalidLoginRequest 无效的登录请求
func InvalidLoginRequest() LoginRequestData {
	return LoginRequestData{
		Username: "testuser",
		Password: "wrongpassword",
	}
}

// RegisterRequestData 注册请求测试数据
type RegisterRequestData struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

// ValidRegisterRequest 有效的注册请求
func ValidRegisterRequest() RegisterRequestData {
	return RegisterRequestData{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "123456",
		Nickname: "新用户",
	}
}

// InvalidRegisterRequest 无效的注册请求（用户名太短）
func InvalidRegisterRequest() RegisterRequestData {
	return RegisterRequestData{
		Username: "nu", // 太短
		Email:    "invalid-email", // 无效邮箱
		Password: "123", // 密码太短
		Nickname: "新用户",
	}
}

// ArticleRequestData 文章请求测试数据
type ArticleRequestData struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	Summary    string `json:"summary"`
	CategoryID uint   `json:"category_id"`
	Status     string `json:"status"`
}

// ValidArticleRequest 有效的文章创建请求
func ValidArticleRequest() ArticleRequestData {
	return ArticleRequestData{
		Title:      "新文章标题",
		Content:    "这是新文章的详细内容...",
		Summary:    "新文章摘要",
		CategoryID: 1,
		Status:     "draft",
	}
}

// InvalidArticleRequest 无效的文章创建请求
func InvalidArticleRequest() ArticleRequestData {
	return ArticleRequestData{
		Title:   "", // 标题为空
		Content: "", // 内容为空
	}
}

// ArticleRow 用于数据库Mock的文章行数据
type ArticleRow struct {
	ID             uint
	Title          string
	Content        string
	Summary        string
	AuthorID       uint
	CategoryID     uint
	Status         string
	ViewCount      int
	AuthorUsername string
	AuthorNickname string
	CategoryName   string
}