package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// QwenConfig 通义千问配置
type QwenConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
}

// QwenClient 通义千问客户端
type QwenClient struct {
	config *QwenConfig
	client *http.Client
}

// 确保实现AIProvider接口
var _ AIProvider = (*QwenClient)(nil)

// QwenMessage 消息格式
type QwenMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// QwenRequest 请求格式
type QwenRequest struct {
	Model       string        `json:"model"`
	Messages    []QwenMessage `json:"messages"`
	Temperature float32       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream"`
}

// QwenResponse 响应格式
type QwenResponse struct {
	Output struct {
		Text         string `json:"text"`
		FinishReason string `json:"finish_reason"`
	} `json:"output"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
	RequestID string `json:"request_id"`
}

// NewQwenClient 创建通义千问客户端
func NewQwenClient(config *QwenConfig) *QwenClient {
	if config.BaseURL == "" {
		config.BaseURL = "https://dashscope.aliyuncs.com"
	}
	if config.Model == "" {
		config.Model = "qwen-turbo"
	}

	return &QwenClient{
		config: config,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// GetProviderName 获取提供商名称
func (c *QwenClient) GetProviderName() string {
	return "qwen"
}

// IsAvailable 检查提供商是否可用
func (c *QwenClient) IsAvailable(ctx context.Context) bool {
	return c.config != nil && c.config.APIKey != ""
}

// GetQuotaInfo 获取配额信息
func (c *QwenClient) GetQuotaInfo(ctx context.Context) (*QuotaInfo, error) {
	return &QuotaInfo{
		Provider:       "qwen",
		TotalTokens:    -1,
		UsedTokens:     -1,
		RemainingQuota: 1.0,
	}, nil
}

// callQwenAPI 调用通义千问API
func (c *QwenClient) callQwenAPI(ctx context.Context, messages []QwenMessage) (string, error) {
	req := QwenRequest{
		Model:       c.config.Model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   4000,
		Stream:      false,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", &AIError{
			Provider:  "qwen",
			Code:     "marshal_error",
			Message:  fmt.Sprintf("序列化请求失败: %v", err),
			Retryable: false,
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/api/v1/services/aigc/text-generation/generation", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", &AIError{
			Provider:  "qwen",
			Code:     "request_error",
			Message:  fmt.Sprintf("创建请求失败: %v", err),
			Retryable: false,
		}
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", &AIError{
			Provider:  "qwen",
			Code:     "network_error",
			Message:  fmt.Sprintf("网络请求失败: %v", err),
			Retryable: true,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &AIError{
			Provider:  "qwen",
			Code:     "read_error",
			Message:  fmt.Sprintf("读取响应失败: %v", err),
			Retryable: true,
		}
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 429 {
			return "", &AIError{
				Provider:  "qwen",
				Code:     "rate_limit",
				Message:  "请求频率超限",
				Retryable: true,
			}
		}
		return "", &AIError{
			Provider:  "qwen",
			Code:     "api_error",
			Message:  fmt.Sprintf("API错误 (状态码: %d): %s", resp.StatusCode, string(body)),
			Retryable: resp.StatusCode >= 500,
		}
	}

	var apiResp QwenResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", &AIError{
			Provider:  "qwen",
			Code:     "unmarshal_error",
			Message:  fmt.Sprintf("解析响应失败: %v", err),
			Retryable: false,
		}
	}

	if apiResp.Output.Text == "" {
		return "", &AIError{
			Provider:  "qwen",
			Code:     "empty_response",
			Message:  "API返回空响应",
			Retryable: false,
		}
	}

	return apiResp.Output.Text, nil
}

// GenerateContent 生成内容
func (c *QwenClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	messages := []QwenMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	return c.callQwenAPI(ctx, messages)
}

// GenerateArticle 生成文章
func (c *QwenClient) GenerateArticle(ctx context.Context, title, topic string) (*ArticleContent, error) {
	prompt := fmt.Sprintf(`请根据以下信息生成一篇高质量的技术文章:

标题: %s
主题: %s

要求:
1. 文章结构清晰，包含引言、主体内容、总结
2. 内容专业且易懂，适合技术博客
3. 字数控制在800-1500字
4. 使用Markdown格式
5. 包含适当的代码示例（如适用）
6. 最后提供3-5个相关标签

请按以下JSON格式返回:
{
  "title": "文章标题",
  "content": "完整的文章内容（Markdown格式）",
  "summary": "文章摘要（100字以内）",
  "tags": ["标签1", "标签2", "标签3"]
}`, title, topic)

	response, err := c.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 尝试解析JSON格式的响应
	var articleData struct {
		Title   string   `json:"title"`
		Content string   `json:"content"`
		Summary string   `json:"summary"`
		Tags    []string `json:"tags"`
	}

	if err := json.Unmarshal([]byte(response), &articleData); err != nil {
		// 如果不是JSON格式，则作为纯文本处理
		return &ArticleContent{
			Title:   title,
			Content: response,
			Summary: c.generateSummary(response),
			Tags:    c.extractTags(topic),
		}, nil
	}

	return &ArticleContent{
		Title:   articleData.Title,
		Content: articleData.Content,
		Summary: articleData.Summary,
		Tags:    articleData.Tags,
	}, nil
}

// generateSummary 生成摘要
func (c *QwenClient) generateSummary(content string) string {
	if len(content) <= 100 {
		return content
	}
	
	// 取前100个字符作为摘要
	runes := []rune(content)
	if len(runes) > 100 {
		return string(runes[:97]) + "..."
	}
	return string(runes)
}

// extractTags 提取标签
func (c *QwenClient) extractTags(topic string) []string {
	// 简单的标签提取逻辑
	words := strings.Fields(strings.ToLower(topic))
	tags := make([]string, 0, len(words))
	
	for _, word := range words {
		if len(word) > 2 && len(tags) < 5 {
			tags = append(tags, word)
		}
	}
	
	if len(tags) == 0 {
		tags = append(tags, "技术")
	}
	
	return tags
}

// LoadQwenConfigFromEnv 从环境变量加载配置
func LoadQwenConfigFromEnv() *QwenConfig {
	return &QwenConfig{
		APIKey:  getEnv("QWEN_API_KEY", ""),
		BaseURL: getEnv("QWEN_BASE_URL", "https://dashscope.aliyuncs.com"),
		Model:   getEnv("QWEN_MODEL", "qwen-turbo"),
	}
}