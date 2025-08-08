package models

import (
	"time"

	"gorm.io/gorm"
)

type Article struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	Title         string         `json:"title" gorm:"not null"`
	Content       string         `json:"content" gorm:"type:text"`
	Summary       string         `json:"summary" gorm:"type:text"`
	Cover         string         `json:"cover"`
	Status        string         `json:"status" gorm:"default:draft"` // draft, published, archived
	ViewCount     int            `json:"view_count" gorm:"default:0"`
	LikeCount     int            `json:"like_count" gorm:"default:0"`
	CategoryID    uint           `json:"category_id"`
	AuthorID      uint           `json:"author_id"`
	IsAIGenerated bool           `json:"is_ai_generated" gorm:"default:false"`
	Tags          string         `json:"tags"` // JSON字符串存储标签
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Author   User      `json:"author" gorm:"foreignKey:AuthorID"`
	Category Category  `json:"category" gorm:"foreignKey:CategoryID"`
	Comments []Comment `json:"comments,omitempty" gorm:"foreignKey:ArticleID"`
}
