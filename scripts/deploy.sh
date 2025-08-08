#!/bin/bash

# 生产环境部署脚本
# 用法: ./deploy.sh [版本号]

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
PROJECT_DIR="/opt/palu-wiki"
BACKUP_DIR="$PROJECT_DIR/backups"
VERSION=${1:-latest}

echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}  Palu Wiki 生产环境部署脚本${NC}"
echo -e "${BLUE}===========================================${NC}"
echo -e "${YELLOW}版本: $VERSION${NC}"
echo -e "${YELLOW}项目目录: $PROJECT_DIR${NC}"
echo ""

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}错误: Docker未运行${NC}"
    exit 1
fi

# 创建必要目录
echo -e "${YELLOW}创建项目目录...${NC}"
sudo mkdir -p $PROJECT_DIR/{postgres_data,redis_data,uploads,logs,backups}

# 进入项目目录
cd $PROJECT_DIR

# 更新代码
echo -e "${YELLOW}更新代码...${NC}"
if [ -d ".git" ]; then
    git fetch origin
    git reset --hard origin/main
else
    echo -e "${RED}警告: 不是Git仓库，跳过代码更新${NC}"
fi

# 检查环境变量文件
if [ ! -f ".env" ]; then
    echo -e "${RED}错误: .env文件不存在${NC}"
    echo -e "${YELLOW}请复制 .env.production 为 .env 并配置环境变量${NC}"
    exit 1
fi

# 备份数据库（如果服务正在运行）
if docker ps | grep -q palu-wiki-db-prod; then
    echo -e "${YELLOW}备份数据库...${NC}"
    ./scripts/backup.sh || echo -e "${YELLOW}备份失败，继续部署${NC}"
fi

# 停止现有服务
echo -e "${YELLOW}停止现有服务...${NC}"
docker compose -f docker-compose.prod.yml down --remove-orphans

# 清理未使用的Docker资源
echo -e "${YELLOW}清理Docker资源...${NC}"
docker system prune -f --volumes

# 构建新镜像
echo -e "${YELLOW}构建应用镜像...${NC}"
docker compose -f docker-compose.prod.yml build --no-cache

# 启动服务
echo -e "${YELLOW}启动服务...${NC}"
docker compose -f docker-compose.prod.yml up -d

# 等待服务启动
echo -e "${YELLOW}等待服务启动...${NC}"
sleep 30

# 健康检查
echo -e "${YELLOW}执行健康检查...${NC}"
RETRY_COUNT=0
MAX_RETRIES=10

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f http://localhost/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ 后端服务健康检查通过${NC}"
        break
    else
        echo -e "${YELLOW}等待后端服务启动... ($((RETRY_COUNT + 1))/$MAX_RETRIES)${NC}"
        sleep 10
        RETRY_COUNT=$((RETRY_COUNT + 1))
    fi
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo -e "${RED}✗ 后端服务健康检查失败${NC}"
    echo -e "${YELLOW}查看服务日志:${NC}"
    docker compose -f docker-compose.prod.yml logs --tail=50
    exit 1
fi

# 检查前端服务
if curl -f http://localhost > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 前端服务健康检查通过${NC}"
else
    echo -e "${RED}✗ 前端服务健康检查失败${NC}"
fi

# 显示服务状态
echo -e "${YELLOW}服务状态:${NC}"
docker compose -f docker-compose.prod.yml ps

# 显示日志
echo -e "${YELLOW}最近日志:${NC}"
docker compose -f docker-compose.prod.yml logs --tail=20

echo ""
echo -e "${GREEN}===========================================${NC}"
echo -e "${GREEN}  🎉 部署完成! ${NC}"
echo -e "${GREEN}===========================================${NC}"
echo -e "${GREEN}网站地址: http://localhost${NC}"
echo -e "${GREEN}健康检查: http://localhost/health${NC}"
echo ""
echo -e "${YELLOW}常用命令:${NC}"
echo -e "${BLUE}查看日志: docker compose -f docker-compose.prod.yml logs -f${NC}"
echo -e "${BLUE}重启服务: docker compose -f docker-compose.prod.yml restart${NC}"
echo -e "${BLUE}停止服务: docker compose -f docker-compose.prod.yml down${NC}"
echo ""