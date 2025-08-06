package upload

import (
	"crypto/md5"
	"fmt"
	"image"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "image/gif"  // 支持GIF格式
	_ "image/jpeg" // 支持JPEG格式
	_ "image/png"  // 支持PNG格式
)

// UploadConfig 上传配置
type UploadConfig struct {
	MaxFileSize   int64    `json:"max_file_size"`   // 最大文件大小(字节)
	AllowedTypes  []string `json:"allowed_types"`   // 允许的文件类型
	UploadDir     string   `json:"upload_dir"`      // 上传目录
	URLPrefix     string   `json:"url_prefix"`      // URL前缀
	CreateSubDirs bool     `json:"create_sub_dirs"` // 是否按日期创建子目录
}

// DefaultConfig 默认配置
func DefaultConfig() *UploadConfig {
	return &UploadConfig{
		MaxFileSize:   10 * 1024 * 1024, // 10MB
		AllowedTypes:  []string{"jpg", "jpeg", "png", "gif", "webp"},
		UploadDir:     "./uploads",
		URLPrefix:     "/uploads",
		CreateSubDirs: true,
	}
}

// FileInfo 上传文件信息
type FileInfo struct {
	FileName    string `json:"file_name"`
	StoredName  string `json:"stored_name"`
	FileSize    int64  `json:"file_size"`
	FileType    string `json:"file_type"`
	MimeType    string `json:"mime_type"`
	FilePath    string `json:"file_path"`
	FileURL     string `json:"file_url"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	IsImage     bool   `json:"is_image"`
}

// UploadService 文件上传服务
type UploadService struct {
	config *UploadConfig
}

// NewUploadService 创建上传服务
func NewUploadService(config *UploadConfig) *UploadService {
	if config == nil {
		config = DefaultConfig()
	}
	return &UploadService{config: config}
}

// ValidateFile 验证文件
func (s *UploadService) ValidateFile(header *multipart.FileHeader) error {
	// 检查文件大小
	if header.Size > s.config.MaxFileSize {
		return fmt.Errorf("文件大小超过限制: %d bytes", s.config.MaxFileSize)
	}

	// 检查文件类型
	fileType := strings.ToLower(filepath.Ext(header.Filename))
	if fileType != "" && fileType[0] == '.' {
		fileType = fileType[1:]
	}

	allowed := false
	for _, t := range s.config.AllowedTypes {
		if fileType == t {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("不支持的文件类型: %s, 支持的类型: %v", fileType, s.config.AllowedTypes)
	}

	return nil
}

// generateStoredName 生成存储文件名
func (s *UploadService) generateStoredName(fileName string) string {
	ext := filepath.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)
	
	// 使用文件名和当前时间生成MD5
	hash := md5.New()
	hash.Write([]byte(fmt.Sprintf("%s_%d", name, time.Now().UnixNano())))
	
	return fmt.Sprintf("%x%s", hash.Sum(nil), ext)
}

// getImageDimensions 获取图片尺寸
func (s *UploadService) getImageDimensions(filePath string) (int, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return config.Width, config.Height, nil
}

// isImageFile 检查是否为图片文件
func (s *UploadService) isImageFile(fileType string) bool {
	imageTypes := []string{"jpg", "jpeg", "png", "gif", "webp"}
	for _, t := range imageTypes {
		if fileType == t {
			return true
		}
	}
	return false
}

// ensureDir 确保目录存在
func (s *UploadService) ensureDir(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

// getUploadPath 获取上传路径
func (s *UploadService) getUploadPath() (string, string, error) {
	baseDir := s.config.UploadDir
	
	var subDir string
	if s.config.CreateSubDirs {
		// 按日期创建子目录: uploads/2024/01/02/
		now := time.Now()
		subDir = now.Format("2006/01/02")
	}
	
	fullPath := filepath.Join(baseDir, subDir)
	
	// 确保目录存在
	if err := s.ensureDir(fullPath); err != nil {
		return "", "", fmt.Errorf("创建上传目录失败: %v", err)
	}
	
	return fullPath, subDir, nil
}

// UploadFile 上传文件
func (s *UploadService) UploadFile(fileHeader *multipart.FileHeader) (*FileInfo, error) {
	// 验证文件
	if err := s.ValidateFile(fileHeader); err != nil {
		return nil, err
	}

	// 打开上传文件
	srcFile, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %v", err)
	}
	defer srcFile.Close()

	// 获取上传路径
	uploadPath, subDir, err := s.getUploadPath()
	if err != nil {
		return nil, err
	}

	// 生成存储文件名
	storedName := s.generateStoredName(fileHeader.Filename)
	filePath := filepath.Join(uploadPath, storedName)

	// 创建目标文件
	dstFile, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer dstFile.Close()

	// 复制文件内容
	fileSize, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return nil, fmt.Errorf("保存文件失败: %v", err)
	}

	// 获取文件类型
	fileType := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if fileType != "" && fileType[0] == '.' {
		fileType = fileType[1:]
	}

	// 构建文件信息
	fileInfo := &FileInfo{
		FileName:   fileHeader.Filename,
		StoredName: storedName,
		FileSize:   fileSize,
		FileType:   fileType,
		MimeType:   fileHeader.Header.Get("Content-Type"),
		FilePath:   filePath,
		FileURL:    s.config.URLPrefix + "/" + filepath.Join(subDir, storedName),
		IsImage:    s.isImageFile(fileType),
	}

	// 如果是图片，获取尺寸信息
	if fileInfo.IsImage {
		width, height, err := s.getImageDimensions(filePath)
		if err == nil {
			fileInfo.Width = width
			fileInfo.Height = height
		}
	}

	return fileInfo, nil
}