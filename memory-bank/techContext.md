# 技术栈和架构设计

## 🏗️ 系统架构概览
**架构模式**: Go微服务 + Python AI引擎混合架构
**通信协议**: gRPC + HTTP RESTful API  
**部署方式**: 容器化微服务 + 服务编排
**设计理念**: 发挥各语言优势，Go处理高并发业务，Python处理AI计算

## 🔧 Go后端服务技术栈
- **Go版本**: 1.24.5
- **Web框架**: Gin v1.10.1  
- **数据库**: PostgreSQL
- **ORM**: GORM v1.30.1 + 官方Postgres驱动v1.6.0
- **缓存**: Redis v8.11.5
- **JWT认证**: golang-jwt/jwt/v5 v5.3.0
- **密码哈希**: golang.org/x/crypto/bcrypt
- **gRPC通信**: google.golang.org/grpc
- **服务发现**: etcd/consul (计划集成)
- **WebSocket**: gorilla/websocket v1.5.3
- **环境配置**: godotenv v1.5.1

## 🤖 Python AI服务技术栈
- **Python版本**: 3.11+
- **Web框架**: FastAPI 0.104+
- **AI框架**: LangChain 0.1+
- **向量数据库**: ChromaDB 0.4+
- **深度学习**: PyTorch 2.1+ / Transformers 4.36+
- **HTTP客户端**: httpx 0.25+
- **异步框架**: asyncio + uvloop
- **gRPC服务**: grpcio 1.59+
- **数据处理**: pandas 2.1+ / numpy 1.24+

## 🌐 AI模型集成
- **主力模型**: DeepSeek (优先级1)
- **备用模型**: 通义千问 (优先级2)  
- **兜底模型**: 讯飞星火 (优先级3)
- **本地模型**: Ollama集成 (计划)
- **向量模型**: text-embedding-3-small
- **图像生成**: DALL-E 3 / Midjourney (计划)

## 📱 前端技术栈
- **框架**: Next.js 14.x (App Router)
- **语言**: TypeScript 5.x
- **样式**: TailwindCSS 3.3.x
- **图标**: Heroicons 2.0
- **HTTP客户端**: Axios 1.6.x
- **状态管理**: React内置 + js-cookie 3.0.x
- **通知**: react-hot-toast 2.5.x
- **开发工具**: ESLint 8.x
- **UI组件**: Headless UI (计划集成)

## 🚀 基础设施和DevOps
- **容器化**: Docker + Docker Compose
- **服务编排**: Kubernetes (生产环境计划)
- **反向代理**: Nginx + 负载均衡
- **文件存储**: 本地存储 (uploads/) + CDN (计划)
- **消息队列**: NATS (计划集成)
- **监控系统**: Prometheus + Grafana
- **链路追踪**: Jaeger (计划集成)
- **日志聚合**: ELK Stack (计划集成)

## 🔌 端口和网络配置
- **Go API服务**: 8080
- **Python AI服务**: 8081  
- **前端开发**: 3001
- **PostgreSQL**: 5432
- **Redis**: 6379
- **ChromaDB**: 8000
- **gRPC通信**: 50051
- **网络代理**: 127.0.0.1:15236 (开发环境)

## 📊 数据流架构
```
前端 → Nginx → Go API → gRPC → Python AI服务
                ↓              ↓
            PostgreSQL      ChromaDB
                ↓
             Redis Cache
```

## 🎯 技术选型理由
- **Go**: 高并发API处理、系统服务、快速响应
- **Python**: AI计算、机器学习、丰富生态
- **gRPC**: 高效的服务间通信、类型安全
- **ChromaDB**: 专业向量存储、AI检索优化
- **LangChain**: AI应用开发框架、快速集成