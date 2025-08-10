package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	pb "palu-wiki/pkg/grpc/ai"
)

// BenchmarkGRPCAIProvider_GenerateContent 内容生成性能基准测试
func BenchmarkGRPCAIProvider_GenerateContent(b *testing.B) {
	// Setup
	mockClient := &MockAIServiceClient{}
	provider := &GRPCAIProvider{
		client:    mockClient,
		connected: true,
		name:      "benchmark-provider",
	}

	// 模拟响应
	response := &pb.ContentResponse{
		Content:        "这是性能测试生成的内容",
		ModelUsed:      "benchmark-model",
		TokensUsed:     20,
		GenerationTime: 0.1,
	}

	mockClient.On("GenerateContent", mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	ctx := context.Background()
	prompt := "性能测试提示"

	// Reset timer before benchmark
	b.ResetTimer()

	// Run benchmark
	for i := 0; i < b.N; i++ {
		_, err := provider.GenerateContent(ctx, prompt)
		if err != nil {
			b.Fatalf("生成内容失败: %v", err)
		}
	}
}

// BenchmarkGRPCAIProvider_GenerateArticle 文章生成性能基准测试
func BenchmarkGRPCAIProvider_GenerateArticle(b *testing.B) {
	// Setup
	mockClient := &MockAIServiceClient{}
	provider := &GRPCAIProvider{
		client:    mockClient,
		connected: true,
		name:      "benchmark-provider",
	}

	// 模拟文章响应
	response := &pb.ArticleResponse{
		Title:          "基准测试文章",
		Content:        "这是一篇用于基准测试的文章内容...",
		Summary:        "基准测试摘要",
		Tags:           []string{"基准", "测试"},
		WordCount:      100,
		ModelUsed:      "article-benchmark-model",
		TokensUsed:     50,
		GenerationTime: 0.5,
	}

	mockClient.On("GenerateArticle", mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	ctx := context.Background()
	title := "基准测试文章标题"
	category := "技术"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := provider.GenerateArticle(ctx, title, category)
		if err != nil {
			b.Fatalf("文章生成失败: %v", err)
		}
	}
}