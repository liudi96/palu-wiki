# 🏗️ Palu Wiki 混合架构系统设计

## 🎯 架构设计理念

### 核心原则
- **技术适配性**: Go处理高并发业务，Python专注AI计算
- **服务解耦**: 松耦合微服务架构，独立部署和扩展
- **高可用性**: 服务冗余、故障隔离、优雅降级
- **可扩展性**: 水平扩展、弹性伸缩、资源优化
- **可观测性**: 全链路监控、实时告警、性能分析

## 🔧 系统架构图

```
                    🌐 用户请求
                         ↓
                   ┌─────────────┐
                   │    Nginx    │ ← 负载均衡 + SSL
                   │  反向代理    │
                   └─────────────┘
                         ↓
                   ┌─────────────┐
                   │  API网关     │ ← 路由 + 认证 + 限流
                   │ (Kong/Istio) │
                   └─────────────┘
                    ↙️     ↓     ↘️
        ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
        │   用户服务   │ │   文章服务   │ │  文件服务    │
        │(Go Service) │ │(Go Service) │ │(Go Service) │
        └─────────────┘ └─────────────┘ └─────────────┘
               ↓               ↓               ↓
        ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
        │ PostgreSQL  │ │ PostgreSQL  │ │ MinIO/S3    │
        │  用户数据   │ │  文章数据   │ │  文件存储   │
        └─────────────┘ └─────────────┘ └─────────────┘
                         ↓
                   📡 gRPC 通信
                         ↓
                   ┌─────────────┐
                   │  AI服务集群  │
                   │(Python/Fast)│ ← AI计算 + 向量搜索
                   └─────────────┘
                    ↙️     ↓     ↘️
        ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
        │ LangChain   │ │  ChromaDB   │ │  AI模型     │
        │   Agent     │ │  向量存储   │ │ DeepSeek等  │
        └─────────────┘ └─────────────┘ └─────────────┘
                         ↓
                   ┌─────────────┐
                   │  Redis集群  │ ← 缓存 + 会话
                   │    缓存     │
                   └─────────────┘
                         ↓
                   ┌─────────────┐
                   │  NATS队列   │ ← 异步任务
                   │  消息系统   │
                   └─────────────┘
```

## 🔄 数据流设计

### 1. 用户认证流程
```
用户登录 → API网关 → 用户服务 → PostgreSQL → JWT Token → Redis缓存
```

### 2. 文章浏览流程  
```
文章请求 → API网关 → 文章服务 → PostgreSQL → Redis缓存 → 响应用户
```

### 3. AI生成流程
```
AI请求 → 文章服务 → gRPC → Python AI服务 → LangChain → 外部AI API
                                    ↓
                         ChromaDB向量检索 → 生成结果返回
```

### 4. 智能搜索流程
```
搜索请求 → 文章服务 → gRPC → Python AI服务 → 向量化查询
                                        ↓
                              ChromaDB相似度搜索 → 结果排序返回
```

## 🔌 服务通信设计

### gRPC接口定义
```protobuf
// ai_service.proto
syntax = "proto3";
package ai;

service AIService {
  // 内容生成
  rpc GenerateContent(ContentRequest) returns (ContentResponse);
  
  // 文章生成
  rpc GenerateArticle(ArticleRequest) returns (ArticleResponse);
  
  // 向量搜索
  rpc SearchSimilar(SearchRequest) returns (SearchResponse);
  
  // 流式对话
  rpc StreamChat(stream ChatRequest) returns (stream ChatResponse);
  
  // 模型状态
  rpc GetModelStats(StatsRequest) returns (StatsResponse);
}

message ContentRequest {
  string prompt = 1;
  string topic = 2;
  int32 max_length = 3;
  string user_id = 4;
}

message ContentResponse {
  string content = 1;
  string summary = 2;
  repeated string tags = 3;
  string model_used = 4;
  int32 tokens_used = 5;
}
```

### HTTP API设计
```yaml
# Go微服务HTTP接口
/api/v1/users/*      # 用户服务
/api/v1/articles/*   # 文章服务  
/api/v1/files/*      # 文件服务
/api/v1/categories/* # 分类服务
/api/v1/search/*     # 搜索服务

# Python AI服务HTTP接口
/ai/v1/generate      # 内容生成
/ai/v1/chat         # 对话接口
/ai/v1/search       # 向量搜索
/ai/v1/models       # 模型管理
```

## 💾 数据存储设计

### PostgreSQL数据库分片
```sql
-- 用户服务数据库
Database: palu_users
- users (用户基础信息)
- user_profiles (用户扩展资料)  
- user_sessions (会话管理)

-- 文章服务数据库  
Database: palu_content
- articles (文章主体)
- categories (分类信息)
- article_stats (统计数据)

-- AI服务元数据库
Database: palu_ai
- ai_generations (生成记录)
- model_usage_stats (使用统计)
- chat_histories (对话历史)
```

### ChromaDB向量存储设计
```python
# 集合设计
collections = {
    "articles": {
        "embedding_model": "text-embedding-3-small",
        "metadata": ["title", "category", "author", "tags"],
        "documents": "article_content_chunks"
    },
    "user_queries": {
        "embedding_model": "text-embedding-3-small", 
        "metadata": ["user_id", "query_type", "timestamp"],
        "documents": "processed_queries"
    }
}
```

### Redis缓存策略
```redis
# 缓存key设计
user:session:{user_id}     # 用户会话 (TTL: 24h)
article:cache:{article_id} # 文章缓存 (TTL: 1h)
search:cache:{query_hash}  # 搜索缓存 (TTL: 30m)
ai:quota:{provider}        # AI配额缓存 (TTL: 1h)
```

## 🔒 安全架构设计

### 认证授权体系
```
JWT Token → API网关验证 → 服务间传递 → 细粒度权限控制
```

### 安全中间件链
```go
// 中间件调用顺序
CORS → Rate Limiting → JWT Auth → Permission Check → Business Logic
```

### API安全策略
- **输入验证**: 参数校验、SQL注入防护、XSS过滤
- **访问控制**: 基于角色的权限管理 (RBAC)
- **传输安全**: HTTPS强制、证书管理、密钥轮换
- **监控告警**: 异常行为检测、攻击模式识别

## 📊 监控和可观测性

### 监控指标体系
```yaml
# 业务指标
- 用户注册/登录成功率
- 文章创建/浏览量
- AI生成成功率和响应时间
- 搜索查询量和相关性

# 技术指标  
- 服务响应时间 (P95, P99)
- 错误率和可用性 (99.9%)
- 资源使用率 (CPU, Memory, Disk)
- 数据库连接池和查询性能
```

### 告警规则设计
```yaml
# 关键告警
- 服务不可用 > 30秒
- API响应时间 > 2秒  
- 错误率 > 5%
- AI服务响应时间 > 10秒
- 数据库连接数 > 80%
```

## 🚀 部署架构设计

### 容器化部署
```yaml
# docker-compose.yml 服务配置
services:
  # Go微服务集群
  user-service:
    replicas: 2
    resources: {cpu: 0.5, memory: 512Mi}
    
  article-service: 
    replicas: 3
    resources: {cpu: 1, memory: 1Gi}
    
  # Python AI服务集群
  ai-service:
    replicas: 2
    resources: {cpu: 2, memory: 4Gi, gpu: 1}
    
  # 数据存储服务
  postgresql:
    replicas: 1 (主从复制)
    resources: {cpu: 2, memory: 4Gi, storage: 100Gi}
    
  chromadb:
    replicas: 1
    resources: {cpu: 1, memory: 2Gi, storage: 50Gi}
```

### Kubernetes部署配置
```yaml
# k8s部署策略
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ai-service
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  template:
    spec:
      containers:
      - name: ai-service
        image: palu-wiki/ai-service:latest
        resources:
          requests:
            cpu: 1000m
            memory: 2Gi
          limits:
            cpu: 2000m  
            memory: 4Gi
```

## 🔄 CI/CD流水线设计

### 构建流程
```yaml
# .github/workflows/deploy.yml
stages:
  - code_check: # 代码检查
    - go fmt, go vet, golangci-lint
    - python black, flake8, mypy
    - security scan with gosec/bandit
    
  - unit_test: # 单元测试
    - go test with coverage
    - pytest with coverage
    - test report generation
    
  - build: # 构建镜像
    - docker build with multi-stage
    - image security scan
    - push to registry
    
  - deploy: # 部署应用
    - helm upgrade with rollback
    - health check validation
    - notification to team
```

## 🎯 性能优化策略

### 服务性能优化
- **Go服务**: 连接池优化、内存复用、并发控制
- **Python AI**: 模型预加载、批处理、异步处理
- **数据库**: 索引优化、查询缓存、读写分离
- **缓存策略**: 多层缓存、缓存预热、失效策略

### 网络性能优化
- **CDN加速**: 静态资源分发、边缘缓存
- **负载均衡**: 智能路由、健康检查、会话保持
- **压缩传输**: Gzip压缩、Brotli算法
- **Keep-Alive**: 连接复用、管道化请求

## 🔧 运维自动化

### 自动化运维脚本
```bash
# 运维工具集
scripts/
├── deploy.sh      # 一键部署
├── backup.sh      # 数据备份
├── monitor.sh     # 健康检查
├── scale.sh       # 弹性扩容
└── rollback.sh    # 快速回滚
```

### 故障自愈机制
- **健康检查**: 服务状态实时监控
- **自动重启**: 故障服务自动恢复  
- **流量切换**: 故障节点自动摘除
- **数据恢复**: 自动备份和恢复机制

---

## 💡 架构演进规划

### 短期优化 (1-2月)
- 完善监控告警体系
- 优化AI服务性能
- 加强安全防护

### 中期演进 (3-6月)  
- 引入服务网格 (Istio)
- 实现多云部署
- 建设数据湖架构

### 长期规划 (6-12月)
- AI模型训练平台
- 边缘计算节点
- 国际化多区域部署

**这个架构设计支持从小规模起步到大规模扩展的完整演进路径！**