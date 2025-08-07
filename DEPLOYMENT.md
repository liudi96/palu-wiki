# 🚀 Palu Wiki 腾讯云部署指南

这是一个完整的自动化部署指南，包含从服务器准备到应用上线的全套脚本和文档。

## 🚀 快速部署

### 准备工作
1. **腾讯云服务器** - Ubuntu 20.04+ 或 CentOS 8+，建议配置：2核4GB内存，40GB硬盘
2. **域名** - 已注册并解析到服务器IP的域名
3. **SSH访问** - 确保可以SSH登录到服务器

### 一键部署流程

#### 1. 环境检查
```bash
# 上传脚本到服务器
scp -r scripts/ user@your-server:/tmp/

# 登录服务器
ssh user@your-server

# 进入脚本目录
cd /tmp/scripts

# 运行环境检查
./server_check.sh
```

#### 2. 环境搭建
```bash
# 一键安装所有必需软件和配置
./server_setup.sh

# 注意：如果脚本添加了用户到docker组，需要重新登录
exit
ssh user@your-server
```

#### 3. 数据库初始化
```bash
cd /opt/palu-wiki
./db_init.sh

# 记录输出的数据库密码和管理员密码
```

#### 4. 安全加固
```bash
# 执行安全加固脚本
./security_hardening.sh

# 注意：这会修改SSH配置，确保您有SSH密钥访问权限
```

#### 5. SSL证书配置
```bash
# 配置域名和SSL证书
./ssl_setup.sh yourdomain.com admin@yourdomain.com
```

#### 6. 部署应用
```bash
# 部署Palu Wiki应用
./deploy.sh
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