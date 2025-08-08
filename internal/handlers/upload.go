package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"palu-wiki/internal/models"
	"palu-wiki/pkg/upload"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	db            *gorm.DB
	uploadService *upload.UploadService
}

// NewUploadHandler 创建文件上传处理器
func NewUploadHandler(db *gorm.DB) *UploadHandler {
	// 创建上传配置
	config := &upload.UploadConfig{
		MaxFileSize:   10 * 1024 * 1024, // 10MB
		AllowedTypes:  []string{"jpg", "jpeg", "png", "gif", "webp"},
		UploadDir:     "./uploads",
		URLPrefix:     "/uploads",
		CreateSubDirs: true,
	}

	return &UploadHandler{
		db:            db,
		uploadService: upload.NewUploadService(config),
	}
}

// UploadImage 上传图片
func (h *UploadHandler) UploadImage(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取上传的文件
	fileHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取上传文件失败: " + err.Error()})
		return
	}

	// 上传文件
	fileInfo, err := h.uploadService.UploadFile(fileHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 保存文件信息到数据库
	file := &models.File{
		FileName:    fileInfo.FileName,
		StoredName:  fileInfo.StoredName,
		FileSize:    fileInfo.FileSize,
		FileType:    fileInfo.FileType,
		MimeType:    fileInfo.MimeType,
		FilePath:    fileInfo.FilePath,
		FileURL:     fileInfo.FileURL,
		Width:       fileInfo.Width,
		Height:      fileInfo.Height,
		UploaderID:  userID.(uint),
		Status:      "active",
		Description: c.PostForm("description"),
	}

	if err := h.db.Create(file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件信息失败: " + err.Error()})
		return
	}

	// 预加载上传者信息
	h.db.Preload("Uploader").First(file, file.ID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "图片上传成功",
		"data":    file,
	})
}

// UploadFile 上传文件(通用)
func (h *UploadHandler) UploadFile(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取上传的文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取上传文件失败: " + err.Error()})
		return
	}

	// 上传文件
	fileInfo, err := h.uploadService.UploadFile(fileHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 保存文件信息到数据库
	file := &models.File{
		FileName:    fileInfo.FileName,
		StoredName:  fileInfo.StoredName,
		FileSize:    fileInfo.FileSize,
		FileType:    fileInfo.FileType,
		MimeType:    fileInfo.MimeType,
		FilePath:    fileInfo.FilePath,
		FileURL:     fileInfo.FileURL,
		Width:       fileInfo.Width,
		Height:      fileInfo.Height,
		UploaderID:  userID.(uint),
		Status:      "active",
		Description: c.PostForm("description"),
	}

	if err := h.db.Create(file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件信息失败: " + err.Error()})
		return
	}

	// 预加载上传者信息
	h.db.Preload("Uploader").First(file, file.ID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "文件上传成功",
		"data":    file,
	})
}

// GetFiles 获取文件列表
func (h *UploadHandler) GetFiles(c *gin.Context) {
	var files []models.File
	var total int64

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 筛选条件
	query := h.db.Model(&models.File{}).Where("status = ?", "active")

	// 文件类型筛选
	if fileType := c.Query("type"); fileType != "" {
		if fileType == "image" {
			query = query.Where("file_type IN ?", []string{"jpg", "jpeg", "png", "gif", "webp"})
		} else {
			query = query.Where("file_type = ?", fileType)
		}
	}

	// 用户筛选
	if uploaderID := c.Query("uploader_id"); uploaderID != "" {
		query = query.Where("uploader_id = ?", uploaderID)
	}

	// 统计总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	result := query.Preload("Uploader").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&files)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    files,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetFile 获取文件详情
func (h *UploadHandler) GetFile(c *gin.Context) {
	id := c.Param("id")
	var file models.File

	result := h.db.Preload("Uploader").Where("status = ?", "active").First(&file, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    file,
	})
}

// DeleteFile 删除文件
func (h *UploadHandler) DeleteFile(c *gin.Context) {
	id := c.Param("id")
	var file models.File

	// 查找文件
	result := h.db.Where("status = ?", "active").First(&file, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文件失败"})
		return
	}

	// 检查权限（只有文件上传者或管理员可以删除）
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")
	if file.UploaderID != userID.(uint) && userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限删除此文件"})
		return
	}

	// 软删除文件记录
	if err := h.db.Model(&file).Update("status", "deleted").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除文件失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "文件删除成功",
	})
}

// UpdateFile 更新文件信息
func (h *UploadHandler) UpdateFile(c *gin.Context) {
	id := c.Param("id")
	var file models.File

	// 查找文件
	result := h.db.Where("status = ?", "active").First(&file, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询文件失败"})
		return
	}

	// 检查权限
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")
	if file.UploaderID != userID.(uint) && userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限修改此文件"})
		return
	}

	// 更新描述
	if description := c.PostForm("description"); description != "" {
		file.Description = description
	}

	if err := h.db.Save(&file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新文件失败"})
		return
	}

	// 预加载关联数据
	h.db.Preload("Uploader").First(&file, file.ID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "文件信息更新成功",
		"data":    file,
	})
}
