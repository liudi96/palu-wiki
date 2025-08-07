package utils

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

// 常用验证规则
var (
	EmailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	UsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`)
	PasswordRegex = regexp.MustCompile(`^.{6,50}$`) // 至少6位，最多50位
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validator 验证器
type Validator struct {
	errors []ValidationError
}

func NewValidator() *Validator {
	return &Validator{
		errors: make([]ValidationError, 0),
	}
}

// 添加错误
func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// 检查是否有错误
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// 获取错误列表
func (v *Validator) GetErrors() []ValidationError {
	return v.errors
}

// 字符串验证
func (v *Validator) ValidateString(field, value string, minLen, maxLen int, required bool) *Validator {
	value = strings.TrimSpace(value)
	
	if required && value == "" {
		v.AddError(field, "不能为空")
		return v
	}
	
	if value == "" && !required {
		return v
	}
	
	length := utf8.RuneCountInString(value)
	if length < minLen {
		v.AddError(field, fmt.Sprintf("长度不能少于%d个字符", minLen))
	}
	
	if maxLen > 0 && length > maxLen {
		v.AddError(field, fmt.Sprintf("长度不能超过%d个字符", maxLen))
	}
	
	return v
}

// 邮箱验证
func (v *Validator) ValidateEmail(field, email string, required bool) *Validator {
	email = strings.TrimSpace(email)
	
	if required && email == "" {
		v.AddError(field, "邮箱不能为空")
		return v
	}
	
	if email == "" && !required {
		return v
	}
	
	if !EmailRegex.MatchString(email) {
		v.AddError(field, "邮箱格式不正确")
	}
	
	return v
}

// 用户名验证
func (v *Validator) ValidateUsername(field, username string) *Validator {
	username = strings.TrimSpace(username)
	
	if username == "" {
		v.AddError(field, "用户名不能为空")
		return v
	}
	
	if !UsernameRegex.MatchString(username) {
		v.AddError(field, "用户名只能包含字母、数字、下划线和短横线，长度3-20位")
	}
	
	return v
}

// 密码验证
func (v *Validator) ValidatePassword(field, password string) *Validator {
	if password == "" {
		v.AddError(field, "密码不能为空")
		return v
	}
	
	if !PasswordRegex.MatchString(password) {
		v.AddError(field, "密码长度必须在6-50位之间")
	}
	
	return v
}

// 整数验证
func (v *Validator) ValidateInt(field string, value, min, max int) *Validator {
	if value < min {
		v.AddError(field, fmt.Sprintf("不能小于%d", min))
	}
	
	if max > 0 && value > max {
		v.AddError(field, fmt.Sprintf("不能大于%d", max))
	}
	
	return v
}

// 枚举验证
func (v *Validator) ValidateEnum(field, value string, allowed []string) *Validator {
	value = strings.TrimSpace(value)
	
	if value == "" {
		v.AddError(field, "不能为空")
		return v
	}
	
	for _, allowedValue := range allowed {
		if value == allowedValue {
			return v
		}
	}
	
	v.AddError(field, fmt.Sprintf("必须是以下值之一: %s", strings.Join(allowed, ", ")))
	return v
}

// XSS防护 - 清理HTML
func SanitizeHTML(input string) string {
	// 基本的HTML转义
	sanitized := html.EscapeString(input)
	
	// 移除可能的脚本标签
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	sanitized = scriptRegex.ReplaceAllString(sanitized, "")
	
	return sanitized
}

// SQL注入防护 - 检查可疑字符
func CheckSQLInjection(input string) bool {
	input = strings.ToLower(input)
	
	// 常见的SQL注入模式
	patterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"union", "select", "insert", "update", "delete",
		"drop", "create", "alter", "exec", "execute",
	}
	
	for _, pattern := range patterns {
		if strings.Contains(input, pattern) {
			return true
		}
	}
	
	return false
}

// 文件名安全检查
func ValidateFileName(filename string) bool {
	// 检查文件名长度
	if len(filename) == 0 || len(filename) > 255 {
		return false
	}
	
	// 检查危险字符
	dangerousChars := []string{
		"..", "/", "\\", ":", "*", "?", "\"", "<", ">", "|",
		"\x00", "\x01", "\x02", "\x03", "\x04", "\x05",
	}
	
	for _, char := range dangerousChars {
		if strings.Contains(filename, char) {
			return false
		}
	}
	
	// 检查文件扩展名
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf", ".txt", ".md"}
	filename = strings.ToLower(filename)
	
	hasValidExt := false
	for _, ext := range allowedExts {
		if strings.HasSuffix(filename, ext) {
			hasValidExt = true
			break
		}
	}
	
	return hasValidExt
}

// IP地址验证
func ValidateIP(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}
	
	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		
		for _, char := range part {
			if char < '0' || char > '9' {
				return false
			}
		}
	}
	
	return true
}