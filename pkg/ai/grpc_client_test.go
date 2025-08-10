package ai

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"palu-wiki/pkg/grpc/ai"
)

// MockAIServiceClient 模拟的gRPC客户端
type MockAIServiceClient struct {
	mock.Mock
}

func (m *MockAIServiceClient) GenerateContent(ctx context.Context, in *ai.ContentRequest, opts ...grpc.CallOption) (*ai.ContentResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*ai.ContentResponse), args.Error(1)
}

func (m *MockAIServiceClient) GenerateArticle(ctx context.Context, in *ai.ArticleRequest, opts ...grpc.CallOption) (*ai.ArticleResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*ai.ArticleResponse), args.Error(1)
}

func (m *MockAIServiceClient) SearchSimilar(ctx context.Context, in *ai.SearchRequest, opts ...grpc.CallOption) (*ai.SearchResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*ai.SearchResponse), args.Error(1)
}

func (m *MockAIServiceClient) StreamGenerateContent(ctx context.Context, in *ai.ContentRequest, opts ...grpc.CallOption) (ai.AIService_StreamGenerateContentClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(ai.AIService_StreamGenerateContentClient), args.Error(1)
}

func (m *MockAIServiceClient) BatchGenerate(ctx context.Context, in *ai.BatchRequest, opts ...grpc.CallOption) (*ai.BatchResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*ai.BatchResponse), args.Error(1)
}

func (m *MockAIServiceClient) GetServiceStats(ctx context.Context, in *ai.StatsRequest, opts ...grpc.CallOption) (*ai.StatsResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*ai.StatsResponse), args.Error(1)
}

func (m *MockAIServiceClient) HealthCheck(ctx context.Context, in *ai.HealthRequest, opts ...grpc.CallOption) (*ai.HealthResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*ai.HealthResponse), args.Error(1)
}

// TestGRPCAIProvider 测试gRPC AI提供商
func TestGRPCAIProvider(t *testing.T) {
	
	t.Run("NewGRPCAIProvider创建实例", func(t *testing.T) {
		// Given
		serverAddr := "localhost:50051"
		name := "test-grpc-provider"
		
		// When
		provider := NewGRPCAIProvider(serverAddr, name)
		
		// Then
		assert.Equal(t, serverAddr, provider.serverAddr)
		assert.Equal(t, name, provider.name)
		assert.False(t, provider.connected)
		assert.Nil(t, provider.client)
		assert.Nil(t, provider.conn)
	})

	t.Run("GenerateContent成功场景", func(t *testing.T) {
		// Given
		mockClient := &MockAIServiceClient{}
		provider := &GRPCAIProvider{
			client:     mockClient,
			serverAddr: "localhost:50051",
			name:       "test-provider",
			connected:  true,
		}

		expectedResponse := &ai.ContentResponse{
			Content:        "这是生成的内容",
			Summary:        func() *string { v := "内容摘要"; return &v }(),
			Tags:          []string{"测试", "AI"},
			ModelUsed:     "test-model",
			TokensUsed:    25,
			GenerationTime: 0.8,
			Confidence:    func() *float32 { v := float32(0.92); return &v }(),
			Sources:       []string{"test_source"},
		}

		mockClient.On("GenerateContent", mock.Anything, mock.AnythingOfType("*ai.ContentRequest"), mock.Anything).Return(expectedResponse, nil)

		// When
		ctx := context.Background()
		content, err := provider.GenerateContent(ctx, "测试提示")

		// Then
		require.NoError(t, err)
		assert.Equal(t, expectedResponse.Content, content)
		mockClient.AssertExpectations(t)
		
		// 验证请求参数
		calls := mockClient.Calls
		require.Len(t, calls, 1)
		request := calls[0].Arguments[1].(*ai.ContentRequest)
		assert.Equal(t, "测试提示", request.Prompt)
		assert.Equal(t, int32(500), *request.MaxLength)
	})

	t.Run("GenerateContent未连接错误", func(t *testing.T) {
		// Given
		provider := &GRPCAIProvider{
			serverAddr: "localhost:50051",
			name:       "test-provider",
			connected:  false,
		}

		// When
		ctx := context.Background()
		content, err := provider.GenerateContent(ctx, "测试提示")

		// Then
		assert.Empty(t, content)
		require.Error(t, err)
		
		var aiErr *AIError
		assert.ErrorAs(t, err, &aiErr)
		assert.Equal(t, "test-provider", aiErr.Provider)
		assert.Equal(t, "not_connected", aiErr.Code)
		assert.True(t, aiErr.Retryable)
	})

	t.Run("GenerateContent gRPC错误处理", func(t *testing.T) {
		// Given
		mockClient := &MockAIServiceClient{}
		provider := &GRPCAIProvider{
			client:     mockClient,
			serverAddr: "localhost:50051", 
			name:       "test-provider",
			connected:  true,
		}

		grpcError := status.Error(codes.Unavailable, "服务不可用")
		mockClient.On("GenerateContent", mock.Anything, mock.AnythingOfType("*ai.ContentRequest"), mock.Anything).Return((*ai.ContentResponse)(nil), grpcError)

		// When
		ctx := context.Background()
		content, err := provider.GenerateContent(ctx, "测试提示")

		// Then
		assert.Empty(t, content)
		require.Error(t, err)
		
		var aiErr *AIError
		assert.ErrorAs(t, err, &aiErr)
		assert.Equal(t, "test-provider", aiErr.Provider)
		assert.Equal(t, "generation_failed", aiErr.Code)
		assert.Contains(t, aiErr.Message, "服务不可用")
		assert.True(t, aiErr.Retryable)
	})

	t.Run("GenerateArticle成功场景", func(t *testing.T) {
		// Given
		mockClient := &MockAIServiceClient{}
		provider := &GRPCAIProvider{
			client:     mockClient,
			connected:  true,
			name:       "test-provider",
		}

		expectedResponse := &ai.ArticleResponse{
			Title:          "测试文章标题",
			Content:        "这是文章内容",
			Summary:        "文章摘要",
			Tags:          []string{"测试", "文章"},
			Category:      func() *string { v := "技术"; return &v }(),
			WordCount:     100,
			ModelUsed:     "article-model",
			TokensUsed:    50,
			GenerationTime: 1.5,
			Sources:       []string{"参考资料"},
		}

		mockClient.On("GenerateArticle", mock.Anything, mock.AnythingOfType("*ai.ArticleRequest"), mock.Anything).Return(expectedResponse, nil)

		// When
		ctx := context.Background()
		article, err := provider.GenerateArticle(ctx, "测试标题", "技术")

		// Then
		require.NoError(t, err)
		assert.Equal(t, expectedResponse.Title, article.Title)
		assert.Equal(t, expectedResponse.Content, article.Content)
		assert.Equal(t, expectedResponse.Summary, article.Summary)
		assert.Equal(t, expectedResponse.Tags, article.Tags)
		assert.Equal(t, "test-provider", article.GeneratedBy)

		mockClient.AssertExpectations(t)
	})

	t.Run("HealthCheck成功场景", func(t *testing.T) {
		// Given
		mockClient := &MockAIServiceClient{}
		provider := &GRPCAIProvider{
			client:     mockClient,
			connected:  false, // 测试健康检查会更新连接状态
			name:       "test-provider",
		}

		expectedResponse := &ai.HealthResponse{
			Status:      ai.HealthStatus_SERVING,
			ServiceName: "test-service",
			Version:     "1.0.0",
			Message:     func() *string { v := "服务正常"; return &v }(),
		}

		mockClient.On("HealthCheck", mock.Anything, mock.AnythingOfType("*ai.HealthRequest"), mock.Anything).Return(expectedResponse, nil)

		// When
		ctx := context.Background()
		err := provider.healthCheck(ctx)

		// Then
		require.NoError(t, err)
		mockClient.AssertExpectations(t)
		
		// 验证请求参数
		calls := mockClient.Calls
		require.Len(t, calls, 1)
		request := calls[0].Arguments[1].(*ai.HealthRequest)
		assert.NotNil(t, request)
	})

	t.Run("HealthCheck服务不可用", func(t *testing.T) {
		// Given
		mockClient := &MockAIServiceClient{}
		provider := &GRPCAIProvider{
			client: mockClient,
			name:   "test-provider",
		}

		unhealthyResponse := &ai.HealthResponse{
			Status:  ai.HealthStatus_NOT_SERVING,
			Message: func() *string { v := "服务异常"; return &v }(),
		}

		mockClient.On("HealthCheck", mock.Anything, mock.AnythingOfType("*ai.HealthRequest"), mock.Anything).Return(unhealthyResponse, nil)

		// When
		ctx := context.Background()
		err := provider.healthCheck(ctx)

		// Then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "服务状态异常")
		assert.Contains(t, err.Error(), "服务异常")
	})
}

// TestGRPCAIProviderTimeout 测试超时处理
func TestGRPCAIProviderTimeout(t *testing.T) {
	t.Run("GenerateContent超时", func(t *testing.T) {
		// Given
		mockClient := &MockAIServiceClient{}
		provider := &GRPCAIProvider{
			client:    mockClient,
			connected: true,
			name:      "test-provider",
		}

		// 模拟超时
		mockClient.On("GenerateContent", mock.Anything, mock.Anything, mock.Anything).Return(
			(*ai.ContentResponse)(nil), 
			status.Error(codes.DeadlineExceeded, "context deadline exceeded"),
		)

		// When
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		
		content, err := provider.GenerateContent(ctx, "测试提示")

		// Then
		assert.Empty(t, content)
		require.Error(t, err)
		
		var aiErr *AIError
		assert.ErrorAs(t, err, &aiErr)
		assert.Equal(t, "generation_failed", aiErr.Code)
		assert.True(t, aiErr.Retryable)
	})
}