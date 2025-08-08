# 技术栈和版本

## 后端技术栈
- **Go版本**: 1.24.5
- **Web框架**: Gin v1.10.1  
- **数据库**: PostgreSQL
- **ORM**: GORM v1.30.1 + 官方Postgres驱动v1.6.0
- **缓存**: Redis v8.11.5
- **JWT认证**: golang-jwt/jwt/v5 v5.3.0
- **密码哈希**: golang.org/x/crypto/bcrypt
- **AI集成**: 讯飞星火AI (iflytek/spark-ai-go)
- **WebSocket**: gorilla/websocket v1.5.3
- **环境配置**: godotenv v1.5.1

## 前端技术栈
- **框架**: Next.js 14.x (App Router)
- **语言**: TypeScript 5.x
- **样式**: TailwindCSS 3.3.x
- **图标**: Heroicons 2.0
- **HTTP客户端**: Axios 1.6.x
- **状态管理**: React内置 + js-cookie 3.0.x
- **通知**: react-hot-toast 2.5.x
- **开发工具**: ESLint 8.x

## 基础设施
- **容器**: Docker + Docker Compose
- **反向代理**: Nginx
- **文件上传**: 本地存储 (uploads/)
- **网络代理**: 127.0.0.1:15236 (开发环境)
- **端口配置**: 
  - 后端: 8080
  - 前端: 3001
  - PostgreSQL: 5432
  - Redis: 6379