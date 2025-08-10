package testutils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"palu-wiki/pkg/utils"
)

const (
	// TestJWTSecret 测试用的JWT密钥
	TestJWTSecret = "test-jwt-secret-key-for-testing-only"
)

// GenerateTestToken 生成测试用的JWT token
func GenerateTestToken(userID uint, username string, role string) (string, error) {
	return utils.GenerateToken(userID, username, role, TestJWTSecret)
}

// GenerateExpiredToken 生成过期的JWT token
func GenerateExpiredToken(userID uint, username string, role string) (string, error) {
	claims := &utils.Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // 过期1小时
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)), // 签发时间2小时前
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(TestJWTSecret))
}

// GenerateInvalidToken 生成格式错误的token
func GenerateInvalidToken() string {
	return "invalid.jwt.token"
}

// ParseTestToken 解析测试token
func ParseTestToken(tokenString string) (*utils.Claims, error) {
	return utils.ParseToken(tokenString, TestJWTSecret)
}

// TestTokens 常用的测试token集合
var TestTokens = struct {
	User   string
	Editor string 
	Admin  string
}{
	User:   "",
	Editor: "",
	Admin:  "",
}

// InitTestTokens 初始化测试token（在测试开始前调用）
func InitTestTokens() error {
	var err error
	
	TestTokens.User, err = GenerateTestToken(1, "testuser", "user")
	if err != nil {
		return err
	}
	
	TestTokens.Editor, err = GenerateTestToken(2, "editor", "editor")
	if err != nil {
		return err
	}
	
	TestTokens.Admin, err = GenerateTestToken(3, "admin", "admin")
	if err != nil {
		return err
	}
	
	return nil
}