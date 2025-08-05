#!/bin/bash

# 幻兽帕鲁攻略网站启动脚本
echo "🎮 启动幻兽帕鲁攻略网站..."

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 检查端口是否被占用
check_port() {
    local port=$1
    local service=$2
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null ; then
        echo -e "${YELLOW}警告: 端口 $port 已被占用，$service 可能已在运行${NC}"
        return 1
    fi
    return 0
}

# 启动后端服务
start_backend() {
    echo -e "${GREEN}🚀 启动后端服务...${NC}"
    
    # 检查Go是否安装
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ 错误: Go未安装，请先安装Go${NC}"
        exit 1
    fi
    
    # 检查PostgreSQL是否运行
    if ! pg_isready -h localhost -p 5432 &> /dev/null; then
        echo -e "${YELLOW}⚠️  PostgreSQL未运行，尝试启动...${NC}"
        brew services start postgresql@15 || brew services start postgresql
    fi
    
    # 检查后端端口
    if check_port 8080 "后端服务"; then
        # 编译并运行后端
        if [ ! -f "bin/palu-wiki" ]; then
            echo "📦 编译后端..."
            if [ -n "$HTTP_PROXY" ]; then
                HTTP_PROXY=$HTTP_PROXY HTTPS_PROXY=$HTTPS_PROXY go build -o bin/palu-wiki cmd/server/main.go
            else
                go build -o bin/palu-wiki cmd/server/main.go
            fi
        fi
        
        echo "🎯 启动后端API服务 (端口: 8080)..."
        ./bin/palu-wiki &
        BACKEND_PID=$!
        echo "后端PID: $BACKEND_PID"
        
        # 等待后端启动
        echo "⏳ 等待后端服务启动..."
        sleep 3
        
        # 检查后端是否成功启动
        if curl -s http://localhost:8080/health > /dev/null; then
            echo -e "${GREEN}✅ 后端服务启动成功${NC}"
        else
            echo -e "${RED}❌ 后端服务启动失败${NC}"
            kill $BACKEND_PID 2>/dev/null
            exit 1
        fi
    fi
}

# 启动前端服务
start_frontend() {
    echo -e "${GREEN}🎨 启动前端服务...${NC}"
    
    cd frontend
    
    # 检查Node.js是否安装
    if ! command -v node &> /dev/null; then
        echo -e "${RED}❌ 错误: Node.js未安装，请先安装Node.js${NC}"
        exit 1
    fi
    
    # 检查前端端口
    if check_port 3001 "前端服务"; then
        # 检查依赖是否安装
        if [ ! -d "node_modules" ]; then
            echo "📦 安装前端依赖..."
            if [ -n "$HTTP_PROXY" ]; then
                HTTP_PROXY=$HTTP_PROXY HTTPS_PROXY=$HTTPS_PROXY npm install
            else
                npm install
            fi
        fi
        
        echo "🎯 启动前端开发服务器 (端口: 3001)..."
        npm run dev &
        FRONTEND_PID=$!
        echo "前端PID: $FRONTEND_PID"
        
        cd ..
    else
        cd ..
    fi
}

# 创建管理员用户
create_admin() {
    echo -e "${GREEN}👑 设置管理员用户...${NC}"
    
    # 等待数据库连接
    sleep 2
    
    # 创建管理员用户（如果不存在）
    psql -U palu_user -d palu_wiki -c "
    INSERT INTO users (username, email, password, nickname, role, status, created_at, updated_at)
    VALUES ('admin', 'admin@palu-wiki.com', '\$2a\$10\$YourHashedPasswordHere', '系统管理员', 'admin', 'active', NOW(), NOW())
    ON CONFLICT (username) DO NOTHING;
    
    UPDATE users SET role='admin' WHERE username='admin';
    " 2>/dev/null || echo -e "${YELLOW}⚠️  管理员用户可能已存在或数据库连接失败${NC}"
}

# 显示服务信息
show_info() {
    echo ""
    echo -e "${GREEN}🎉 启动完成！${NC}"
    echo ""
    echo "📋 服务信息:"
    echo "  🔧 后端API: http://localhost:8080"
    echo "  🎨 管理后台: http://localhost:3001"
    echo "  📊 健康检查: http://localhost:8080/health"
    echo ""
    echo "🔐 管理员登录:"
    echo "  用户名: admin"
    echo "  密码: admin123"
    echo "  (需要先在数据库中设置角色为admin)"
    echo ""
    echo "🧪 测试用户:"
    echo "  用户名: testuser"
    echo "  密码: 123456"
    echo ""
    echo "📝 快速测试命令:"
    echo "  curl http://localhost:8080/health"
    echo "  curl http://localhost:8080/api/v1/articles"
    echo ""
    echo "🛑 停止服务: Ctrl+C 或运行 ./stop.sh"
}

# 清理函数
cleanup() {
    echo ""
    echo -e "${YELLOW}🛑 正在停止服务...${NC}"
    
    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null
        echo "后端服务已停止"
    fi
    
    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null
        echo "前端服务已停止"
    fi
    
    # 杀死可能还在运行的进程
    pkill -f "palu-wiki" 2>/dev/null
    pkill -f "next dev" 2>/dev/null
    
    echo -e "${GREEN}✅ 所有服务已停止${NC}"
    exit 0
}

# 设置信号处理
trap cleanup SIGINT SIGTERM

# 主流程
main() {
    echo "📁 当前目录: $(pwd)"
    
    # 启动后端
    start_backend
    
    # 启动前端
    start_frontend
    
    # 创建管理员用户
    create_admin
    
    # 显示信息
    show_info
    
    # 保持脚本运行
    echo "⌨️  按 Ctrl+C 停止所有服务"
    wait
}

# 运行主程序
main