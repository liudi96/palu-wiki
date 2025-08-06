#!/bin/bash

# 生产环境部署脚本 - 学习自动化部署
echo "🚀 开始部署帕鲁攻略网站到生产环境..."

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 错误处理
set -e

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker未安装，请先安装Docker${NC}"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        echo -e "${RED}❌ Docker Compose未安装，请先安装Docker Compose${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Docker环境检查通过${NC}"
}

# 检查环境变量文件
check_env() {
    if [ ! -f ".env" ]; then
        if [ -f ".env.production" ]; then
            echo -e "${YELLOW}⚠️  .env文件不存在，从.env.production复制...${NC}"
            cp .env.production .env
            echo -e "${YELLOW}⚠️  请编辑.env文件，填写正确的配置信息！${NC}"
            echo -e "${YELLOW}⚠️  特别是SPARK_APP_ID, SPARK_API_KEY, SPARK_API_SECRET${NC}"
            read -p "配置完成后按回车继续..." -r
        else
            echo -e "${RED}❌ 环境配置文件缺失${NC}"
            exit 1
        fi
    fi
    echo -e "${GREEN}✅ 环境配置检查通过${NC}"
}

# 创建必要的目录
create_dirs() {
    echo -e "${GREEN}📁 创建必要的目录...${NC}"
    mkdir -p uploads
    mkdir -p nginx/ssl
    mkdir -p scripts
    echo -e "${GREEN}✅ 目录创建完成${NC}"
}

# 停止旧服务
stop_old_services() {
    echo -e "${YELLOW}🛑 停止旧服务...${NC}"
    docker-compose down || true
    echo -e "${GREEN}✅ 旧服务已停止${NC}"
}

# 构建和启动服务
start_services() {
    echo -e "${GREEN}🔨 构建Docker镜像...${NC}"
    docker-compose build --no-cache
    
    echo -e "${GREEN}🚀 启动服务...${NC}"
    docker-compose up -d
    
    # 等待服务启动
    echo -e "${GREEN}⏳ 等待服务启动...${NC}"
    sleep 30
}

# 检查服务状态
check_services() {
    echo -e "${GREEN}🔍 检查服务状态...${NC}"
    
    # 检查后端健康
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ 后端服务运行正常${NC}"
    else
        echo -e "${RED}❌ 后端服务启动失败${NC}"
        docker-compose logs backend
        exit 1
    fi
    
    # 检查前端健康  
    if curl -f http://localhost:3001 > /dev/null 2>&1; then
        echo -e "${GREEN}✅ 前端服务运行正常${NC}"
    else
        echo -e "${RED}❌ 前端服务启动失败${NC}"
        docker-compose logs frontend
        exit 1
    fi
    
    # 检查数据库连接
    if docker-compose exec -T postgres pg_isready -U palu_user -d palu_wiki > /dev/null 2>&1; then
        echo -e "${GREEN}✅ 数据库连接正常${NC}"
    else
        echo -e "${RED}❌ 数据库连接失败${NC}"
        docker-compose logs postgres
        exit 1
    fi
}

# 显示部署信息
show_info() {
    echo ""
    echo -e "${GREEN}🎉 部署完成！${NC}"
    echo ""
    echo "📋 服务信息:"
    echo "  🔧 后端API: http://localhost:8080"
    echo "  🎨 前端管理: http://localhost:3001"
    echo "  🌐 Nginx代理: http://localhost (如果启用)"
    echo "  📊 健康检查: http://localhost:8080/health"
    echo ""
    echo "🔍 查看服务状态:"
    echo "  docker-compose ps"
    echo ""
    echo "📋 查看日志:"
    echo "  docker-compose logs -f"
    echo ""
    echo "🛑 停止服务:"
    echo "  docker-compose down"
}

# 主函数
main() {
    echo "📁 当前目录: $(pwd)"
    
    check_docker
    check_env
    create_dirs
    stop_old_services
    start_services
    check_services
    show_info
}

# 运行主程序
main "$@"