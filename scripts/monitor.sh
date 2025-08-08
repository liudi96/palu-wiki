#!/bin/bash

# 服务监控脚本
# 用法: ./monitor.sh

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}  Palu Wiki 服务监控面板${NC}"
echo -e "${BLUE}===========================================${NC}"
echo ""

# 检查Docker服务状态
echo -e "${YELLOW}Docker 服务状态:${NC}"
docker compose -f docker-compose.prod.yml ps
echo ""

# 检查服务健康状态
echo -e "${YELLOW}服务健康检查:${NC}"

# 后端健康检查
if curl -f -s http://localhost/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 后端服务正常${NC}"
else
    echo -e "${RED}✗ 后端服务异常${NC}"
fi

# 前端健康检查
if curl -f -s http://localhost > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 前端服务正常${NC}"
else
    echo -e "${RED}✗ 前端服务异常${NC}"
fi

# 数据库健康检查
if docker exec palu-wiki-db-prod pg_isready -U palu_user -d palu_wiki > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 数据库服务正常${NC}"
else
    echo -e "${RED}✗ 数据库服务异常${NC}"
fi

# Redis健康检查
if docker exec palu-wiki-redis-prod redis-cli ping | grep -q PONG; then
    echo -e "${GREEN}✓ Redis服务正常${NC}"
else
    echo -e "${RED}✗ Redis服务异常${NC}"
fi

echo ""

# 系统资源监控
echo -e "${YELLOW}系统资源使用情况:${NC}"
echo -e "${BLUE}CPU 使用率:${NC}"
top -bn1 | grep "Cpu(s)" | awk '{print $2 + $4"%"}'

echo -e "${BLUE}内存使用情况:${NC}"
free -h | grep -E "^Mem:" | awk '{print "已使用: " $3 " / 总计: " $2 " (" $3/$2*100 "%)"}'

echo -e "${BLUE}磁盘使用情况:${NC}"
df -h | grep -vE '^Filesystem|tmpfs|cdrom' | awk '{print $5 " " $1 " (" $4 " 剩余)"}'

echo ""

# Docker容器资源使用
echo -e "${YELLOW}容器资源使用情况:${NC}"
docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}" \
    palu-wiki-backend-prod palu-wiki-frontend-prod palu-wiki-db-prod palu-wiki-redis-prod palu-wiki-nginx-prod 2>/dev/null || echo "无法获取容器统计信息"

echo ""

# 最近日志
echo -e "${YELLOW}最近错误日志:${NC}"
docker compose -f docker-compose.prod.yml logs --tail=5 | grep -i error || echo "无错误日志"

echo ""

# 网络连通性测试
echo -e "${YELLOW}网络连通性测试:${NC}"
echo -e "${BLUE}本地连通性:${NC}"
curl -s -o /dev/null -w "HTTP状态码: %{http_code}, 响应时间: %{time_total}s\n" http://localhost/health

echo ""

# 数据库连接数
echo -e "${YELLOW}数据库连接信息:${NC}"
DB_CONNECTIONS=$(docker exec palu-wiki-db-prod psql -U palu_user -d palu_wiki -t -c "SELECT count(*) FROM pg_stat_activity WHERE state = 'active';" 2>/dev/null || echo "无法获取")
echo -e "${BLUE}活跃连接数: ${DB_CONNECTIONS}${NC}"

echo ""

# 备份状态
echo -e "${YELLOW}备份状态:${NC}"
if [ -d "/opt/palu-wiki/backups" ]; then
    BACKUP_COUNT=$(find /opt/palu-wiki/backups -name "*.sql.gz" | wc -l)
    LATEST_BACKUP=$(find /opt/palu-wiki/backups -name "*.sql.gz" -exec ls -lt {} \; | head -1 | awk '{print $6, $7, $8, $9}')
    echo -e "${BLUE}备份文件总数: ${BACKUP_COUNT}${NC}"
    echo -e "${BLUE}最新备份: ${LATEST_BACKUP}${NC}"
else
    echo -e "${YELLOW}备份目录不存在${NC}"
fi

echo ""
echo -e "${GREEN}监控完成！${NC}"
echo ""
echo -e "${YELLOW}常用维护命令:${NC}"
echo -e "${BLUE}重启所有服务: docker compose -f docker-compose.prod.yml restart${NC}"
echo -e "${BLUE}查看实时日志: docker compose -f docker-compose.prod.yml logs -f${NC}"
echo -e "${BLUE}执行数据备份: ./scripts/backup.sh${NC}"
echo -e "${BLUE}服务性能统计: docker stats${NC}"