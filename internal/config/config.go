package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	Server   ServerConfig
	AI       AIConfig
	JWT      JWTConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type ServerConfig struct {
	Port string
	Mode string
}

type AIConfig struct {
	SparkAppID     string
	SparkAPIKey    string
	SparkAPISecret string
	SparkDomain    string
	SparkBaseURL   string
}

type JWTConfig struct {
	Secret string
}

func LoadConfig() *Config {
	// 尝试加载.env文件
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: .env file not found, using environment variables or defaults: %v", err)
	} else {
		log.Printf("Successfully loaded .env file")
	}

	// 调试：输出关键环境变量
	log.Printf("ENV DEBUG: SPARK_APP_ID=%s", getEnv("SPARK_APP_ID", "NOT_SET"))
	log.Printf("ENV DEBUG: SPARK_API_KEY=%s", getEnv("SPARK_API_KEY", "NOT_SET")[:min(10, len(getEnv("SPARK_API_KEY", "NOT_SET")))]+"...")

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "palu_user"),
			Password: getEnv("DB_PASSWORD", "palu_password"),
			Name:     getEnv("DB_NAME", "palu_wiki"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		AI: AIConfig{
			SparkAppID:     getEnv("SPARK_APP_ID", ""),
			SparkAPIKey:    getEnv("SPARK_API_KEY", ""),
			SparkAPISecret: getEnv("SPARK_API_SECRET", ""),
			SparkDomain:    getEnv("SPARK_DOMAIN", "lite"),
			SparkBaseURL:   getEnv("SPARK_BASE_URL", "wss://spark-api.xf-yun.com/v1.1/chat"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "default_secret"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// min 返回两个数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
