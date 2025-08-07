package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders 安全头中间件
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// XSS防护
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// 内容类型嗅探防护
		c.Header("X-Content-Type-Options", "nosniff")
		
		// 防止页面被嵌入iframe（点击劫持防护）
		c.Header("X-Frame-Options", "DENY")
		
		// Referrer策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// 隐藏服务器版本信息
		c.Header("Server", "Palu-Wiki")
		
		// HTTPS传输安全（仅在HTTPS环境下）
		if c.Request.Header.Get("X-Forwarded-Proto") == "https" || c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		
		// 内容安全策略
		csp := []string{
			"default-src 'self'",
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'", // Next.js需要unsafe-inline
			"style-src 'self' 'unsafe-inline'",               // CSS需要unsafe-inline
			"img-src 'self' data: https:",                     // 允许图片来源
			"font-src 'self'",
			"connect-src 'self'",
			"media-src 'self'",
			"object-src 'none'",
			"base-uri 'self'",
			"form-action 'self'",
			"frame-ancestors 'none'",
		}
		c.Header("Content-Security-Policy", strings.Join(csp, "; "))
		
		// 权限策略（Feature Policy的替代）
		permissions := []string{
			"geolocation=()",
			"microphone=()",
			"camera=()",
			"payment=()",
			"usb=()",
		}
		c.Header("Permissions-Policy", strings.Join(permissions, ", "))
		
		c.Next()
	}
}

// CORS配置
func SecureCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// 允许的源列表
		allowedOrigins := []string{
			"http://localhost:3001",
			"http://localhost:3000", 
		}
		
		// 从环境变量获取生产环境允许的源
		if corsOrigins := os.Getenv("CORS_ORIGINS"); corsOrigins != "" {
			prodOrigins := strings.Split(corsOrigins, ",")
			allowedOrigins = append(allowedOrigins, prodOrigins...)
		}
		
		// 检查请求来源是否被允许
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if strings.TrimSpace(allowedOrigin) == origin {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24小时
		
		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

// 输入验证中间件
func InputValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求体大小
		if c.Request.ContentLength > 10*1024*1024 { // 10MB限制
			c.JSON(400, gin.H{
				"error": "请求体过大",
				"code":  "REQUEST_TOO_LARGE",
			})
			c.Abort()
			return
		}
		
		// 检查User-Agent（防止空User-Agent的爬虫）
		userAgent := c.Request.Header.Get("User-Agent")
		if userAgent == "" && c.Request.Method != "OPTIONS" {
			c.JSON(400, gin.H{
				"error": "缺少User-Agent",
				"code":  "MISSING_USER_AGENT",
			})
			c.Abort()
			return
		}
		
		// 检查可疑的路径遍历尝试
		path := c.Request.URL.Path
		if strings.Contains(path, "..") || strings.Contains(path, "~") {
			c.JSON(400, gin.H{
				"error": "无效的路径",
				"code":  "INVALID_PATH",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// 反爬虫中间件（简单版本）
func AntiBot() gin.HandlerFunc {
	return func(c *gin.Context) {
		userAgent := strings.ToLower(c.Request.Header.Get("User-Agent"))
		
		// 常见爬虫特征
		botSignatures := []string{
			"bot", "crawler", "spider", "scraper", "wget", "curl",
			"python-requests", "python-urllib", "java/", "apache-httpclient",
		}
		
		for _, signature := range botSignatures {
			if strings.Contains(userAgent, signature) {
				// 记录可疑访问
				c.Set("is_bot", true)
				break
			}
		}
		
		c.Next()
	}
}