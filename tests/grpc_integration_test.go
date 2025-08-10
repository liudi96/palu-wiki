package tests

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"palu-wiki/pkg/ai"
	pb "palu-wiki/pkg/grpc/ai"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

// MockGRPCServer 模拟gRPC服务器用于集成测试
type MockGRPCServer struct {
	pb.UnimplementedAIServiceServer
}

func (s *MockGRPCServer) GenerateContent(ctx context.Context, req *pb.ContentRequest) (*pb.ContentResponse, error) {
	// 模拟AI内容生成
	response := &pb.ContentResponse{
		Content:        "这是基于提示'" + req.Prompt + "'生成的内容",
		Summary:        "内容摘要",
		Tags:          []string{"集成测试", "AI"},
		ModelUsed:     "integration-test-model",
		TokensUsed:    int32(len(req.Prompt) * 2), // 简单的token计算
		GenerationTime: 0.5,
		Confidence:    0.9,
		Sources:       []string{"mock_source"},
		Metadata:      make(map[string]string),
	}

	// 复制请求元数据到响应
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			response.Metadata[k] = v
		}
	}

	return response, nil
}

func (s *MockGRPCServer) GenerateArticle(ctx context.Context, req *pb.ArticleRequest) (*pb.ArticleResponse, error) {
	// 模拟AI文章生成
	content := "这是关于'" + req.Title + "'的详细文章内容。文章分为以下几个部分..."
	
	response := &pb.ArticleResponse{
		Title:          req.Title,
		Content:        content,
		Summary:        "关于" + req.Title + "的文章摘要",
		Tags:          append(req.Keywords, "集成测试"),
		Category:      req.Category,
		WordCount:     int32(len(content)),
		ModelUsed:     "article-generation-model",
		TokensUsed:    int32(len(content) / 4),
		GenerationTime: 1.2,
		Sources:       []string{"knowledge_base", "web_search"},
		Metadata:      make(map[string]string),
	}

	// 复制请求元数据
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			response.Metadata[k] = v
		}
	}

	return response, nil
}

func (s *MockGRPCServer) HealthCheck(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Status:      pb.HealthStatus_SERVING,
		ServiceName: "mock-ai-service",
		Version:     "test-1.0.0",
		Message:     "服务运行正常",
	}, nil
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func setupTestServer(t *testing.T) (*grpc.Server, func()) {
	lis = bufconn.Listen(bufSize)
	server := grpc.NewServer()
	pb.RegisterAIServiceServer(server, &MockGRPCServer{})

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("服务器启动失败: %v", err)
		}
	}()

	return server, func() {
		server.Stop()
		lis.Close()
	}
}

// TestGRPCIntegration 完整的gRPC集成测试
func TestGRPCIntegration(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// 创建gRPC连接
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", 
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewAIServiceClient(conn)

	t.Run("内容生成集成测试", func(t *testing.T) {
		// Given
		maxLength := int32(500)
		request := &pb.ContentRequest{
			Prompt:     "请介绍《幻兽帕鲁》游戏",
			Topic:      "游戏介绍",
			MaxLength:  &maxLength,
			Metadata: map[string]string{
				"test_id": "integration_001",
				"env":     "test",
			},
		}

		// When
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response, err := client.GenerateContent(ctx, request)

		// Then
		require.NoError(t, err)
		assert.NotEmpty(t, response.Content)
		assert.Contains(t, response.Content, "《幻兽帕鲁》")
		assert.Equal(t, "integration-test-model", response.ModelUsed)
		assert.Greater(t, response.TokensUsed, int32(0))
		assert.Greater(t, response.GenerationTime, float64(0))
		assert.Equal(t, "integration_001", response.Metadata["test_id"])
		assert.Equal(t, "test", response.Metadata["env"])

		// 验证标签和来源
		assert.Contains(t, response.Tags, "集成测试")
		assert.Contains(t, response.Sources, "mock_source")
	})

	t.Run("健康检查集成测试", func(t *testing.T) {
		// Given
		request := &pb.HealthRequest{}

		// When
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		response, err := client.HealthCheck(ctx, request)

		// Then
		require.NoError(t, err)
		assert.Equal(t, pb.HealthStatus_SERVING, response.Status)
		assert.Equal(t, "mock-ai-service", response.ServiceName)
		assert.Equal(t, "test-1.0.0", response.Version)
		assert.NotEmpty(t, response.Message)
	})
}