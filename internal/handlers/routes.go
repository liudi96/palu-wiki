package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"palu-wiki/internal/config"
	"palu-wiki/internal/middleware"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config, aiHandler *AIHandler) {
	// 创建处理器实例
	articleHandler := NewArticleHandler(db)
	categoryHandler := NewCategoryHandler(db)
	authHandler := NewAuthHandler(db, cfg)
	adminHandler := NewAdminHandler(db)

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 认证相关路由（无需认证）
		auth := v1.Group("/auth")
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
			ai.POST("/generate", aiHandler.GenerateArticle)
			ai.POST("/articles/:id/regenerate", aiHandler.RegenerateArticleContent)
			ai.POST("/articles/:id/optimize", aiHandler.OptimizeArticleContent)
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

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "Palu Wiki API is running",
		})
	})
}