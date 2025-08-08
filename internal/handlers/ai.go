package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"palu-wiki/internal/models"
	"palu-wiki/pkg/ai"
)

// AIHandler AI相关处理器
type AIHandler struct {
	db       *gorm.DB
	aiClient *ai.SparkAIClient
}

// NewAIHandler 创建AI处理器
func NewAIHandler(db *gorm.DB, aiClient *ai.SparkAIClient) *AIHandler {
	return &AIHandler{
		db:       db,
		aiClient: aiClient,
	}
}

// GenerateArticleRequest AI生成文章请求
type GenerateArticleRequest struct {
	Title      string `json:"title" binding:"required"`
	Topic      string `json:"topic"`       // 可选字段
	CategoryID uint   `json:"category_id"` // 可选字段
}

// GenerateArticle AI生成文章
func (h *AIHandler) GenerateArticle(c *gin.Context) {
	var req GenerateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 获取当前用户ID（从JWT中间件获取）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 设置默认topic（如果未提供）
	topic := req.Topic
	if topic == "" {
		topic = "帕鲁游戏攻略"
	}

	// 设置默认分类ID（如果未提供，使用第一个可用分类）
	categoryID := req.CategoryID
	if categoryID == 0 {
		var category models.Category
		if err := h.db.First(&category).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "未找到可用分类"})
			return
		}
		categoryID = category.ID
	} else {
		// 验证分类是否存在
		var category models.Category
		if err := h.db.First(&category, categoryID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "分类不存在"})
			return
		}
	}

	// 调用AI生成内容
	aiContent, err := h.aiClient.GenerateArticle(c.Request.Context(), req.Title, topic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI生成内容失败: " + err.Error()})
		return
	}

	// 创建文章
	article := models.Article{
		Title:         aiContent.Title,
		Content:       aiContent.Content,
		Summary:       aiContent.Summary,
		CategoryID:    categoryID,
		AuthorID:      userID.(uint),
		Status:        "draft", // AI生成的文章默认为草稿状态
		IsAIGenerated: true,
		Tags:          `["` + topic + `", "AI生成", "攻略"]`,
	}

	if err := h.db.Create(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文章失败: " + err.Error()})
		return
	}

	// 预加载关联数据
	h.db.Preload("Author").Preload("Category").First(&article, article.ID)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "AI文章生成成功",
		"data": gin.H{
			"content": aiContent.Content,
			"summary": aiContent.Summary,
		},
	})
}

// GenerateContentRequest 简单内容生成请求
type GenerateContentRequest struct {
	Title      string `json:"title" binding:"required"`
	Topic      string `json:"topic"`
	CategoryID uint   `json:"category_id"`
}

// GenerateContent 生成内容（不保存为文章）
func (h *AIHandler) GenerateContent(c *gin.Context) {
	var req GenerateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 设置默认topic
	topic := req.Topic
	if topic == "" {
		topic = "帕鲁游戏攻略"
	}

	// 调用AI生成内容
	aiContent, err := h.aiClient.GenerateArticle(c.Request.Context(), req.Title, topic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI生成内容失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "内容生成成功",
		"data": gin.H{
			"content": aiContent.Content,
			"summary": aiContent.Summary,
		},
	})
}

// RegenerateArticleContent 重新生成文章内容
func (h *AIHandler) RegenerateArticleContent(c *gin.Context) {
	articleIDStr := c.Param("id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文章ID格式错误"})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 查找文章
	var article models.Article
	if err := h.db.First(&article, uint(articleID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	// 检查权限（只有作者可以重新生成）
	if article.AuthorID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作此文章"})
		return
	}

	// 重新生成内容
	topic := c.DefaultQuery("topic", "帕鲁攻略")
	aiContent, err := h.aiClient.GenerateArticle(c.Request.Context(), article.Title, topic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI重新生成内容失败: " + err.Error()})
		return
	}

	// 更新文章内容
	article.Content = aiContent.Content
	article.Summary = aiContent.Summary
	article.Status = "draft" // 重新生成后设为草稿

	if err := h.db.Save(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新文章失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "文章内容重新生成成功",
		"data":    article,
	})
}

// OptimizeArticleContent 优化文章内容
func (h *AIHandler) OptimizeArticleContent(c *gin.Context) {
	articleIDStr := c.Param("id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文章ID格式错误"})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 查找文章
	var article models.Article
	if err := h.db.First(&article, uint(articleID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	// 检查权限
	if article.AuthorID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作此文章"})
		return
	}

	// 构建优化提示词
	optimizePrompt := fmt.Sprintf(`请优化以下《幻兽帕鲁》攻略文章，使其更加实用和详细：

原文章标题：%s
原文章内容：
%s

优化要求：
1. 保持原有结构和主要内容
2. 补充更多实用细节和技巧
3. 修正可能的错误信息
4. 优化语言表达，使其更加流畅
5. 保持Markdown格式

请直接返回优化后的文章内容：`, article.Title, article.Content)

	optimizedContent, err := h.aiClient.GenerateContent(c.Request.Context(), optimizePrompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI优化内容失败: " + err.Error()})
		return
	}

	// 更新文章内容
	article.Content = optimizedContent
	article.Status = "draft" // 优化后设为草稿

	if err := h.db.Save(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新文章失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "文章内容优化成功",
		"data":    article,
	})
}
