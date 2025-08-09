package ai

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// ProviderStrategy 提供商选择策略
type ProviderStrategy string

const (
	StrategyPriority   ProviderStrategy = "priority"   // 按优先级选择
	StrategyRoundRobin ProviderStrategy = "roundrobin" // 轮询选择
	StrategyRandom     ProviderStrategy = "random"     // 随机选择
)

// ProviderConfig 提供商配置
type ProviderConfig struct {
	Name     string `json:"name"`
	Priority int    `json:"priority"` // 数字越小优先级越高
	Weight   int    `json:"weight"`   // 权重，用于负载均衡
	Enabled  bool   `json:"enabled"`
}

// AIManager AI管理器，负责多模型调度和降级
type AIManager struct {
	providers map[string]AIProvider
	configs   []ProviderConfig
	strategy  ProviderStrategy
	
	// 断路器状态
	circuitBreakers map[string]*CircuitBreaker
	
	// 统计信息
	stats map[string]*ProviderStats
	
	mu sync.RWMutex
}

// CircuitBreaker 断路器
type CircuitBreaker struct {
	FailureThreshold int
	RecoveryTimeout  time.Duration
	
	failures    int
	lastFailure time.Time
	state       string // "closed", "open", "half-open"
	mu          sync.RWMutex
}

// ProviderStats 提供商统计
type ProviderStats struct {
	TotalRequests int64
	SuccessCount  int64
	FailureCount  int64
	AvgLatency    time.Duration
	LastUsed      time.Time
}

// NewAIManager 创建AI管理器
func NewAIManager(strategy ProviderStrategy) *AIManager {
	return &AIManager{
		providers:       make(map[string]AIProvider),
		configs:         make([]ProviderConfig, 0),
		strategy:        strategy,
		circuitBreakers: make(map[string]*CircuitBreaker),
		stats:           make(map[string]*ProviderStats),
	}
}

// RegisterProvider 注册AI提供商
func (m *AIManager) RegisterProvider(provider AIProvider, config ProviderConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	name := config.Name
	m.providers[name] = provider
	m.configs = append(m.configs, config)
	
	// 初始化断路器
	m.circuitBreakers[name] = &CircuitBreaker{
		FailureThreshold: 5,
		RecoveryTimeout:  30 * time.Second,
		state:           "closed",
	}
	
	// 初始化统计
	m.stats[name] = &ProviderStats{}
	
	log.Printf("注册AI提供商: %s (优先级: %d)", name, config.Priority)
}

// selectProvider 选择可用的提供商
func (m *AIManager) selectProvider(ctx context.Context) (AIProvider, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// 按优先级排序可用提供商
	availableProviders := m.getAvailableProviders(ctx)
	if len(availableProviders) == 0 {
		return nil, "", fmt.Errorf("没有可用的AI提供商")
	}
	
	switch m.strategy {
	case StrategyPriority:
		return m.selectByPriority(availableProviders)
	case StrategyRoundRobin:
		return m.selectByRoundRobin(availableProviders)
	default:
		return m.selectByPriority(availableProviders)
	}
}

// getAvailableProviders 获取可用提供商列表
func (m *AIManager) getAvailableProviders(ctx context.Context) []ProviderConfig {
	var available []ProviderConfig
	
	for _, config := range m.configs {
		if !config.Enabled {
			continue
		}
		
		// 检查断路器状态
		if !m.isCircuitClosed(config.Name) {
			continue
		}
		
		// 检查提供商是否可用
		if provider, exists := m.providers[config.Name]; exists {
			if provider.IsAvailable(ctx) {
				available = append(available, config)
			}
		}
	}
	
	return available
}

// selectByPriority 按优先级选择
func (m *AIManager) selectByPriority(providers []ProviderConfig) (AIProvider, string, error) {
	if len(providers) == 0 {
		return nil, "", fmt.Errorf("没有可用提供商")
	}
	
	// 找到优先级最高的（数字最小）
	bestConfig := providers[0]
	for _, config := range providers[1:] {
		if config.Priority < bestConfig.Priority {
			bestConfig = config
		}
	}
	
	provider := m.providers[bestConfig.Name]
	return provider, bestConfig.Name, nil
}

// selectByRoundRobin 轮询选择（简化版）
func (m *AIManager) selectByRoundRobin(providers []ProviderConfig) (AIProvider, string, error) {
	// 这里简化处理，实际应该维护轮询状态
	return m.selectByPriority(providers)
}

// isCircuitClosed 检查断路器是否关闭
func (m *AIManager) isCircuitClosed(providerName string) bool {
	cb, exists := m.circuitBreakers[providerName]
	if !exists {
		return true
	}
	
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	switch cb.state {
	case "closed":
		return true
	case "open":
		// 检查是否到了恢复时间
		if time.Since(cb.lastFailure) > cb.RecoveryTimeout {
			cb.state = "half-open"
			return true
		}
		return false
	case "half-open":
		return true
	default:
		return false
	}
}

// recordSuccess 记录成功调用
func (m *AIManager) recordSuccess(providerName string, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 更新统计
	if stats, exists := m.stats[providerName]; exists {
		stats.TotalRequests++
		stats.SuccessCount++
		stats.LastUsed = time.Now()
		// 简化的平均延迟计算
		stats.AvgLatency = (stats.AvgLatency + latency) / 2
	}
	
	// 重置断路器
	if cb, exists := m.circuitBreakers[providerName]; exists {
		cb.mu.Lock()
		cb.failures = 0
		cb.state = "closed"
		cb.mu.Unlock()
	}
}

// recordFailure 记录失败调用
func (m *AIManager) recordFailure(providerName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 更新统计
	if stats, exists := m.stats[providerName]; exists {
		stats.TotalRequests++
		stats.FailureCount++
	}
	
	// 更新断路器
	if cb, exists := m.circuitBreakers[providerName]; exists {
		cb.mu.Lock()
		cb.failures++
		cb.lastFailure = time.Now()
		
		if cb.failures >= cb.FailureThreshold {
			cb.state = "open"
			log.Printf("断路器开启: %s (失败次数: %d)", providerName, cb.failures)
		}
		cb.mu.Unlock()
	}
}

// GenerateContent 生成内容（支持降级）
func (m *AIManager) GenerateContent(ctx context.Context, prompt string) (string, error) {
	maxRetries := 3
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		provider, providerName, err := m.selectProvider(ctx)
		if err != nil {
			return "", fmt.Errorf("选择提供商失败: %v", err)
		}
		
		start := time.Now()
		content, err := provider.GenerateContent(ctx, prompt)
		latency := time.Since(start)
		
		if err != nil {
			log.Printf("提供商 %s 调用失败: %v", providerName, err)
			m.recordFailure(providerName)
			
			// 如果是配额问题，直接尝试下一个提供商
			if aiErr, ok := err.(*AIError); ok && aiErr.IsQuotaExceeded() {
				continue
			}
			
			// 其他错误也尝试下一个提供商
			continue
		}
		
		m.recordSuccess(providerName, latency)
		log.Printf("成功使用提供商: %s (耗时: %v)", providerName, latency)
		return content, nil
	}
	
	return "", fmt.Errorf("所有AI提供商都不可用")
}

// GenerateArticle 生成文章（支持降级）
func (m *AIManager) GenerateArticle(ctx context.Context, title, topic string) (*ArticleContent, error) {
	maxRetries := 3
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		provider, providerName, err := m.selectProvider(ctx)
		if err != nil {
			return nil, fmt.Errorf("选择提供商失败: %v", err)
		}
		
		start := time.Now()
		content, err := provider.GenerateArticle(ctx, title, topic)
		latency := time.Since(start)
		
		if err != nil {
			log.Printf("提供商 %s 调用失败: %v", providerName, err)
			m.recordFailure(providerName)
			continue
		}
		
		// 设置生成信息
		content.GeneratedBy = providerName
		content.GeneratedAt = time.Now()
		
		m.recordSuccess(providerName, latency)
		log.Printf("成功使用提供商: %s (耗时: %v)", providerName, latency)
		return content, nil
	}
	
	return nil, fmt.Errorf("所有AI提供商都不可用")
}

// GetStats 获取统计信息
func (m *AIManager) GetStats() map[string]*ProviderStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// 深拷贝避免并发问题
	result := make(map[string]*ProviderStats)
	for k, v := range m.stats {
		result[k] = &ProviderStats{
			TotalRequests: v.TotalRequests,
			SuccessCount:  v.SuccessCount,
			FailureCount:  v.FailureCount,
			AvgLatency:    v.AvgLatency,
			LastUsed:      v.LastUsed,
		}
	}
	
	return result
}