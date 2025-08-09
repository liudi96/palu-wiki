package ai

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// SparkAIConfig 星火AI配置
type SparkAIConfig struct {
	AppID     string
	APIKey    string
	APISecret string
	Domain    string
	BaseURL   string
}

// SparkAIClient 星火AI客户端
type SparkAIClient struct {
	config *SparkAIConfig
}

// 确保SparkAIClient实现AIProvider接口
var _ AIProvider = (*SparkAIClient)(nil)

// NewSparkAIClient 创建星火AI客户端
func NewSparkAIClient(config *SparkAIConfig) (*SparkAIClient, error) {
	if config.AppID == "" || config.APIKey == "" || config.APISecret == "" {
		return nil, fmt.Errorf("星火AI配置不完整")
	}

	// 设置默认配置
	if config.Domain == "" {
		config.Domain = "lite" // 使用免费的Spark Lite版本
	}
	if config.BaseURL == "" {
		config.BaseURL = "wss://spark-api.xf-yun.com/v1.1/chat" // Spark Lite的WebSocket地址
	}

	return &SparkAIClient{
		config: config,
	}, nil
}

// SparkMessage 星火消息结构
type SparkMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// SparkRequest 星火请求结构
type SparkRequest struct {
	Header struct {
		AppID string `json:"app_id"`
		UID   string `json:"uid,omitempty"`
	} `json:"header"`
	Parameter struct {
		Chat struct {
			Domain      string  `json:"domain"`
			Temperature float64 `json:"temperature,omitempty"`
			MaxTokens   int     `json:"max_tokens,omitempty"`
		} `json:"chat"`
	} `json:"parameter"`
	Payload struct {
		Message struct {
			Text []SparkMessage `json:"text"`
		} `json:"message"`
	} `json:"payload"`
}

// SparkResponse 星火响应结构
type SparkResponse struct {
	Header struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Sid     string `json:"sid"`
		Status  int    `json:"status"`
	} `json:"header"`
	Payload struct {
		Choices struct {
			Status int `json:"status"`
			Seq    int `json:"seq"`
			Text   []struct {
				Content string `json:"content"`
				Role    string `json:"role"`
				Index   int    `json:"index"`
			} `json:"text"`
		} `json:"choices"`
		Usage struct {
			Text struct {
				QuestionTokens   int `json:"question_tokens"`
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"text"`
		} `json:"usage"`
	} `json:"payload"`
}

// generateAuthURL 生成认证URL
func (c *SparkAIClient) generateAuthURL() (string, error) {
	u, err := url.Parse(c.config.BaseURL)
	if err != nil {
		return "", err
	}

	// RFC1123格式的时间戳
	now := time.Now().UTC()
	date := now.Format("Mon, 02 Jan 2006 15:04:05 GMT")

	// 构建签名字符串
	signatureOrigin := fmt.Sprintf("host: %s\ndate: %s\nGET %s HTTP/1.1", u.Host, date, u.Path)

	// HMAC SHA256签名
	h := hmac.New(sha256.New, []byte(c.config.APISecret))
	h.Write([]byte(signatureOrigin))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// 构建Authorization header
	authorizationOrigin := fmt.Sprintf(`api_key="%s", algorithm="hmac-sha256", headers="host date request-line", signature="%s"`,
		c.config.APIKey, signature)
	authorization := base64.StdEncoding.EncodeToString([]byte(authorizationOrigin))

	// 构建完整的WebSocket URL
	params := url.Values{}
	params.Add("authorization", authorization)
	params.Add("date", date)
	params.Add("host", u.Host)

	u.RawQuery = params.Encode()
	return u.String(), nil
}

// min 返回两个数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maskString 掩盖敏感字符串
func maskString(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

// GenerateContent 生成内容
func (c *SparkAIClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	// 检查配置完整性，不允许降级
	if c.config.AppID == "" || c.config.APIKey == "" || c.config.APISecret == "" {
		return "", fmt.Errorf("讯飞星火AI配置不完整：AppID=%s, APIKey=%s, APISecret=%s",
			c.config.AppID, maskString(c.config.APIKey), maskString(c.config.APISecret))
	}

	// 直接调用真实API，不使用任何降级逻辑
	return c.callSparkAPI(ctx, prompt)
}

// callSparkAPI 调用星火AI WebSocket API
func (c *SparkAIClient) callSparkAPI(ctx context.Context, prompt string) (string, error) {
	// 生成认证URL
	authURL, err := c.generateAuthURL()
	if err != nil {
		return "", fmt.Errorf("生成认证URL失败: %v", err)
	}

	// 建立WebSocket连接（启用压缩与超时）
	dialer := websocket.Dialer{
		HandshakeTimeout:  15 * time.Second,
		EnableCompression: true,
		ReadBufferSize:    4096,
		WriteBufferSize:   4096,
	}

	conn, _, err := dialer.DialContext(ctx, authURL, nil)
	if err != nil {
		return "", fmt.Errorf("WebSocket连接失败: %v", err)
	}
	defer conn.Close()

	// 构建请求消息
	request := SparkRequest{
		Header: struct {
			AppID string `json:"app_id"`
			UID   string `json:"uid,omitempty"`
		}{
			AppID: c.config.AppID,
			UID:   "palu-wiki-user",
		},
		Parameter: struct {
			Chat struct {
				Domain      string  `json:"domain"`
				Temperature float64 `json:"temperature,omitempty"`
				MaxTokens   int     `json:"max_tokens,omitempty"`
			} `json:"chat"`
		}{
			Chat: struct {
				Domain      string  `json:"domain"`
				Temperature float64 `json:"temperature,omitempty"`
				MaxTokens   int     `json:"max_tokens,omitempty"`
			}{
				Domain:      c.config.Domain,
				Temperature: 0.7,
				MaxTokens:   4096,
			},
		},
		Payload: struct {
			Message struct {
				Text []SparkMessage `json:"text"`
			} `json:"message"`
		}{
			Message: struct {
				Text []SparkMessage `json:"text"`
			}{
				Text: []SparkMessage{
					{Role: "system", Content: "你是一名资深游戏攻略写手，需输出结构化、详实、可执行的中文Markdown内容。"},
					{Role: "user", Content: prompt},
				},
			},
		},
	}

	// 发送请求
	if err := conn.WriteJSON(request); err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}

	// 读取流式响应，直到状态为2
	var fullContent strings.Builder
	// 设置读超时，避免无限等待
	_ = conn.SetReadDeadline(time.Now().Add(180 * time.Second))

	for {
		var response SparkResponse
		if err := conn.ReadJSON(&response); err != nil {
			return "", fmt.Errorf("读取响应失败: %v", err)
		}

		if response.Header.Code != 0 {
			return "", fmt.Errorf("API错误: %s (代码: %d)", response.Header.Message, response.Header.Code)
		}

		for _, piece := range response.Payload.Choices.Text {
			fullContent.WriteString(piece.Content)
		}

		if response.Payload.Choices.Status == 2 || response.Header.Status == 2 {
			break
		}
	}

	result := fullContent.String()
	if result == "" {
		return "", fmt.Errorf("星火AI返回空内容，请检查API配置与Domain/BaseURL是否匹配")
	}

	return result, nil
}

// GenerateArticle 生成攻略文章
func (c *SparkAIClient) GenerateArticle(ctx context.Context, title, topic string) (*ArticleContent, error) {
	// 构建专门的攻略生成提示词
	prompt := fmt.Sprintf(`请为《幻兽帕鲁》游戏写一篇高质量攻略文章。

标题：%s
主题：%s

写作要求（务必严格遵循）：
1. 采用Markdown格式，包含 H1 标题、目录、分级小节、列表与表格（如有必要）
2. 正文需详尽且可执行，包含步骤、注意事项、技巧与常见错误
3. 先给出100-200字的摘要
4. 文章长度不少于1200字
5. 内容真实、准确、无编造；适合新手与进阶玩家
6. 用词客观简洁，避免废话

仅输出以下JSON格式的数据，不要包含任何markdown代码块标记，不要任何额外说明文字：
{
  "title": "文章标题",
  "summary": "文章摘要（100-200字）",
  "content": "文章正文（Markdown格式）",
  "tags": ["标签1", "标签2", "标签3"]
}`, title, topic)

	content, err := c.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 清理内容，移除可能的markdown代码块标记
	cleanedContent := content
	// 移除开头的```json或```
	cleanedContent = strings.TrimPrefix(cleanedContent, "```json")
	cleanedContent = strings.TrimPrefix(cleanedContent, "```")
	// 移除结尾的```
	cleanedContent = strings.TrimSuffix(cleanedContent, "```")
	cleanedContent = strings.TrimSpace(cleanedContent)

	// 优先尝试解析为JSON结构
	var parsed ArticleContent
	if err := json.Unmarshal([]byte(cleanedContent), &parsed); err == nil && parsed.Content != "" {
		if parsed.Title == "" {
			parsed.Title = title
		}
		if len(parsed.Tags) == 0 {
			parsed.Tags = []string{topic, "攻略", "新手指南"}
		}
		return &parsed, nil
	}

	// 回退：返回原始内容
	return &ArticleContent{
		Title:   title,
		Summary: fmt.Sprintf("关于%s的攻略文章", topic),
		Content: content,
		Tags:    []string{topic, "攻略", "新手指南"},
	}, nil
}

// LoadSparkConfigFromEnv 从环境变量加载星火AI配置
// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func LoadSparkConfigFromEnv() *SparkAIConfig {
	return &SparkAIConfig{
		AppID:     os.Getenv("SPARK_APP_ID"),
		APIKey:    os.Getenv("SPARK_API_KEY"),
		APISecret: os.Getenv("SPARK_API_SECRET"),
		Domain:    os.Getenv("SPARK_DOMAIN"),
		BaseURL:   os.Getenv("SPARK_BASE_URL"),
	}
}

// GetProviderName 获取提供商名称
func (c *SparkAIClient) GetProviderName() string {
	return "spark"
}

// IsAvailable 检查提供商是否可用
func (c *SparkAIClient) IsAvailable(ctx context.Context) bool {
	// 简单检查配置是否完整
	return c.config != nil && 
		   c.config.AppID != "" && 
		   c.config.APIKey != "" && 
		   c.config.APISecret != ""
}

// GetQuotaInfo 获取配额信息
func (c *SparkAIClient) GetQuotaInfo(ctx context.Context) (*QuotaInfo, error) {
	// 星火API通常不提供配额查询接口，返回默认信息
	return &QuotaInfo{
		Provider:       "spark",
		TotalTokens:    -1, // -1 表示未知
		UsedTokens:     -1,
		RemainingQuota: 1.0, // 假设可用
	}, nil
}
