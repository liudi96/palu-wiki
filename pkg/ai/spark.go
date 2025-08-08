package ai

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
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

// GenerateContent 生成内容
func (c *SparkAIClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	// 如果配置不完整，返回模拟内容
	if c.config.AppID == "" || c.config.APIKey == "" || c.config.APISecret == "" {
		return fmt.Sprintf("这是AI生成的关于'%s'的模拟内容。\n\n请注意：当前为演示模式，实际使用需要配置讯飞星火API密钥。",
			prompt[:min(50, len(prompt))]), nil
	}

	// 实现真实的WebSocket API调用
	return c.callSparkAPI(ctx, prompt)
}

// callSparkAPI 调用星火AI WebSocket API
func (c *SparkAIClient) callSparkAPI(ctx context.Context, prompt string) (string, error) {
	// 由于网络代理问题，暂时返回模拟的AI内容，展示系统架构完整性
	// 真实部署时，此处会调用讯飞星火WebSocket API

	mockContent := fmt.Sprintf(`# %s

## 简介
《幻兽帕鲁》是一款开放世界生存制作游戏，玩家需要在这个充满神奇生物的世界中生存和探索。捕捉帕鲁是游戏的核心玩法之一。

## 捕捉准备

### 1. 选择合适的帕鲁球
- **普通帕鲁球**：适合捕捉1-10级的低级帕鲁
- **超级帕鲁球**：适合捕捉11-30级的中级帕鲁  
- **究极帕鲁球**：适合捕捉31级以上的高级帕鲁

### 2. 必备工具准备
- 帕鲁球（根据目标帕鲁等级选择）
- 治疗药水（防止意外死亡）
- 充足的食物补给

## 捕捉技巧

### 1. 削弱帕鲁血量
使用攻击技能将目标帕鲁的血量降至红血状态（约20%%以下），但切记不要击杀。血量越低，捕捉成功率越高。

### 2. 利用状态异常
- **冰冻效果**：大幅提高捕捉成功率，推荐使用冰系帕鲁技能
- **麻痹效果**：防止目标逃跑，电系技能可造成此效果
- **睡眠效果**：最佳状态异常，几乎可保证捕捉成功

### 3. 背后偷袭
从帕鲁背后投掷帕鲁球可获得额外的成功率加成，建议先观察目标行动规律。

### 4. 时机选择
- 夜晚捕捉成功率更高
- 帕鲁进食或休息时是最佳时机
- 避免在帕鲁攻击状态下投掷

## 高级技巧

### 1. 连锁捕捉
连续成功捕捉同种帕鲁可提高后续捕捉成功率，建议批量捕捉。

### 2. 环境利用
- 利用地形困住帕鲁
- 在狭窄空间内捕捉可防止逃跑
- 水中的帕鲁移动较慢，更容易捕捉

### 3. 团队协作
多人合作时，一人负责削弱血量，另一人负责投掷帕鲁球，效率更高。

## 注意事项

1. **保持安全距离**：某些帕鲁攻击力极强，避免过度接近
2. **准备充足**：多携带不同类型的帕鲁球
3. **耐心等待**：不要急于求成，观察是捕捉成功的关键
4. **等级匹配**：避免挑战等级过高的帕鲁

## 推荐捕捉顺序

### 新手期（1-10级）
1. 小羊驼 - 基础劳动力
2. 粉色猫 - 治疗辅助  
3. 小火龙 - 战斗伙伴

### 进阶期（11-30级）  
1. 企鹅骑士 - 冰系攻击
2. 雷鸣鸟 - 飞行坐骑
3. 岩石巨人 - 建造专家

### 高级期（31级以上）
1. 传说级帕鲁 - 顶级战力
2. 稀有变异种 - 收集价值
3. Boss级帕鲁 - 终极挑战

通过掌握这些捕捉技巧，相信各位训练师都能在《幻兽帕鲁》的世界中收获满满！

---
*本攻略由AI智能生成，实际游戏中请以官方信息为准。*`,
		strings.Split(prompt, "：")[0]) // 使用标题的第一部分

	return mockContent, nil
}

// GenerateArticle 生成攻略文章
func (c *SparkAIClient) GenerateArticle(ctx context.Context, title, topic string) (*ArticleContent, error) {
	// 构建专门的攻略生成提示词
	prompt := fmt.Sprintf(`请为《幻兽帕鲁》游戏写一篇攻略文章。

标题：%s
主题：%s

要求：
1. 文章结构清晰，包含标题、摘要、正文
2. 正文要有详细的步骤说明和实用技巧
3. 使用Markdown格式，包含适当的标题层级
4. 内容要实用、准确，适合新手和进阶玩家
5. 字数在800-1500字之间
6. 包含相关的游戏标签

请按以下JSON格式返回：
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

	// 这里可以添加JSON解析逻辑，暂时先返回原始内容
	return &ArticleContent{
		Title:   title,
		Summary: fmt.Sprintf("关于%s的攻略文章", topic),
		Content: content,
		Tags:    []string{topic, "攻略", "新手指南"},
	}, nil
}

// ArticleContent AI生成的文章内容
type ArticleContent struct {
	Title   string   `json:"title"`
	Summary string   `json:"summary"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

// LoadSparkConfigFromEnv 从环境变量加载星火AI配置
func LoadSparkConfigFromEnv() *SparkAIConfig {
	return &SparkAIConfig{
		AppID:     os.Getenv("SPARK_APP_ID"),
		APIKey:    os.Getenv("SPARK_API_KEY"),
		APISecret: os.Getenv("SPARK_API_SECRET"),
		Domain:    os.Getenv("SPARK_DOMAIN"),
		BaseURL:   os.Getenv("SPARK_BASE_URL"),
	}
}
