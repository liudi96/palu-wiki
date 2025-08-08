package models

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

// File 文件模型
type File struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	FileName    string         `json:"file_name" gorm:"size:255;not null;comment:原始文件名"`
	StoredName  string         `json:"stored_name" gorm:"size:255;not null;unique;comment:存储文件名"`
	FileSize    int64          `json:"file_size" gorm:"not null;comment:文件大小(字节)"`
	FileType    string         `json:"file_type" gorm:"size:100;not null;comment:文件类型"`
	MimeType    string         `json:"mime_type" gorm:"size:255;not null;comment:MIME类型"`
	FilePath    string         `json:"file_path" gorm:"size:500;not null;comment:文件路径"`
	FileURL     string         `json:"file_url" gorm:"size:500;not null;comment:访问URL"`
	Width       int            `json:"width" gorm:"default:0;comment:图片宽度"`
	Height      int            `json:"height" gorm:"default:0;comment:图片高度"`
	UploaderID  uint           `json:"uploader_id" gorm:"not null;comment:上传用户ID"`
	Uploader    User           `json:"uploader" gorm:"foreignKey:UploaderID"`
	Status      string         `json:"status" gorm:"size:20;default:'active';comment:状态:active,deleted"`
	Description string         `json:"description" gorm:"size:500;comment:文件描述"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName 指定表名
func (File) TableName() string {
	return "files"
}

// IsImage 检查是否为图片文件
func (f *File) IsImage() bool {
	switch f.FileType {
	case "jpg", "jpeg", "png", "gif", "webp", "svg", "bmp":
		return true
	default:
		return false
	}
}

// GetSizeFormatted 获取格式化的文件大小
func (f *File) GetSizeFormatted() string {
	size := float64(f.FileSize)
	units := []string{"B", "KB", "MB", "GB"}

	for i, unit := range units {
		if size < 1024 || i == len(units)-1 {
			if i == 0 {
				return fmt.Sprintf("%.0f %s", size, unit)
			}
			return fmt.Sprintf("%.2f %s", size, unit)
		}
		size /= 1024
	}

	return fmt.Sprintf("%.2f %s", size, units[len(units)-1])
}
