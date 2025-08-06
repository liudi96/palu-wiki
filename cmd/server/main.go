package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"palu-wiki/internal/config"
	"palu-wiki/internal/handlers"
	"palu-wiki/internal/middleware"
	"palu-wiki/pkg/ai"
	"palu-wiki/pkg/database"
	"palu-wiki/pkg/redis"
)

// min 返回两个数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库连接
	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 初始化Redis连接
	rdb, err := redis.NewRedisClient(cfg)
	if err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
		// Redis连接失败不退出程序，仅记录日志
	} else {
		defer rdb.Close()
	}

	// 初始化AI客户端
	sparkConfig := &ai.SparkAIConfig{
		AppID:     cfg.AI.SparkAppID,
		APIKey:    cfg.AI.SparkAPIKey,
		APISecret: cfg.AI.SparkAPISecret,
		Domain:    cfg.AI.SparkDomain,
		BaseURL:   cfg.AI.SparkBaseURL,
	}
	aiClient, err := ai.NewSparkAIClient(sparkConfig)
	if err != nil {
		log.Printf("Failed to initialize AI client: %v", err)
		log.Printf("AI功能将不可用，请检查环境变量配置")
		log.Printf("当前配置：AppID=%s, APIKey=%s...", sparkConfig.AppID, sparkConfig.APIKey[:min(10, len(sparkConfig.APIKey))])
		// AI客户端初始化失败不退出程序，但AI功能不可用
	} else {
		log.Printf("AI客户端初始化成功，使用星火AI")
	}

	// 创建AI处理器
	aiHandler := handlers.NewAIHandler(db, aiClient)

	// 创建Gin路由器
	router := gin.New()

	// 添加中间件
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(gin.Recovery())

	// 设置路由
	handlers.SetupRoutes(router, db, cfg, aiHandler)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:           ":" + cfg.Server.Port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Server starting on port %s", cfg.Server.Port)
	log.Printf("Health check: http://localhost:%s/health", cfg.Server.Port)
	log.Printf("API base URL: http://localhost:%s/api/v1", cfg.Server.Port)

	// 启动服务器
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}
}