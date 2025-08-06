# 🚀 帕鲁攻略网站部署指南

这是一个完整的Docker化部署指南，让你学习从开发到生产的完整流程。

## 📋 部署前准备

### 1. 环境要求
- **服务器**: Linux系统（推荐Ubuntu 20.04+）
- **Docker**: 20.10+
- **Docker Compose**: 2.0+
- **内存**: 最少2GB，推荐4GB
- **存储**: 最少20GB可用空间

### 2. 腾讯云服务器配置
```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 安装Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 将用户添加到docker组
sudo usermod -aG docker $USER
```

### 3. 防火墙配置
```bash
# 开放必要端口
sudo ufw allow 22    # SSH
sudo ufw allow 80    # HTTP
sudo ufw allow 443   # HTTPS
sudo ufw allow 8080  # 后端API（可选，用于调试）
sudo ufw allow 3001  # 前端服务（可选，用于调试）
sudo ufw enable
```

## 🛠️ 本地测试部署

### 1. 环境配置
```bash
# 复制环境变量文件
cp .env.production .env

# 编辑配置文件，填写真实的配置
nano .env
```

**重要**: 必须填写星火AI的配置信息：
- `SPARK_APP_ID`: 你的App ID
- `SPARK_API_KEY`: 你的API Key  
- `SPARK_API_SECRET`: 你的API Secret

### 2. 开发环境测试
```bash
# 启动开发环境（用于本地测试）
./dev-start.sh

# 或者手动启动
docker-compose -f docker-compose.dev.yml up --build
```

### 3. 生产环境测试
```bash
# 启动生产环境
./deploy.sh

# 或者手动启动
docker-compose up --build -d
```

### 4. 验证部署
```bash
# 检查服务状态
docker-compose ps

# 检查健康状态
curl http://localhost:8080/health

# 查看日志
docker-compose logs -f
```

## 🌐 腾讯云生产部署

### 1. 代码部署
```bash
# 在服务器上克隆代码
git clone <你的仓库地址>
cd palu-wiki

# 配置环境变量
cp .env.production .env
nano .env  # 填写真实配置
```

### 2. 一键部署
```bash
# 运行部署脚本
./deploy.sh
```

### 3. 域名配置（可选）
如果你有域名，可以配置Nginx反向代理：

```bash
# 修改docker-compose.yml中的nginx配置
# 启用nginx服务
docker-compose up -d nginx
```

### 4. SSL证书配置（可选）
```bash
# 使用Let's Encrypt免费证书
sudo apt install certbot
certbot --nginx -d your-domain.com
```

## 📊 服务监控

### 1. 查看服务状态
```bash
# 查看所有容器状态
docker-compose ps

# 查看资源使用情况
docker stats

# 查看特定服务日志
docker-compose logs -f backend
docker-compose logs -f frontend
```

### 2. 常用运维命令
```bash
# 重启服务
docker-compose restart

# 重新构建服务
docker-compose up --build -d

# 停止服务
docker-compose down

# 清理未使用的镜像
docker image prune -a
```

### 3. 数据备份
```bash
# 备份数据库
docker-compose exec postgres pg_dump -U palu_user palu_wiki > backup.sql

# 备份上传文件
tar -czf uploads_backup.tar.gz uploads/

# 恢复数据库
cat backup.sql | docker-compose exec -T postgres psql -U palu_user -d palu_wiki
```

## 🔧 常见问题排查

### 1. 服务启动失败
```bash
# 查看详细错误日志
docker-compose logs backend
docker-compose logs frontend

# 检查端口占用
netstat -tlnp | grep :8080
netstat -tlnp | grep :3001
```

### 2. 数据库连接失败
```bash
# 检查PostgreSQL状态
docker-compose exec postgres pg_isready -U palu_user

# 查看数据库日志
docker-compose logs postgres

# 手动连接测试
docker-compose exec postgres psql -U palu_user -d palu_wiki
```

### 3. AI功能不工作
- 检查.env文件中的星火AI配置是否正确
- 确认API密钥是否有效
- 查看后端日志是否有AI相关错误

### 4. 前端访问404
- 检查前端构建是否成功
- 确认Nginx配置是否正确
- 检查前端服务是否正常运行

## 🎯 学习收获总结

通过这次部署，你将学会：

### 技术技能
- **Docker容器化**: 掌握Dockerfile编写和多阶段构建
- **服务编排**: 理解docker-compose的使用
- **反向代理**: 学会Nginx配置和负载均衡
- **Linux运维**: 服务器管理和故障排查
- **云服务使用**: 腾讯云CVM的实际应用

### 实践经验
- **完整部署流程**: 从开发到生产的完整链路
- **自动化脚本**: 编写部署和运维脚本
- **监控运维**: 服务健康检查和日志管理
- **问题解决**: 独立排查和解决部署问题

### 架构理解
- **微服务架构**: 前后端分离的实际应用
- **数据持久化**: Docker卷的使用
- **网络通信**: 容器间的网络配置
- **安全考虑**: 基础的生产环境安全配置

## 📞 后续优化方向

1. **监控系统**: 集成Prometheus + Grafana
2. **日志系统**: 使用ELK Stack
3. **CI/CD**: 集成GitHub Actions
4. **负载均衡**: 多实例部署
5. **缓存优化**: Redis集群配置

这是一个很好的实战项目，能让你快速掌握现代化的Web应用部署技能！