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

	// 初始化AI管理器
	aiManager := ai.NewAIManager(ai.StrategyPriority)
	
	// 注册DeepSeek客户端（优先级1）
	deepSeekConfig := ai.LoadDeepSeekConfigFromEnv()
	if deepSeekConfig.APIKey != "" {
		deepSeekClient := ai.NewDeepSeekClient(deepSeekConfig)
		aiManager.RegisterProvider(deepSeekClient, ai.ProviderConfig{
			Name:     "deepseek",
			Priority: 1,
			Weight:   100,
			Enabled:  true,
		})
		log.Printf("DeepSeek客户端注册成功")
	} else {
		log.Printf("DeepSeek API Key未配置，跳过注册")
	}
	
	// 注册通义千问客户端（优先级2）
	qwenConfig := ai.LoadQwenConfigFromEnv()
	if qwenConfig.APIKey != "" {
		qwenClient := ai.NewQwenClient(qwenConfig)
		aiManager.RegisterProvider(qwenClient, ai.ProviderConfig{
			Name:     "qwen",
			Priority: 2,
			Weight:   80,
			Enabled:  true,
		})
		log.Printf("通义千问客户端注册成功")
	} else {
		log.Printf("通义千问 API Key未配置，跳过注册")
	}
	
	// 注册星火AI客户端（优先级3，兜底）
	sparkConfig := &ai.SparkAIConfig{
		AppID:     cfg.AI.SparkAppID,
		APIKey:    cfg.AI.SparkAPIKey,
		APISecret: cfg.AI.SparkAPISecret,
		Domain:    cfg.AI.SparkDomain,
		BaseURL:   cfg.AI.SparkBaseURL,
	}
	sparkClient, err := ai.NewSparkAIClient(sparkConfig)
	if err != nil {
		log.Printf("Failed to initialize Spark AI client: %v", err)
	} else {
		aiManager.RegisterProvider(sparkClient, ai.ProviderConfig{
			Name:     "spark",
			Priority: 3,
			Weight:   60,
			Enabled:  true,
		})
		log.Printf("星火AI客户端注册成功")
	}

	// 创建AI处理器
	aiHandler := handlers.NewAIHandler(db, aiManager)

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
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   180 * time.Second,
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
