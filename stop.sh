#!/bin/bash

# 幻兽帕鲁攻略网站停止脚本
echo "🛑 停止幻兽帕鲁攻略网站服务..."

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 停止后端服务
echo -e "${YELLOW}停止后端服务...${NC}"
pkill -f "palu-wiki" 2>/dev/null
pkill -f "bin/palu-wiki" 2>/dev/null

# 停止前端服务
echo -e "${YELLOW}停止前端服务...${NC}"
pkill -f "next dev" 2>/dev/null
pkill -f "npm run dev" 2>/dev/null

# 等待进程停止
sleep 2

# 检查是否还有服务在运行
if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null ; then
    echo -e "${RED}⚠️  端口 8080 仍被占用，强制停止...${NC}"
    kill -9 $(lsof -ti :8080) 2>/dev/null
fi

if lsof -Pi :3001 -sTCP:LISTEN -t >/dev/null ; then
    echo -e "${RED}⚠️  端口 3001 仍被占用，强制停止...${NC}"
    kill -9 $(lsof -ti :3001) 2>/dev/null
fi

echo -e "${GREEN}✅ 所有服务已停止${NC}"