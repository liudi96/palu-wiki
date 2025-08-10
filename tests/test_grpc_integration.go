package main

import (
	"context"
	"log"
	"time"

	"palu-wiki/pkg/ai"
)

func main() {
	log.Println("🧪 Go gRPC客户端集成测试")
	log.Println("=" + string(make([]byte, 50)))
	
	// 创建context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// 初始化支持gRPC的AI管理器
	aiManager, err := ai.SetupGRPCAIManager(ctx)
	if err != nil {
		log.Fatalf("❌ AI管理器初始化失败: %v", err)
	}
	
	log.Println("✅ AI管理器初始化成功")
	
	// 测试1: 简单内容生成
	log.Println("\n🤖 测试1: 简单内容生成")
	content, err := aiManager.GenerateContent(ctx, "请简单介绍一下《幻兽帕鲁》这款游戏")
	if err != nil {
		log.Printf("❌ 内容生成失败: %v", err)
	} else {
		log.Printf("✅ 内容生成成功")
		log.Printf("📝 内容长度: %d 字符", len(content))
		log.Printf("📄 内容预览: %s...", truncateString(content, 100))
	}
	
	// 测试2: 文章生成
	log.Println("\n📝 测试2: 文章生成")
	article, err := aiManager.GenerateArticle(ctx, "帕鲁新手入门指南", "游戏攻略")
	if err != nil {
		log.Printf("❌ 文章生成失败: %v", err)
	} else {
		log.Printf("✅ 文章生成成功")
		log.Printf("📰 标题: %s", article.Title)
		log.Printf("📊 字数: %d 字符", len(article.Content))
		log.Printf("🏷️  标签: %v", article.Tags)
		log.Printf("💡 摘要: %s...", truncateString(article.Summary, 80))
		log.Printf("🔧 生成者: %s", article.GeneratedBy)
	}
	
	// 测试3: 获取统计信息
	log.Println("\n📊 测试3: 获取AI管理器统计")
	stats := aiManager.GetStats()
	log.Printf("📈 统计信息:")
	for provider, stat := range stats {
		log.Printf("  🔸 %s:", provider)
		log.Printf("    总请求: %d", stat.TotalRequests)
		log.Printf("    成功请求: %d", stat.SuccessCount)
		log.Printf("    失败请求: %d", stat.FailureCount)
		log.Printf("    平均延迟: %v", stat.AvgLatency)
		log.Printf("    最后使用: %v", stat.LastUsed.Format("2006-01-02 15:04:05"))
	}
	
	log.Println("\n" + string(make([]byte, 50)))
	log.Println("🎉 gRPC集成测试完成！")
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}