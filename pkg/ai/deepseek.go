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

// DeepSeekConfig DeepSeek配置
type DeepSeekConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
}

// DeepSeekClient DeepSeek客户端
type DeepSeekClient struct {
	config *DeepSeekConfig
	client *http.Client
}

// 确保实现AIProvider接口
var _ AIProvider = (*DeepSeekClient)(nil)

// DeepSeekMessage 消息格式
type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekRequest 请求格式
type DeepSeekRequest struct {
	Model       string            `json:"model"`
	Messages    []DeepSeekMessage `json:"messages"`
	Temperature float32           `json:"temperature,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Stream      bool              `json:"stream"`
}

// DeepSeekResponse 响应格式
type DeepSeekResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewDeepSeekClient 创建DeepSeek客户端
func NewDeepSeekClient(config *DeepSeekConfig) *DeepSeekClient {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.deepseek.com"
	}
	if config.Model == "" {
		config.Model = "deepseek-chat"
	}

	return &DeepSeekClient{
		config: config,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// GetProviderName 获取提供商名称
func (c *DeepSeekClient) GetProviderName() string {
	return "deepseek"
}

// IsAvailable 检查提供商是否可用
func (c *DeepSeekClient) IsAvailable(ctx context.Context) bool {
	return c.config != nil && c.config.APIKey != ""
}

// GetQuotaInfo 获取配额信息
func (c *DeepSeekClient) GetQuotaInfo(ctx context.Context) (*QuotaInfo, error) {
	return &QuotaInfo{
		Provider:       "deepseek",
		TotalTokens:    -1,
		UsedTokens:     -1,
		RemainingQuota: 1.0,
	}, nil
}

// callDeepSeekAPI 调用DeepSeek API
func (c *DeepSeekClient) callDeepSeekAPI(ctx context.Context, messages []DeepSeekMessage) (string, error) {
	req := DeepSeekRequest{
		Model:       c.config.Model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   4000,
		Stream:      false,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", &AIError{
			Provider:  "deepseek",
			Code:     "marshal_error",
			Message:  fmt.Sprintf("序列化请求失败: %v", err),
			Retryable: false,
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", &AIError{
			Provider:  "deepseek",
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
			Provider:  "deepseek",
			Code:     "network_error",
			Message:  fmt.Sprintf("网络请求失败: %v", err),
			Retryable: true,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &AIError{
			Provider:  "deepseek",
			Code:     "read_error",
			Message:  fmt.Sprintf("读取响应失败: %v", err),
			Retryable: true,
		}
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 429 {
			return "", &AIError{
				Provider:  "deepseek",
				Code:     "rate_limit",
				Message:  "请求频率超限",
				Retryable: true,
			}
		}
		return "", &AIError{
			Provider:  "deepseek",
			Code:     "api_error",
			Message:  fmt.Sprintf("API错误 (状态码: %d): %s", resp.StatusCode, string(body)),
			Retryable: resp.StatusCode >= 500,
		}
	}

	var apiResp DeepSeekResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", &AIError{
			Provider:  "deepseek",
			Code:     "unmarshal_error",
			Message:  fmt.Sprintf("解析响应失败: %v", err),
			Retryable: false,
		}
	}

	if len(apiResp.Choices) == 0 {
		return "", &AIError{
			Provider:  "deepseek",
			Code:     "empty_response",
			Message:  "API返回空响应",
			Retryable: false,
		}
	}

	return apiResp.Choices[0].Message.Content, nil
}

// GenerateContent 生成内容
func (c *DeepSeekClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	messages := []DeepSeekMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	return c.callDeepSeekAPI(ctx, messages)
}

// GenerateArticle 生成文章
func (c *DeepSeekClient) GenerateArticle(ctx context.Context, title, topic string) (*ArticleContent, error) {
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
func (c *DeepSeekClient) generateSummary(content string) string {
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
func (c *DeepSeekClient) extractTags(topic string) []string {
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

// LoadDeepSeekConfigFromEnv 从环境变量加载配置
func LoadDeepSeekConfigFromEnv() *DeepSeekConfig {
	return &DeepSeekConfig{
		APIKey:  getEnv("DEEPSEEK_API_KEY", ""),
		BaseURL: getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
		Model:   getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
	}
}