package ai

import (
	"context"
	"time"
)

// AIProvider 定义AI提供商接口
type AIProvider interface {
	// GenerateContent 生成简单内容
	GenerateContent(ctx context.Context, prompt string) (string, error)
	
	// GenerateArticle 生成文章内容
	GenerateArticle(ctx context.Context, title, topic string) (*ArticleContent, error)
	
	// GetProviderName 获取提供商名称
	GetProviderName() string
	
	// IsAvailable 检查提供商是否可用
	IsAvailable(ctx context.Context) bool
	
	// GetQuotaInfo 获取配额信息
	GetQuotaInfo(ctx context.Context) (*QuotaInfo, error)
}

// ArticleContent 统一的文章内容结构
type ArticleContent struct {
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Summary     string    `json:"summary"`
	Tags        []string  `json:"tags"`
	GeneratedBy string    `json:"generated_by"`
	GeneratedAt time.Time `json:"generated_at"`
}

// QuotaInfo 配额信息
type QuotaInfo struct {
	Provider       string  `json:"provider"`
	TotalTokens    int64   `json:"total_tokens"`
	UsedTokens     int64   `json:"used_tokens"`
	RemainingQuota float64 `json:"remaining_quota"`
	ResetTime      time.Time `json:"reset_time,omitempty"`
}

// AIError AI调用错误
type AIError struct {
	Provider string
	Code     string
	Message  string
	Retryable bool
}

func (e *AIError) Error() string {
	return e.Message
}

// IsQuotaExceeded 判断是否配额超限
func (e *AIError) IsQuotaExceeded() bool {
	return e.Code == "quota_exceeded" || e.Code == "rate_limit"
}