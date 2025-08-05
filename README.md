# 🎮 幻兽帕鲁攻略网站

基于 Go + Next.js 构建的现代化游戏攻略网站，集成 AI 内容生成、用户管理和社区功能。

## ✨ 功能特性

### 🎯 已完成功能 (MVP)
- **🔐 用户系统**: 注册、登录、JWT认证、角色权限管理
- **📝 文章管理**: CRUD操作、状态管理、分类系统
- **🤖 AI集成**: 讯飞星火大模型内容生成
- **🔍 搜索功能**: 关键词搜索、高级过滤、分页
- **👑 管理后台**: 仪表板、用户管理、文章审核
- **📱 响应式设计**: 支持桌面端和移动端

### 🚧 计划功能
- **💬 社区互动**: 评论系统、用户等级、积分体系
- **🛠️ 实用工具**: 帕鲁计算器、服务器监控
- **💰 商业化**: 会员订阅、付费内容

## 🏗️ 技术架构

### 后端技术栈
- **语言**: Go 1.21+
- **框架**: Gin Web Framework
- **数据库**: PostgreSQL + GORM ORM
- **缓存**: Redis
- **认证**: JWT + bcrypt
- **AI**: 讯飞星火大模型 API

### 前端技术栈
- **框架**: Next.js 14 (App Router)
- **语言**: TypeScript
- **样式**: Tailwind CSS
- **图标**: Heroicons
- **HTTP客户端**: Axios

## 🚀 快速开始

### 1. 环境要求
- Go 1.21+
- Node.js 18+
- PostgreSQL 13+
- Redis (可选)

### 2. 一键启动
```bash
# 克隆项目
git clone <repository-url>
cd palu-wiki

# 运行启动脚本
./start.sh
```

启动脚本会自动：
- 检查依赖环境
- 启动数据库服务
- 编译并运行后端 (端口: 8080)
- 安装前端依赖并启动 (端口: 3001)
- 创建管理员账户

### 3. 访问应用
- **后端API**: http://localhost:8080
- **管理后台**: http://localhost:3001
- **API文档**: http://localhost:8080/health

## API 接口

### 健康检查
```
GET /health
```

### 文章相关
```
GET /api/v1/articles          # 获取文章列表
GET /api/v1/articles/:id      # 获取文章详情
GET /api/v1/articles/search   # 搜索文章
POST /api/v1/articles         # 创建文章 (需认证)
```

### 分类相关
```
GET /api/v1/categories              # 获取分类列表
GET /api/v1/categories/:id/articles # 获取分类下的文章
```

## 项目结构

```
├── cmd/server/           # 主程序入口
├── internal/
│   ├── config/          # 配置管理
│   ├── handlers/        # HTTP处理器
│   ├── middleware/      # 中间件
│   ├── models/          # 数据模型
│   └── services/        # 业务逻辑
├── pkg/
│   ├── database/        # 数据库连接
│   ├── redis/           # Redis连接
│   └── utils/           # 工具函数
├── docs/               # 文档
└── scripts/            # 脚本文件
```

## 开发命令

```bash
make dev     # 开发模式运行
make build   # 构建项目
make test    # 运行测试
make fmt     # 格式化代码
make clean   # 清理构建文件
```

## 部署

TODO: 添加Docker和生产环境部署说明

## 贡献

欢迎提交Pull Request和Issue！