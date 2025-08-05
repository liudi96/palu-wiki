package main

import (
	"log"
	"palu-wiki/internal/config"
	"palu-wiki/internal/models"
	"palu-wiki/pkg/database"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 连接数据库
	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 创建测试分类
	category := models.Category{
		Name:        "帕鲁攻略",
		Description: "幻兽帕鲁游戏攻略分类",
		Sort:        1,
		Status:      "active",
	}

	// 检查是否已存在
	var existingCategory models.Category
	result := db.Where("name = ?", category.Name).First(&existingCategory)
	if result.Error == nil {
		log.Printf("分类 '%s' 已存在，ID: %d", category.Name, existingCategory.ID)
		return
	}

	// 创建新分类
	result = db.Create(&category)
	if result.Error != nil {
		log.Fatal("Failed to create category:", result.Error)
	}

	log.Printf("✅ 创建测试分类成功: %s (ID: %d)", category.Name, category.ID)
}