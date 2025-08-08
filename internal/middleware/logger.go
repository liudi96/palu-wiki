package middleware

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	// 检查是否为生产环境，使用JSON格式日志
	logFormat := os.Getenv("LOG_FORMAT")
	if logFormat == "json" {
		return JSONLogger()
	}

	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// JSONLogger 生产环境JSON格式日志
func JSONLogger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			logData := map[string]interface{}{
				"timestamp":  param.TimeStamp.Format(time.RFC3339),
				"level":      "INFO",
				"method":     param.Method,
				"path":       param.Path,
				"status":     param.StatusCode,
				"latency_ms": float64(param.Latency.Nanoseconds()) / 1e6,
				"client_ip":  param.ClientIP,
				"user_agent": param.Request.UserAgent(),
				"proto":      param.Request.Proto,
				"referer":    param.Request.Referer(),
			}

			if param.ErrorMessage != "" {
				logData["error"] = param.ErrorMessage
				logData["level"] = "ERROR"
			}

			jsonBytes, _ := json.Marshal(logData)
			return string(jsonBytes) + "\n"
		},
		Output:    gin.DefaultWriter,
		SkipPaths: []string{"/health"}, // 跳过健康检查日志
	})
}

// ErrorLogger 错误日志中间件
func ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 如果有错误，记录详细错误信息
		if len(c.Errors) > 0 {
			errorData := map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"level":     "ERROR",
				"method":    c.Request.Method,
				"path":      c.Request.URL.Path,
				"client_ip": c.ClientIP(),
				"errors":    c.Errors,
			}

			if os.Getenv("LOG_FORMAT") == "json" {
				jsonBytes, _ := json.Marshal(errorData)
				fmt.Fprintf(gin.DefaultErrorWriter, "%s\n", string(jsonBytes))
			} else {
				fmt.Fprintf(gin.DefaultErrorWriter, "[ERROR] %s %s %s - %v\n",
					time.Now().Format(time.RFC3339),
					c.Request.Method,
					c.Request.URL.Path,
					c.Errors)
			}
		}
	}
}
