# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

用中文回答我
每次都用审视的目光，仔细看我输入的潜在问题，你要指出我的问题，并给出明显在我思考框架之外的建议。
如果你觉得我说的太离谱了，你就骂回来，帮我瞬间清醒。
开发必须遵循TDO(测试驱动开发)的方法论。
外部大脑(memory-bank文件夹)
## 🚀 快速开发

### 一键启动
```bash
./start.sh              # 启动完整应用（后端+前端+数据库）
```

### 单独启动
```bash
# 后端开发
make dev                 # go run cmd/server/main.go (端口8080)

# 前端开发  
cd frontend && npm run dev  # 开发服务器 (端口3001)
```

### 代码检查
```bash
make fmt && make lint    # Go代码格式化和检查
cd frontend && npm run lint  # TypeScript检查
```
## 云服务器


## 🌐 网络配置
外网走代理：127.0.0.1:15236
本地直连

## 🏗️ 核心架构

### 关键文件位置
- **入口**: `cmd/server/main.go` - 程序启动点
- **路由**: `internal/handlers/routes.go:11` - API路由配置
- **数据库**: `pkg/database/postgres.go:15` - 数据库连接
- **AI集成**: `pkg/ai/spark.go` - 星火AI客户端
- **前端API**: `frontend/src/lib/api.ts` - 统一API调用
- **前端布局**: `frontend/src/app/layout.tsx` - 根组件

### 数据流
```
前端 → api.ts → Gin Router → 中间件 → Handler → GORM → PostgreSQL
```

### 权限体系
- **普通用户**: 读取文章、注册登录
- **编辑用户**: 创建和编辑自己的文章  
- **管理员**: 管理所有用户和文章，访问管理后台(`/admin`)

## ✅ 测试要求

功能开发完成后必须进行端到端测试：
```bash
# API测试
curl http://localhost:8080/health

# 使用现有测试脚本
python3 create_test_data.py
```

## 🔧 常见问题

- 端口被占用：检查8080(后端)和3001(前端)端口
- 数据库连接失败：确认PostgreSQL服务运行状态
- 网络请求超时：检查代理设置127.0.0.1:15236
- AI生成失败：检查星火API密钥配置

### 讯飞星火AI配置
SPARK_APP_ID=8ddf9157
SPARK_API_KEY=59b8dde668da384b876e5f3640ce9197
SPARK_API_SECRET=MzliZmY5ODBjYzQyODBjMjQ0MTY1Yzdi
SPARK_DOMAIN=lite
SPARK_BASE_URL=wss://spark-api.xf-yun.com/v1.1/chat


---