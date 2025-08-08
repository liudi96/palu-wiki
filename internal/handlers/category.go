package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"palu-wiki/internal/models"
)

type CategoryHandler struct {
	db *gorm.DB
}

func NewCategoryHandler(db *gorm.DB) *CategoryHandler {
	return &CategoryHandler{db: db}
}

// 获取分类列表
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	var categories []models.Category

	result := h.db.Where("status = ?", "active").
		Order("sort ASC, created_at DESC").
		Find(&categories)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取分类失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// 获取分类下的文章
func (h *CategoryHandler) GetCategoryArticles(c *gin.Context) {
	id := c.Param("id")
	var articles []models.Article

	result := h.db.Preload("Author").Preload("Category").
		Where("category_id = ? AND status = ?", id, "published").
		Order("created_at DESC").
		Find(&articles)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取分类文章失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": articles})
}

// 创建分类
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.db.Create(&category)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建分类失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": category})
}

// 更新分类
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	// 查找分类
	if err := h.db.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "分类不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询分类失败"})
		return
	}

	// 绑定更新数据
	var updateData models.Category
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新分类
	if err := h.db.Model(&category).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新分类失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": category})
}

// 删除分类
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	// 查找分类
	if err := h.db.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "分类不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询分类失败"})
		return
	}

	// 检查是否有文章使用此分类
	var articleCount int64
	h.db.Model(&models.Article{}).Where("category_id = ?", id).Count(&articleCount)
	if articleCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该分类下还有文章，无法删除"})
		return
	}

	// 软删除分类
	if err := h.db.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除分类失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "分类删除成功"})
}
