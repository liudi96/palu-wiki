# 架构和设计模式

## 整体架构
采用**分层微服务架构**，前后端分离设计：

```
前端层 (Next.js + TypeScript)
    ↓ HTTP/REST API
中间件层 (Gin Middleware)
    ↓
路由层 (Gin Router)
    ↓  
处理器层 (Handlers)
    ↓
业务逻辑层 (Services - 待完善)
    ↓
数据访问层 (GORM)
    ↓
数据存储层 (PostgreSQL + Redis)
```

## 核心设计模式

### 1. MVC模式变种
- **Models**: `internal/models/` - 数据结构定义
- **Controllers**: `internal/handlers/` - 请求处理逻辑
- **Views**: `frontend/src/` - 用户界面

### 2. 中间件模式
- **认证中间件**: JWT验证 (`middleware/auth.go`)
- **权限中间件**: 角色权限控制
- **日志中间件**: 请求日志记录
- **CORS中间件**: 跨域处理

### 3. 工厂模式
- Handler实例创建: `NewArticleHandler()`, `NewAuthHandler()`
- 数据库连接: `database/postgres.go`

### 4. 依赖注入
- 通过构造函数注入数据库连接和配置

## 权限控制体系

### 角色层次
1. **Guest** - 匿名用户（只读权限）
2. **User** - 注册用户（创建文章）
3. **Editor** - 编辑用户（编辑自己的文章）
4. **Admin** - 管理员（全部权限）

### 权限控制点
- **路由级**: 中间件控制访问
- **处理器级**: 业务逻辑验证
- **数据库级**: 所有者关系验证

## 数据流模式

### API请求流
```
Client Request → CORS → Logger → JWT Auth → Role Check → Handler → GORM → Database
```

### AI内容生成流
```
用户请求 → 权限验证 → 星火AI API → 内容处理 → 数据库存储 → 返回结果
```

## 错误处理模式
- 统一错误响应格式
- 分层错误处理（中间件 → Handler → Service）
- 日志记录和监控

## 文件组织模式
```
cmd/          - 应用入口点
internal/     - 私有应用代码
  ├── config/   - 配置管理
  ├── handlers/ - HTTP处理器
  ├── middleware/ - 中间件
  └── models/   - 数据模型
pkg/          - 可重用包
  ├── ai/       - AI集成
  ├── database/ - 数据库连接
  └── utils/    - 工具函数
```