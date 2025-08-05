package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"palu-wiki/internal/models"
)

type ArticleHandler struct {
	db *gorm.DB
}

func NewArticleHandler(db *gorm.DB) *ArticleHandler {
	return &ArticleHandler{db: db}
}

// 获取文章列表
func (h *ArticleHandler) GetArticles(c *gin.Context) {
	var articles []models.Article
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	result := h.db.Preload("Author").Preload("Category").
		Where("status = ?", "published").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&articles)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文章列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": articles,
		"page": page,
		"limit": limit,
	})
}

// 获取文章详情
func (h *ArticleHandler) GetArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article

	result := h.db.Preload("Author").Preload("Category").Preload("Comments.User").
		First(&article, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文章失败"})
		return
	}

	// 增加浏览量
	h.db.Model(&article).Update("view_count", article.ViewCount+1)

	c.JSON(http.StatusOK, gin.H{"data": article})
}

// 创建文章
func (h *ArticleHandler) CreateArticle(c *gin.Context) {
	var article models.Article
	if err := c.ShouldBindJSON(&article); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	article.AuthorID = userID.(uint)

	result := h.db.Create(&article)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文章失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "文章创建成功",
		"data":    article,
	})
}

// 更新文章
func (h *ArticleHandler) UpdateArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article

	// 查找文章
	if err := h.db.First(&article, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文章失败"})
		return
	}

	// 检查权限（只有作者或管理员可以修改）
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")
	if article.AuthorID != userID.(uint) && userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限修改此文章"})
		return
	}

	// 绑定更新数据
	var updateData models.Article
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新文章
	if err := h.db.Model(&article).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新文章失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": article})
}

// 删除文章
func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article

	// 查找文章
	if err := h.db.First(&article, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文章失败"})
		return
	}

	// 软删除文章
	if err := h.db.Delete(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除文章失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文章删除成功"})
}

// 搜索文章
func (h *ArticleHandler) SearchArticles(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	
	// 搜索条件
	categoryID := c.Query("category_id")
	authorID := c.Query("author_id")
	isAIGenerated := c.Query("is_ai_generated")
	status := c.DefaultQuery("status", "published")
	
	// 排序方式
	sortBy := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")
	
	// 构建查询
	query := h.db.Preload("Author").Preload("Category").
		Where("title ILIKE ? OR content ILIKE ? OR summary ILIKE ?", 
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	
	// 添加过滤条件
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	
	if authorID != "" {
		query = query.Where("author_id = ?", authorID)
	}
	
	if isAIGenerated != "" {
		if isAIGenerated == "true" {
			query = query.Where("is_ai_generated = ?", true)
		} else if isAIGenerated == "false" {
			query = query.Where("is_ai_generated = ?", false)
		}
	}
	
	// 计算总数
	var total int64
	if err := query.Model(&models.Article{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "统计搜索结果失败"})
		return
	}
	
	// 分页查询
	var articles []models.Article
	offset := (page - 1) * pageSize
	
	// 排序
	orderBy := fmt.Sprintf("%s %s", sortBy, strings.ToUpper(order))
	result := query.Order(orderBy).Limit(pageSize).Offset(offset).Find(&articles)
	
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败"})
		return
	}
	
	// 返回分页结果
	c.JSON(http.StatusOK, gin.H{
		"data": articles,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
		"search_info": gin.H{
			"keyword":         keyword,
			"category_id":     categoryID,
			"author_id":       authorID,
			"is_ai_generated": isAIGenerated,
			"status":          status,
			"sort_by":         sortBy,
			"order":           order,
		},
	})
}