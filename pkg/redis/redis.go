package redis

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"

	"palu-wiki/internal/config"
)

// 全局Redis客户端实例
var client *redis.Client

func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试连接
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	// 设置全局客户端实例
	client = rdb

	log.Println("Redis connected successfully")
	return rdb, nil
}

// GetClient 获取全局Redis客户端实例
func GetClient() *redis.Client {
	return client
}
