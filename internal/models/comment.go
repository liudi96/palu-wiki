package models

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Content   string         `json:"content" gorm:"type:text;not null"`
	ArticleID uint           `json:"article_id"`
	UserID    uint           `json:"user_id"`
	ParentID  *uint          `json:"parent_id"` // 支持回复评论
	Status    string         `json:"status" gorm:"default:approved"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	User    User      `json:"user" gorm:"foreignKey:UserID"`
	Article Article   `json:"article" gorm:"foreignKey:ArticleID"`
	Parent  *Comment  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Replies []Comment `json:"replies,omitempty" gorm:"foreignKey:ParentID"`
}
