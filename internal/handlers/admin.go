package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"palu-wiki/internal/models"
)

// AdminHandler 管理后台处理器
type AdminHandler struct {
	db *gorm.DB
}

// NewAdminHandler 创建管理后台处理器
func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// Dashboard 获取管理后台仪表板数据
func (h *AdminHandler) Dashboard(c *gin.Context) {
	var stats struct {
		TotalUsers        int64 `json:"total_users"`
		TotalArticles     int64 `json:"total_articles"`
		TotalCategories   int64 `json:"total_categories"`
		TotalComments     int64 `json:"total_comments"`
		AIArticles        int64 `json:"ai_articles"`
		PendingArticles   int64 `json:"pending_articles"`
		PublishedArticles int64 `json:"published_articles"`
		DraftArticles     int64 `json:"draft_articles"`
	}

	// 统计用户数
	h.db.Model(&models.User{}).Count(&stats.TotalUsers)

	// 统计文章数
	h.db.Model(&models.Article{}).Count(&stats.TotalArticles)
	h.db.Model(&models.Article{}).Where("is_ai_generated = ?", true).Count(&stats.AIArticles)
	h.db.Model(&models.Article{}).Where("status = ?", "pending").Count(&stats.PendingArticles)
	h.db.Model(&models.Article{}).Where("status = ?", "published").Count(&stats.PublishedArticles)
	h.db.Model(&models.Article{}).Where("status = ?", "draft").Count(&stats.DraftArticles)

	// 统计分类数
	h.db.Model(&models.Category{}).Count(&stats.TotalCategories)

	// 统计评论数
	h.db.Model(&models.Comment{}).Count(&stats.TotalComments)

	// 获取最近的文章
	var recentArticles []models.Article
	h.db.Preload("Author").Preload("Category").
		Order("created_at DESC").
		Limit(5).
		Find(&recentArticles)

	// 获取最近的用户
	var recentUsers []models.User
	h.db.Select("id, username, email, nickname, role, status, created_at").
		Order("created_at DESC").
		Limit(5).
		Find(&recentUsers)

	c.JSON(http.StatusOK, gin.H{
		"stats":           stats,
		"recent_articles": recentArticles,
		"recent_users":    recentUsers,
	})
}

// GetAllUsers 获取所有用户（管理员功能）
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 搜索参数
	keyword := c.Query("q")
	role := c.Query("role")
	status := c.Query("status")

	// 构建查询
	query := h.db.Select("id, username, email, nickname, role, status, created_at, updated_at")

	if keyword != "" {
		query = query.Where("username ILIKE ? OR email ILIKE ? OR nickname ILIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if role != "" {
		query = query.Where("role = ?", role)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 计算总数
	var total int64
	if err := query.Model(&models.User{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "统计用户数失败"})
		return
	}

	// 分页查询
	var users []models.User
	offset := (page - 1) * pageSize
	result := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&users)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// UpdateUserStatus 更新用户状态
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required"`
		Role   string `json:"role,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 验证状态值
	if req.Status != "active" && req.Status != "inactive" && req.Status != "banned" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态值"})
		return
	}

	// 验证角色值（如果提供）
	if req.Role != "" && req.Role != "user" && req.Role != "editor" && req.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色值"})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败"})
		return
	}

	// 更新用户状态
	updates := map[string]interface{}{
		"status": req.Status,
	}

	if req.Role != "" {
		updates["role"] = req.Role
	}

	if err := h.db.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户状态更新成功",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"status":   req.Status,
			"role":     user.Role,
		},
	})
}

// GetAllArticles 获取所有文章（管理员功能）
func (h *AdminHandler) GetAllArticles(c *gin.Context) {
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 过滤参数
	status := c.Query("status")
	categoryID := c.Query("category_id")
	authorID := c.Query("author_id")
	isAIGenerated := c.Query("is_ai_generated")

	// 构建查询
	query := h.db.Preload("Author").Preload("Category")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	if authorID != "" {
		query = query.Where("author_id = ?", authorID)
	}

	if isAIGenerated == "true" {
		query = query.Where("is_ai_generated = ?", true)
	} else if isAIGenerated == "false" {
		query = query.Where("is_ai_generated = ?", false)
	}

	// 计算总数
	var total int64
	if err := query.Model(&models.Article{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "统计文章数失败"})
		return
	}

	// 分页查询
	var articles []models.Article
	offset := (page - 1) * pageSize
	result := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&articles)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文章失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": articles,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// UpdateArticleStatus 更新文章状态（审核功能）
func (h *AdminHandler) UpdateArticleStatus(c *gin.Context) {
	articleID := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 验证状态值
	if req.Status != "draft" && req.Status != "pending" && req.Status != "published" && req.Status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态值"})
		return
	}

	// 查找文章
	var article models.Article
	if err := h.db.First(&article, articleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文章失败"})
		return
	}

	// 更新文章状态
	if err := h.db.Model(&article).Update("status", req.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新文章状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("文章状态已更新为: %s", req.Status),
		"article": gin.H{
			"id":     article.ID,
			"title":  article.Title,
			"status": req.Status,
		},
	})
}

// GetSystemInfo 获取系统信息
func (h *AdminHandler) GetSystemInfo(c *gin.Context) {
	// 获取数据库统计信息
	var dbStats struct {
		Tables []struct {
			Name  string `json:"name"`
			Count int64  `json:"count"`
		} `json:"tables"`
	}

	// 统计各表数据量
	tables := []struct {
		Name  string
		Model interface{}
	}{
		{"users", &models.User{}},
		{"articles", &models.Article{}},
		{"categories", &models.Category{}},
		{"comments", &models.Comment{}},
	}

	for _, table := range tables {
		var count int64
		h.db.Model(table.Model).Count(&count)
		dbStats.Tables = append(dbStats.Tables, struct {
			Name  string `json:"name"`
			Count int64  `json:"count"`
		}{
			Name:  table.Name,
			Count: count,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"system": gin.H{
			"version":    "v1.0.0",
			"build_time": time.Now().Format("2006-01-02 15:04:05"),
			"database":   dbStats,
		},
	})
}
