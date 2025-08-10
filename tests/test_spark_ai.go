package main

import (
	"context"
	"log"
	"time"

	"palu-wiki/internal/config"
	"palu-wiki/pkg/ai"
)

func main() {
	log.Println("🧪 测试星火AI接口")
	
	// 加载配置
	cfg := config.LoadConfig()
	
	// 创建context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// 创建星火AI客户端
	sparkConfig := &ai.SparkAIConfig{
		AppID:     cfg.AI.SparkAppID,
		APIKey:    cfg.AI.SparkAPIKey,
		APISecret: cfg.AI.SparkAPISecret,
		Domain:    cfg.AI.SparkDomain,
		BaseURL:   cfg.AI.SparkBaseURL,
	}
	
	sparkClient, err := ai.NewSparkAIClient(sparkConfig)
	if err != nil {
		log.Fatalf("❌ 星火AI客户端初始化失败: %v", err)
	}
	
	log.Println("✅ 星火AI客户端初始化成功")
	
	// 测试简单内容生成
	log.Println("\n🤖 测试内容生成...")
	content, err := sparkClient.GenerateContent(ctx, "请简单介绍一下《幻兽帕鲁》这款游戏")
	if err != nil {
		log.Printf("❌ 内容生成失败: %v", err)
	} else {
		log.Printf("✅ 内容生成成功")
		log.Printf("📝 内容长度: %d 字符", len(content))
		log.Printf("📄 内容预览: %s", content[:min(200, len(content))])
	}
	
	// 测试文章生成
	log.Println("\n📝 测试文章生成...")
	article, err := sparkClient.GenerateArticle(ctx, "帕鲁新手入门指南", "游戏攻略")
	if err != nil {
		log.Printf("❌ 文章生成失败: %v", err)
	} else {
		log.Printf("✅ 文章生成成功")
		log.Printf("📰 标题: %s", article.Title)
		log.Printf("📊 字数: %d 字符", len(article.Content))
		log.Printf("🏷️  标签数量: %d", len(article.Tags))
		log.Printf("💡 摘要: %s", article.Summary[:min(100, len(article.Summary))])
		log.Printf("🔧 生成者: %s", article.GeneratedBy)
	}
	
	log.Println("\n🎉 星火AI测试完成！")
}

// min 返回两个数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}