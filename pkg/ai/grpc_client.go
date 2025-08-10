package ai

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	
	"palu-wiki/pkg/grpc/ai"
)

// GRPCAIProvider gRPC AI提供商实现
type GRPCAIProvider struct {
	client     ai.AIServiceClient
	conn       *grpc.ClientConn
	serverAddr string
	name       string
	connected  bool
}

// NewGRPCAIProvider 创建新的gRPC AI提供商
func NewGRPCAIProvider(serverAddr string, name string) *GRPCAIProvider {
	return &GRPCAIProvider{
		serverAddr: serverAddr,
		name:       name,
		connected:  false,
	}
}

// Initialize 初始化连接
func (g *GRPCAIProvider) Initialize(ctx context.Context) error {
	// 配置连接参数
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	// 建立连接
	conn, err := grpc.Dial(g.serverAddr, opts...)
	if err != nil {
		return fmt.Errorf("连接gRPC服务器失败: %v", err)
	}

	g.conn = conn
	g.client = ai.NewAIServiceClient(conn)
	
	// 测试连接健康检查
	if err := g.healthCheck(ctx); err != nil {
		g.conn.Close()
		return fmt.Errorf("健康检查失败: %v", err)
	}
	
	g.connected = true
	log.Printf("✅ gRPC AI提供商 %s 连接成功: %s", g.name, g.serverAddr)
	
	return nil
}

// healthCheck 健康检查
func (g *GRPCAIProvider) healthCheck(ctx context.Context) error {
	deepCheck := false
	req := &ai.HealthRequest{DeepCheck: &deepCheck}
	
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	resp, err := g.client.HealthCheck(ctx, req)
	if err != nil {
		return err
	}
	
	if resp.Status != ai.HealthStatus_SERVING {
		message := ""
		if resp.Message != nil {
			message = *resp.Message
		}
		return fmt.Errorf("服务状态异常: %s", message)
	}
	
	return nil
}

// GenerateContent 生成内容
func (g *GRPCAIProvider) GenerateContent(ctx context.Context, prompt string) (string, error) {
	if !g.connected {
		return "", &AIError{
			Provider:  g.name,
			Code:      "not_connected",
			Message:   "gRPC客户端未连接",
			Retryable: true,
		}
	}
	
	maxLength := int32(500)
	req := &ai.ContentRequest{
		Prompt:    prompt,
		MaxLength: &maxLength, // 使用指针形式的可选字段
	}
	
	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	resp, err := g.client.GenerateContent(ctx, req)
	if err != nil {
		return "", &AIError{
			Provider:  g.name,
			Code:      "generation_failed",
			Message:   fmt.Sprintf("内容生成失败: %v", err),
			Retryable: true,
		}
	}
	
	return resp.Content, nil
}

// GenerateArticle 生成文章
func (g *GRPCAIProvider) GenerateArticle(ctx context.Context, title, topic string) (*ArticleContent, error) {
	if !g.connected {
		return nil, &AIError{
			Provider:  g.name,
			Code:      "not_connected",
			Message:   "gRPC客户端未连接",
			Retryable: true,
		}
	}
	
	length := int32(1000)
	req := &ai.ArticleRequest{
		Title:    title,
		Category: &topic,  // 使用指针形式的可选字段
		Length:   &length, // 使用指针形式的可选字段
	}
	
	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	
	resp, err := g.client.GenerateArticle(ctx, req)
	if err != nil {
		return nil, &AIError{
			Provider:  g.name,
			Code:      "generation_failed",
			Message:   fmt.Sprintf("文章生成失败: %v", err),
			Retryable: true,
		}
	}
	
	return &ArticleContent{
		Title:       resp.Title,
		Content:     resp.Content,
		Summary:     resp.Summary,
		Tags:        resp.Tags,
		GeneratedBy: g.name,
		GeneratedAt: time.Now(),
	}, nil
}

// GetProviderName 获取提供商名称
func (g *GRPCAIProvider) GetProviderName() string {
	return g.name
}

// IsAvailable 检查是否可用
func (g *GRPCAIProvider) IsAvailable(ctx context.Context) bool {
	if !g.connected {
		return false
	}
	
	// 快速健康检查
	err := g.healthCheck(ctx)
	return err == nil
}

// GetQuotaInfo 获取配额信息
func (g *GRPCAIProvider) GetQuotaInfo(ctx context.Context) (*QuotaInfo, error) {
	if !g.connected {
		return nil, &AIError{
			Provider:  g.name,
			Code:      "not_connected",
			Message:   "gRPC客户端未连接",
			Retryable: true,
		}
	}
	
	includeModels := true
	includeUsage := true
	req := &ai.StatsRequest{
		IncludeModels: &includeModels, // 使用指针形式的可选字段
		IncludeUsage:  &includeUsage,  // 使用指针形式的可选字段
	}
	
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	resp, err := g.client.GetServiceStats(ctx, req)
	if err != nil {
		return nil, &AIError{
			Provider:  g.name,
			Code:      "stats_failed",
			Message:   fmt.Sprintf("获取统计信息失败: %v", err),
			Retryable: true,
		}
	}
	
	// 计算总token数（简化处理）
	var totalTokens int64
	for _, model := range resp.Models {
		totalTokens += int64(model.TotalTokens)
	}
	
	return &QuotaInfo{
		Provider:       g.name,
		TotalTokens:    totalTokens,
		UsedTokens:     totalTokens, // 简化处理
		RemainingQuota: 1.0,         // 假设充足
		ResetTime:      time.Now().Add(24 * time.Hour),
	}, nil
}

// Close 关闭连接
func (g *GRPCAIProvider) Close() error {
	if g.conn != nil {
		g.connected = false
		return g.conn.Close()
	}
	return nil
}

// =================== 扩展功能 ===================

// SearchSimilar 相似内容搜索
func (g *GRPCAIProvider) SearchSimilar(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if !g.connected {
		return nil, &AIError{
			Provider:  g.name,
			Code:      "not_connected",
			Message:   "gRPC客户端未连接",
			Retryable: true,
		}
	}
	
	limitVal := int32(limit)
	threshold := float32(0.7)
	req := &ai.SearchRequest{
		Query:     query,
		Limit:     &limitVal,  // 使用指针形式的可选字段
		Threshold: &threshold, // 使用指针形式的可选字段
	}
	
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	
	resp, err := g.client.SearchSimilar(ctx, req)
	if err != nil {
		return nil, &AIError{
			Provider:  g.name,
			Code:      "search_failed",
			Message:   fmt.Sprintf("相似内容搜索失败: %v", err),
			Retryable: true,
		}
	}
	
	// 转换结果
	results := make([]SearchResult, len(resp.Results))
	for i, result := range resp.Results {
		results[i] = SearchResult{
			ID:         result.Id,
			Title:      result.Title,
			Content:    result.Content,
			Similarity: result.Similarity,
			Metadata:   result.Metadata,
		}
	}
	
	return results, nil
}

// StreamGenerateContent 流式内容生成
func (g *GRPCAIProvider) StreamGenerateContent(ctx context.Context, prompt string, callback func(chunk string) error) error {
	if !g.connected {
		return &AIError{
			Provider:  g.name,
			Code:      "not_connected",
			Message:   "gRPC客户端未连接",
			Retryable: true,
		}
	}
	
	maxLen := int32(500)
	req := &ai.ContentRequest{
		Prompt:    prompt,
		MaxLength: &maxLen, // 使用指针形式的可选字段
	}
	
	stream, err := g.client.StreamGenerateContent(ctx, req)
	if err != nil {
		return &AIError{
			Provider:  g.name,
			Code:      "stream_failed",
			Message:   fmt.Sprintf("流式生成启动失败: %v", err),
			Retryable: true,
		}
	}
	
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return &AIError{
				Provider:  g.name,
				Code:      "stream_error",
				Message:   fmt.Sprintf("流式接收失败: %v", err),
				Retryable: true,
			}
		}
		
		// 处理不同类型的块
		switch chunk.Type {
		case ai.ChunkType_CONTENT:
			if err := callback(chunk.Chunk); err != nil {
				return err
			}
		case ai.ChunkType_ERROR:
			return &AIError{
				Provider:  g.name,
				Code:      "generation_error",
				Message:   chunk.Chunk,
				Retryable: false,
			}
		case ai.ChunkType_FINAL:
			log.Printf("流式生成完成，元数据: %v", chunk.Metadata)
		}
		
		if chunk.IsFinal {
			break
		}
	}
	
	return nil
}

// GetServiceStats 获取服务统计
func (g *GRPCAIProvider) GetServiceStats(ctx context.Context) (*ServiceStats, error) {
	if !g.connected {
		return nil, &AIError{
			Provider:  g.name,
			Code:      "not_connected",
			Message:   "gRPC客户端未连接",
			Retryable: true,
		}
	}
	
	includeModels := true
	includeUsage := true
	req := &ai.StatsRequest{
		IncludeModels: &includeModels, // 使用指针形式的可选字段
		IncludeUsage:  &includeUsage,  // 使用指针形式的可选字段
	}
	
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	resp, err := g.client.GetServiceStats(ctx, req)
	if err != nil {
		return nil, &AIError{
			Provider:  g.name,
			Code:      "stats_failed",
			Message:   fmt.Sprintf("获取统计信息失败: %v", err),
			Retryable: true,
		}
	}
	
	return &ServiceStats{
		Provider:         g.name,
		Uptime:           time.Duration(resp.Uptime) * time.Second,
		TotalRequests:    resp.TotalRequests,    // 已经是int64
		SuccessRequests:  resp.SuccessRequests,  // 已经是int64
		ErrorRequests:    resp.ErrorRequests,    // 已经是int64
		AvgResponseTime:  time.Duration(resp.AvgResponseTime*1000) * time.Millisecond,
		MemoryUsage:      resp.MemoryUsage,      // 已经是float64
		VectorDBSize:     resp.VectorDbSize,
	}, nil
}

// =================== 辅助结构 ===================

// SearchResult 搜索结果
type SearchResult struct {
	ID         string            `json:"id"`
	Title      string            `json:"title"`
	Content    string            `json:"content"`
	Similarity float32           `json:"similarity"`
	Metadata   map[string]string `json:"metadata"`
}

// ServiceStats 服务统计
type ServiceStats struct {
	Provider         string        `json:"provider"`
	Uptime           time.Duration `json:"uptime"`
	TotalRequests    int64         `json:"total_requests"`    // 更新为int64
	SuccessRequests  int64         `json:"success_requests"`  // 更新为int64
	ErrorRequests    int64         `json:"error_requests"`    // 更新为int64
	AvgResponseTime  time.Duration `json:"avg_response_time"`
	MemoryUsage      float64       `json:"memory_usage"`      // 更新为float64
	VectorDBSize     int32         `json:"vector_db_size"`
}