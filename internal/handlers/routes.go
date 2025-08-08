package handlers

import (
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"palu-wiki/internal/config"
	"palu-wiki/internal/middleware"
	"palu-wiki/pkg/database"
	"palu-wiki/pkg/redis"
)

var startTime = time.Now()

func SetupRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config, aiHandler *AIHandler) {
	// 创建处理器实例
	articleHandler := NewArticleHandler(db)
	categoryHandler := NewCategoryHandler(db)
	authHandler := NewAuthHandler(db, cfg)
	adminHandler := NewAdminHandler(db)

	// 全局安全中间件
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.SecureCORS())
	router.Use(middleware.InputValidation())
	router.Use(middleware.AntiBot())
	router.Use(middleware.ErrorLogger())

	// 全局限流
	router.Use(middleware.DefaultRateLimit())

	// API v1 路由组
	v1 := router.Group("/api/v1")
	v1.Use(middleware.APIRateLimit()) // API专用限流
	{
		// 认证相关路由（无需认证，但有严格限流）
		auth := v1.Group("/auth")
		auth.Use(middleware.StrictRateLimit()) // 登录限流
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// 用户相关路由（需要认证）
		user := v1.Group("/user")
		user.Use(middleware.JWTAuth(cfg))
		{
			user.GET("/profile", authHandler.GetProfile)
		}

		// 文章相关路由
		articles := v1.Group("/articles")
		{
			// 公开接口
			articles.GET("", articleHandler.GetArticles)
			articles.GET("/:id", articleHandler.GetArticle)
			articles.GET("/search", articleHandler.SearchArticles)

			// 需要认证的接口
			articlesAuth := articles.Group("")
			articlesAuth.Use(middleware.JWTAuth(cfg))
			{
				// 所有认证用户都可以创建文章（发布后需要审核）
				articlesAuth.POST("", articleHandler.CreateArticle)
				// 编辑和删除需要作者本人或管理员权限
				articlesAuth.PUT("/:id", middleware.RequireAdminOrEditor(), articleHandler.UpdateArticle)
				articlesAuth.DELETE("/:id", middleware.RequireAdmin(), articleHandler.DeleteArticle)
			}
		}

		// 分类相关路由
		categories := v1.Group("/categories")
		{
			// 公开接口
			categories.GET("", categoryHandler.GetCategories)
			categories.GET("/:id/articles", categoryHandler.GetCategoryArticles)

			// 需要管理员权限的接口
			categoriesAdmin := categories.Group("")
			categoriesAdmin.Use(middleware.JWTAuth(cfg), middleware.RequireAdmin())
			{
				categoriesAdmin.POST("", categoryHandler.CreateCategory)
				categoriesAdmin.PUT("/:id", categoryHandler.UpdateCategory)
				categoriesAdmin.DELETE("/:id", categoryHandler.DeleteCategory)
			}
		}

		// AI相关路由（需要认证）
		ai := v1.Group("/ai")
		ai.Use(middleware.JWTAuth(cfg))
		{
			ai.POST("/generate", aiHandler.GenerateContent) // 简单内容生成
			ai.POST("/article", aiHandler.GenerateArticle)  // 生成并保存文章
			ai.POST("/articles/:id/regenerate", aiHandler.RegenerateArticleContent)
			ai.POST("/articles/:id/optimize", aiHandler.OptimizeArticleContent)
		}

		// 文件上传相关路由（需要认证）
		upload := v1.Group("/upload")
		upload.Use(middleware.JWTAuth(cfg))
		{
			uploadHandler := NewUploadHandler(db)
			upload.POST("/image", uploadHandler.UploadImage)      // 上传图片
			upload.POST("/file", uploadHandler.UploadFile)        // 上传文件
			upload.GET("/files", uploadHandler.GetFiles)          // 获取文件列表
			upload.GET("/files/:id", uploadHandler.GetFile)       // 获取文件详情
			upload.PUT("/files/:id", uploadHandler.UpdateFile)    // 更新文件信息
			upload.DELETE("/files/:id", uploadHandler.DeleteFile) // 删除文件
		}

		// 管理后台路由（需要管理员权限）
		admin := v1.Group("/admin")
		admin.Use(middleware.JWTAuth(cfg), middleware.RequireAdmin())
		{
			// 仪表板
			admin.GET("/dashboard", adminHandler.Dashboard)
			admin.GET("/system", adminHandler.GetSystemInfo)

			// 用户管理
			admin.GET("/users", adminHandler.GetAllUsers)
			admin.PUT("/users/:id/status", adminHandler.UpdateUserStatus)

			// 文章管理
			admin.GET("/articles", adminHandler.GetAllArticles)
			admin.PUT("/articles/:id/status", adminHandler.UpdateArticleStatus)
		}
	}

	// 静态文件服务
	router.Static("/uploads", "./uploads")

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Palu Wiki API is running",
		})
	})

	// 详细健康检查和监控信息
	router.GET("/health/detailed", func(c *gin.Context) {
		// 检查数据库连接
		dbStatus := "ok"
		var dbError string
		if err := database.DB.Exec("SELECT 1").Error; err != nil {
			dbStatus = "error"
			dbError = err.Error()
		}

		// 检查Redis连接
		redisStatus := "ok"
		var redisError string
		if redisClient := redis.GetClient(); redisClient != nil {
			if err := redisClient.Ping(c.Request.Context()).Err(); err != nil {
				redisStatus = "error"
				redisError = err.Error()
			}
		} else {
			redisStatus = "not_configured"
		}

		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"services": gin.H{
				"database": gin.H{
					"status": dbStatus,
					"error":  dbError,
				},
				"redis": gin.H{
					"status": redisStatus,
					"error":  redisError,
				},
			},
		})
	})

	// 基础监控指标
	router.GET("/metrics", func(c *gin.Context) {
		// 获取系统基础指标
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// 数据库统计
		var dbStats gin.H
		if sqlDB, err := database.DB.DB(); err == nil {
			stats := sqlDB.Stats()
			dbStats = gin.H{
				"open_connections": stats.OpenConnections,
				"in_use":           stats.InUse,
				"idle":             stats.Idle,
			}
		}

		c.JSON(200, gin.H{
			"memory": gin.H{
				"alloc_mb":       float64(m.Alloc) / 1024 / 1024,
				"total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
				"sys_mb":         float64(m.Sys) / 1024 / 1024,
			},
			"goroutines":     runtime.NumGoroutine(),
			"database":       dbStats,
			"uptime_seconds": time.Since(startTime).Seconds(),
		})
	})
}
