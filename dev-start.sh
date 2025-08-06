#!/bin/bash

# 开发环境启动脚本 - 本地测试用
echo "🛠️ 启动开发环境..."

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 检查Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker未安装${NC}"
    exit 1
fi

# 停止旧服务
echo -e "${YELLOW}🛑 停止旧的开发环境...${NC}"
docker-compose -f docker-compose.dev.yml down

# 创建目录
mkdir -p uploads

# 启动开发环境
echo -e "${GREEN}🚀 启动开发环境...${NC}"
docker-compose -f docker-compose.dev.yml up --build

echo -e "${GREEN}✅ 开发环境已启动${NC}"
echo "访问地址:"
echo "  后端: http://localhost:8080"
echo "  前端: http://localhost:3001"