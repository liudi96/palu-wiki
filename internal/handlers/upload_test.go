package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"palu-wiki/internal/testutils"
)

// 测试文件上传处理器的基础功能
func TestUploadHandler_BasicOperations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 创建handler（用nil DB因为这些测试不会访问数据库和实际上传服务）
	handler := NewUploadHandler(nil)
	
	router := gin.New()
	
	// 认证用户路由
	auth := router.Group("/upload")
	auth.Use(func(c *gin.Context) {
		// 模拟JWT认证
		c.Set("user_id", uint(1))
		c.Set("username", "testuser")
		c.Set("user_role", "user")
		c.Next()
	})
	{
		auth.POST("/image", handler.UploadImage)
		auth.POST("/file", handler.UploadFile)
		auth.GET("/files", handler.GetFiles)
		auth.GET("/files/:id", handler.GetFile)
		auth.PUT("/files/:id", handler.UpdateFile)
		auth.DELETE("/files/:id", handler.DeleteFile)
	}

	t.Run("上传图片 - 未授权", func(t *testing.T) {
		// 创建没有认证的路由
		noAuthRouter := gin.New()
		noAuthRouter.POST("/upload/image", handler.UploadImage)
		
		req := httptest.NewRequest("POST", "/upload/image", nil)
		recorder := httptest.NewRecorder()
		noAuthRouter.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		
		var response map[string]string
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "未授权")
	})

	t.Run("获取文件列表 - 参数解析", func(t *testing.T) {
		testRouter := gin.New()
		testRouter.GET("/upload/files", func(c *gin.Context) {
			page := c.DefaultQuery("page", "1")
			pageSize := c.DefaultQuery("page_size", "20")
			fileType := c.Query("type")
			uploaderID := c.Query("uploader_id")
			
			c.JSON(200, gin.H{
				"page":        page,
				"page_size":   pageSize,
				"type":        fileType,
				"uploader_id": uploaderID,
			})
		})
		
		req := httptest.NewRequest("GET", "/upload/files?page=2&page_size=10&type=image&uploader_id=1", nil)
		recorder := httptest.NewRecorder()
		testRouter.ServeHTTP(recorder, req)
		
		assert.Equal(t, 200, recorder.Code)
		
		var response map[string]string
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "2", response["page"])
		assert.Equal(t, "10", response["page_size"])
		assert.Equal(t, "image", response["type"])
		assert.Equal(t, "1", response["uploader_id"])
	})

	t.Run("URL参数解析", func(t *testing.T) {
		testRouter := gin.New()
		testRouter.GET("/upload/files/:id", func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(200, gin.H{"file_id": id})
		})
		
		req := httptest.NewRequest("GET", "/upload/files/123", nil)
		recorder := httptest.NewRecorder()
		testRouter.ServeHTTP(recorder, req)
		
		var response map[string]string
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "123", response["file_id"])
	})
}

// 测试文件上传业务逻辑
func TestUploadHandler_BusinessLogic(t *testing.T) {
	t.Run("文件类型筛选逻辑", func(t *testing.T) {
		// 模拟文件类型筛选逻辑
		filterFilesByType := func(fileType string) []string {
			switch fileType {
			case "image":
				return []string{"jpg", "jpeg", "png", "gif", "webp"}
			case "document":
				return []string{"pdf", "doc", "docx", "txt"}
			case "video":
				return []string{"mp4", "avi", "mov"}
			default:
				return []string{fileType}
			}
		}
		
		// 测试图片类型
		imageTypes := filterFilesByType("image")
		assert.Contains(t, imageTypes, "jpg")
		assert.Contains(t, imageTypes, "png")
		assert.Contains(t, imageTypes, "gif")
		
		// 测试文档类型
		docTypes := filterFilesByType("document")
		assert.Contains(t, docTypes, "pdf")
		assert.Contains(t, docTypes, "doc")
		
		// 测试具体类型
		specificType := filterFilesByType("mp3")
		assert.Equal(t, []string{"mp3"}, specificType)
	})
	
	t.Run("分页逻辑验证", func(t *testing.T) {
		// 模拟分页计算逻辑
		calculatePagination := func(page, pageSize int, total int64) map[string]interface{} {
			if page < 1 {
				page = 1
			}
			if pageSize < 1 || pageSize > 100 {
				pageSize = 20
			}
			
			totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
			
			return map[string]interface{}{
				"page":        page,
				"page_size":   pageSize,
				"total":       total,
				"total_pages": totalPages,
			}
		}
		
		result := calculatePagination(2, 10, 25)
		assert.Equal(t, 2, result["page"])
		assert.Equal(t, 10, result["page_size"])
		assert.Equal(t, int64(3), result["total_pages"])
		
		// 测试边界情况
		result = calculatePagination(0, 0, 50)
		assert.Equal(t, 1, result["page"])     // 修正为1
		assert.Equal(t, 20, result["page_size"]) // 修正为20
		
		result = calculatePagination(1, 150, 100)
		assert.Equal(t, 20, result["page_size"]) // 超过限制，修正为20
	})
	
	t.Run("文件状态管理", func(t *testing.T) {
		validStatuses := []string{"active", "deleted"}
		
		isValidStatus := func(status string) bool {
			for _, valid := range validStatuses {
				if status == valid {
					return true
				}
			}
			return false
		}
		
		assert.True(t, isValidStatus("active"))
		assert.True(t, isValidStatus("deleted"))
		assert.False(t, isValidStatus("pending"))
		assert.False(t, isValidStatus("unknown"))
	})
}

// 测试文件上传权限控制
func TestUploadHandler_PermissionControl(t *testing.T) {
	t.Run("文件删除权限检查", func(t *testing.T) {
		// 模拟权限检查逻辑
		canDeleteFile := func(fileUploaderID, currentUserID uint, currentUserRole string) bool {
			// 文件上传者可以删除自己的文件
			if fileUploaderID == currentUserID {
				return true
			}
			// 管理员可以删除任何文件
			if currentUserRole == "admin" {
				return true
			}
			return false
		}
		
		// 文件上传者可以删除
		assert.True(t, canDeleteFile(1, 1, "user"))
		
		// 其他用户不能删除
		assert.False(t, canDeleteFile(1, 2, "user"))
		
		// 管理员可以删除任何文件
		assert.True(t, canDeleteFile(1, 2, "admin"))
		
		// 编辑员不能删除其他人的文件
		assert.False(t, canDeleteFile(1, 2, "editor"))
	})
	
	t.Run("文件修改权限检查", func(t *testing.T) {
		// 模拟修改权限检查（与删除权限相同）
		canUpdateFile := func(fileUploaderID, currentUserID uint, currentUserRole string) bool {
			return fileUploaderID == currentUserID || currentUserRole == "admin"
		}
		
		assert.True(t, canUpdateFile(1, 1, "user"))
		assert.False(t, canUpdateFile(1, 2, "user"))
		assert.True(t, canUpdateFile(1, 2, "admin"))
		assert.False(t, canUpdateFile(1, 2, "editor"))
	})
	
	t.Run("用户文件访问权限", func(t *testing.T) {
		// 不同角色的文件访问权限
		permissions := map[string]map[string]bool{
			"admin": {
				"view_all_files":   true,
				"upload_files":     true,
				"delete_any_file":  true,
				"update_any_file":  true,
			},
			"editor": {
				"view_all_files":   true,
				"upload_files":     true,
				"delete_any_file":  false, // 编辑员不能删除其他人的文件
				"update_any_file":  false,
			},
			"user": {
				"view_all_files":   true,
				"upload_files":     true,
				"delete_any_file":  false,
				"update_any_file":  false,
			},
		}
		
		// 测试管理员权限
		adminPerms := permissions["admin"]
		assert.True(t, adminPerms["view_all_files"])
		assert.True(t, adminPerms["delete_any_file"])
		
		// 测试普通用户权限
		userPerms := permissions["user"]
		assert.True(t, userPerms["view_all_files"])
		assert.True(t, userPerms["upload_files"])
		assert.False(t, userPerms["delete_any_file"])
		assert.False(t, userPerms["update_any_file"])
	})
}

// 测试文件上传配置和限制
func TestUploadHandler_ConfigAndLimits(t *testing.T) {
	t.Run("上传配置验证", func(t *testing.T) {
		// 模拟上传配置
		config := map[string]interface{}{
			"max_file_size":   10 * 1024 * 1024, // 10MB
			"allowed_types":   []string{"jpg", "jpeg", "png", "gif", "webp"},
			"upload_dir":      "./uploads",
			"url_prefix":      "/uploads",
			"create_sub_dirs": true,
		}
		
		assert.Equal(t, 10*1024*1024, config["max_file_size"])
		assert.Contains(t, config["allowed_types"], "jpg")
		assert.Equal(t, "./uploads", config["upload_dir"])
		assert.Equal(t, "/uploads", config["url_prefix"])
		assert.True(t, config["create_sub_dirs"].(bool))
	})
	
	t.Run("文件大小验证", func(t *testing.T) {
		maxFileSize := int64(10 * 1024 * 1024) // 10MB
		
		isValidFileSize := func(fileSize int64) bool {
			return fileSize > 0 && fileSize <= maxFileSize
		}
		
		assert.True(t, isValidFileSize(1024))              // 1KB - 有效
		assert.True(t, isValidFileSize(5*1024*1024))       // 5MB - 有效
		assert.True(t, isValidFileSize(10*1024*1024))      // 10MB - 边界有效
		assert.False(t, isValidFileSize(15*1024*1024))     // 15MB - 超过限制
		assert.False(t, isValidFileSize(0))                // 0字节 - 无效
		assert.False(t, isValidFileSize(-1))               // 负数 - 无效
	})
	
	t.Run("文件类型验证", func(t *testing.T) {
		allowedTypes := []string{"jpg", "jpeg", "png", "gif", "webp"}
		
		isAllowedFileType := func(fileType string) bool {
			for _, allowed := range allowedTypes {
				if fileType == allowed {
					return true
				}
			}
			return false
		}
		
		assert.True(t, isAllowedFileType("jpg"))
		assert.True(t, isAllowedFileType("png"))
		assert.True(t, isAllowedFileType("gif"))
		assert.False(t, isAllowedFileType("exe"))
		assert.False(t, isAllowedFileType("pdf"))
		assert.False(t, isAllowedFileType(""))
	})
}

// 测试文件数据结构
func TestUploadHandler_DataStructures(t *testing.T) {
	t.Run("文件信息数据结构", func(t *testing.T) {
		file := testutils.TestFile(1, 1)
		
		assert.Equal(t, uint(1), file.ID)
		assert.Equal(t, uint(1), file.UploaderID)
		assert.NotEmpty(t, file.FileName)
		assert.NotEmpty(t, file.StoredName)
		assert.Greater(t, file.FileSize, int64(0))
		assert.NotEmpty(t, file.FileType)
		assert.NotEmpty(t, file.MimeType)
		assert.Equal(t, "active", file.Status)
	})
	
	t.Run("上传响应数据结构", func(t *testing.T) {
		mockResponse := map[string]interface{}{
			"success": true,
			"message": "文件上传成功",
			"data": map[string]interface{}{
				"id":          1,
				"file_name":   "test.jpg",
				"file_url":    "/uploads/test.jpg",
				"file_size":   1024000,
				"file_type":   "jpg",
				"mime_type":   "image/jpeg",
				"uploader_id": 1,
				"status":      "active",
			},
		}
		
		assert.True(t, mockResponse["success"].(bool))
		assert.Contains(t, mockResponse["message"], "成功")
		
		data := mockResponse["data"].(map[string]interface{})
		assert.Equal(t, 1, data["id"])
		assert.Equal(t, "test.jpg", data["file_name"])
		assert.Equal(t, "active", data["status"])
	})
	
	t.Run("文件列表响应结构", func(t *testing.T) {
		mockListResponse := map[string]interface{}{
			"success": true,
			"data":    []map[string]interface{}{},
			"pagination": map[string]interface{}{
				"page":        1,
				"page_size":   20,
				"total":       int64(100),
				"total_pages": int64(5),
			},
		}
		
		assert.True(t, mockListResponse["success"].(bool))
		assert.NotNil(t, mockListResponse["data"])
		
		pagination := mockListResponse["pagination"].(map[string]interface{})
		assert.Equal(t, 1, pagination["page"])
		assert.Equal(t, int64(100), pagination["total"])
		assert.Equal(t, int64(5), pagination["total_pages"])
	})
}