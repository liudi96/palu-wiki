#!/bin/bash

# 腾讯云服务器一键环境搭建脚本
# 用法: ./server_setup.sh

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置变量
SETUP_USER=${SETUP_USER:-$(whoami)}
INSTALL_PATH=${INSTALL_PATH:-"/opt/palu-wiki"}
LOG_FILE="/tmp/palu-wiki-setup.log"

echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}  Palu Wiki 服务器环境自动搭建${NC}"
echo -e "${BLUE}===========================================${NC}"
echo ""

# 记录日志函数
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') $1" | tee -a "$LOG_FILE"
}

log "开始环境搭建..."

# 检查是否为root用户或有sudo权限
if [[ $EUID -eq 0 ]]; then
    SUDO=""
    log "以root用户运行"
elif sudo -n true 2>/dev/null; then
    SUDO="sudo"
    log "检测到sudo权限"
else
    echo -e "${RED}错误: 需要root权限或sudo权限来安装软件${NC}"
    exit 1
fi

# 1. 系统更新
echo -e "${YELLOW}1. 更新系统软件包...${NC}"
log "更新系统软件包"

if command -v apt >/dev/null 2>&1; then
    # Ubuntu/Debian
    $SUDO apt update
    $SUDO apt upgrade -y
    $SUDO apt install -y curl wget git vim unzip software-properties-common apt-transport-https ca-certificates gnupg lsb-release
    PACKAGE_MANAGER="apt"
elif command -v yum >/dev/null 2>&1; then
    # CentOS/RHEL
    $SUDO yum update -y
    $SUDO yum install -y curl wget git vim unzip epel-release yum-utils device-mapper-persistent-data lvm2
    PACKAGE_MANAGER="yum"
else
    echo -e "${RED}不支持的操作系统${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 系统更新完成${NC}"
log "系统更新完成"

# 2. 安装Docker
echo -e "${YELLOW}2. 安装Docker...${NC}"
log "开始安装Docker"

if ! command -v docker >/dev/null 2>&1; then
    if [ "$PACKAGE_MANAGER" = "apt" ]; then
        # Ubuntu/Debian安装Docker
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | $SUDO gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
        echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | $SUDO tee /etc/apt/sources.list.d/docker.list > /dev/null
        $SUDO apt update
        $SUDO apt install -y docker-ce docker-ce-cli containerd.io
    else
        # CentOS/RHEL安装Docker
        $SUDO yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
        $SUDO yum install -y docker-ce docker-ce-cli containerd.io
    fi
    
    # 启动Docker服务
    $SUDO systemctl start docker
    $SUDO systemctl enable docker
    
    # 将当前用户加入docker组
    if [ "$EUID" -ne 0 ]; then
        $SUDO usermod -aG docker $USER
        echo -e "${YELLOW}注意: 需要重新登录以使docker组生效${NC}"
    fi
    
    echo -e "${GREEN}✓ Docker安装完成${NC}"
else
    echo -e "${GREEN}✓ Docker已安装${NC}"
fi

log "Docker安装完成"

# 3. 安装Docker Compose
echo -e "${YELLOW}3. 安装Docker Compose...${NC}"
log "开始安装Docker Compose"

if ! command -v docker-compose >/dev/null 2>&1 && ! docker compose version >/dev/null 2>&1; then
    # 安装最新版本的Docker Compose
    DOCKER_COMPOSE_VERSION=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep 'tag_name' | cut -d '"' -f 4)
    $SUDO curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    $SUDO chmod +x /usr/local/bin/docker-compose
    
    # 创建软链接
    $SUDO ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose
    
    echo -e "${GREEN}✓ Docker Compose安装完成 (${DOCKER_COMPOSE_VERSION})${NC}"
else
    echo -e "${GREEN}✓ Docker Compose已安装${NC}"
fi

log "Docker Compose安装完成"

# 4. 安装Nginx
echo -e "${YELLOW}4. 安装Nginx...${NC}"
log "开始安装Nginx"

if ! command -v nginx >/dev/null 2>&1; then
    if [ "$PACKAGE_MANAGER" = "apt" ]; then
        $SUDO apt install -y nginx
    else
        $SUDO yum install -y nginx
    fi
    
    $SUDO systemctl start nginx
    $SUDO systemctl enable nginx
    
    echo -e "${GREEN}✓ Nginx安装完成${NC}"
else
    echo -e "${GREEN}✓ Nginx已安装${NC}"
fi

log "Nginx安装完成"

# 5. 安装Node.js (用于前端构建)
echo -e "${YELLOW}5. 安装Node.js...${NC}"
log "开始安装Node.js"

if ! command -v node >/dev/null 2>&1; then
    # 安装Node.js 18 LTS
    curl -fsSL https://deb.nodesource.com/setup_18.x | $SUDO -E bash -
    if [ "$PACKAGE_MANAGER" = "apt" ]; then
        $SUDO apt install -y nodejs
    else
        $SUDO yum install -y nodejs npm
    fi
    
    echo -e "${GREEN}✓ Node.js安装完成 ($(node -v))${NC}"
else
    echo -e "${GREEN}✓ Node.js已安装 ($(node -v))${NC}"
fi

log "Node.js安装完成"

# 6. 配置防火墙
echo -e "${YELLOW}6. 配置防火墙...${NC}"
log "开始配置防火墙"

if command -v ufw >/dev/null 2>&1; then
    # Ubuntu/Debian使用ufw
    $SUDO ufw --force enable
    $SUDO ufw default deny incoming
    $SUDO ufw default allow outgoing
    $SUDO ufw allow ssh
    $SUDO ufw allow 80/tcp
    $SUDO ufw allow 443/tcp
    echo -e "${GREEN}✓ UFW防火墙配置完成${NC}"
elif command -v firewall-cmd >/dev/null 2>&1; then
    # CentOS/RHEL使用firewalld
    $SUDO systemctl start firewalld
    $SUDO systemctl enable firewalld
    $SUDO firewall-cmd --permanent --zone=public --add-service=ssh
    $SUDO firewall-cmd --permanent --zone=public --add-service=http
    $SUDO firewall-cmd --permanent --zone=public --add-service=https
    $SUDO firewall-cmd --reload
    echo -e "${GREEN}✓ Firewalld防火墙配置完成${NC}"
else
    echo -e "${YELLOW}⚠ 未检测到防火墙，请手动配置${NC}"
fi

log "防火墙配置完成"

# 7. 创建应用目录
echo -e "${YELLOW}7. 创建应用目录...${NC}"
log "创建应用目录: $INSTALL_PATH"

$SUDO mkdir -p "$INSTALL_PATH"
$SUDO chown -R $SETUP_USER:$SETUP_USER "$INSTALL_PATH"
chmod 755 "$INSTALL_PATH"

echo -e "${GREEN}✓ 应用目录创建完成: $INSTALL_PATH${NC}"

# 8. 创建系统用户（可选）
echo -e "${YELLOW}8. 配置系统用户...${NC}"
log "配置系统用户"

if ! id "palu-wiki" &>/dev/null; then
    $SUDO useradd -r -s /bin/false -d /var/lib/palu-wiki -m palu-wiki
    echo -e "${GREEN}✓ 系统用户palu-wiki创建完成${NC}"
else
    echo -e "${GREEN}✓ 系统用户palu-wiki已存在${NC}"
fi

# 9. 安装安全工具
echo -e "${YELLOW}9. 安装安全工具...${NC}"
log "安装安全工具"

# 安装fail2ban
if ! command -v fail2ban-client >/dev/null 2>&1; then
    if [ "$PACKAGE_MANAGER" = "apt" ]; then
        $SUDO apt install -y fail2ban
    else
        $SUDO yum install -y fail2ban
    fi
    
    # 配置fail2ban
    $SUDO systemctl start fail2ban
    $SUDO systemctl enable fail2ban
    
    # 创建基础配置
    cat > /tmp/jail.local << 'EOF'
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3

[sshd]
enabled = true
port = ssh
logpath = /var/log/auth.log
backend = systemd

[nginx-http-auth]
enabled = true
port = http,https
logpath = /var/log/nginx/error.log
EOF
    
    $SUDO mv /tmp/jail.local /etc/fail2ban/jail.local
    $SUDO systemctl restart fail2ban
    
    echo -e "${GREEN}✓ Fail2ban安装配置完成${NC}"
else
    echo -e "${GREEN}✓ Fail2ban已安装${NC}"
fi

log "安全工具安装完成"

# 10. 优化系统参数
echo -e "${YELLOW}10. 优化系统参数...${NC}"
log "优化系统参数"

# 增加文件描述符限制
cat > /tmp/limits.conf << 'EOF'
* soft nofile 65535
* hard nofile 65535
* soft nproc 65535
* hard nproc 65535
EOF

$SUDO mv /tmp/limits.conf /etc/security/limits.d/99-palu-wiki.conf

# 优化网络参数
cat > /tmp/99-palu-wiki.conf << 'EOF'
# TCP优化
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216
net.core.netdev_max_backlog = 5000
net.ipv4.tcp_congestion_control = bbr

# 连接优化
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_time = 1200
net.ipv4.tcp_max_tw_buckets = 5000

# 文件系统优化
fs.file-max = 2097152
vm.swappiness = 10
EOF

$SUDO mv /tmp/99-palu-wiki.conf /etc/sysctl.d/99-palu-wiki.conf
$SUDO sysctl --system

echo -e "${GREEN}✓ 系统参数优化完成${NC}"
log "系统参数优化完成"

# 11. 创建部署脚本
echo -e "${YELLOW}11. 创建部署工具脚本...${NC}"
log "创建部署工具脚本"

cat > "$INSTALL_PATH/deploy.sh" << 'EOF'
#!/bin/bash

# Palu Wiki 应用部署脚本
# 用法: ./deploy.sh [branch]

set -e

BRANCH=${1:-main}
REPO_URL=${REPO_URL:-"https://github.com/yourusername/palu-wiki.git"}
INSTALL_PATH=${INSTALL_PATH:-"/opt/palu-wiki"}
BACKUP_DIR="$INSTALL_PATH/backups"

echo "开始部署 Palu Wiki (分支: $BRANCH)..."

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 备份当前版本
if [ -d "$INSTALL_PATH/app" ]; then
    echo "备份当前版本..."
    backup_name="backup_$(date +%Y%m%d_%H%M%S)"
    cp -r "$INSTALL_PATH/app" "$BACKUP_DIR/$backup_name"
    echo "备份完成: $BACKUP_DIR/$backup_name"
fi

# 克隆或更新代码
if [ ! -d "$INSTALL_PATH/app" ]; then
    echo "克隆项目代码..."
    git clone "$REPO_URL" "$INSTALL_PATH/app"
    cd "$INSTALL_PATH/app"
    git checkout "$BRANCH"
else
    echo "更新项目代码..."
    cd "$INSTALL_PATH/app"
    git fetch origin
    git checkout "$BRANCH"
    git pull origin "$BRANCH"
fi

# 构建并启动应用
echo "构建并启动应用..."
docker-compose -f docker-compose.prod.yml down --remove-orphans
docker-compose -f docker-compose.prod.yml build --no-cache
docker-compose -f docker-compose.prod.yml up -d

# 等待服务启动
echo "等待服务启动..."
sleep 30

# 健康检查
if curl -f http://localhost:8080/health >/dev/null 2>&1; then
    echo "✓ 部署成功！"
    echo "应用已启动，访问 http://$(curl -s ifconfig.me) 查看"
else
    echo "✗ 部署失败，正在回滚..."
    
    # 回滚到最新备份
    latest_backup=$(ls -t "$BACKUP_DIR" | head -n1)
    if [ -n "$latest_backup" ]; then
        echo "回滚到: $latest_backup"
        rm -rf "$INSTALL_PATH/app"
        cp -r "$BACKUP_DIR/$latest_backup" "$INSTALL_PATH/app"
        cd "$INSTALL_PATH/app"
        docker-compose -f docker-compose.prod.yml up -d
    fi
    
    exit 1
fi
EOF

chmod +x "$INSTALL_PATH/deploy.sh"

echo -e "${GREEN}✓ 部署脚本创建完成${NC}"

# 12. 创建服务管理脚本
cat > "$INSTALL_PATH/manage.sh" << 'EOF'
#!/bin/bash

# Palu Wiki 服务管理脚本
# 用法: ./manage.sh {start|stop|restart|status|logs|update}

set -e

INSTALL_PATH=${INSTALL_PATH:-"/opt/palu-wiki"}
cd "$INSTALL_PATH/app" 2>/dev/null || { echo "应用未安装"; exit 1; }

case "$1" in
    start)
        echo "启动 Palu Wiki..."
        docker-compose -f docker-compose.prod.yml up -d
        echo "服务已启动"
        ;;
    stop)
        echo "停止 Palu Wiki..."
        docker-compose -f docker-compose.prod.yml down
        echo "服务已停止"
        ;;
    restart)
        echo "重启 Palu Wiki..."
        docker-compose -f docker-compose.prod.yml restart
        echo "服务已重启"
        ;;
    status)
        echo "Palu Wiki 服务状态:"
        docker-compose -f docker-compose.prod.yml ps
        echo ""
        echo "健康检查:"
        curl -f http://localhost:8080/health 2>/dev/null && echo "✓ 服务正常" || echo "✗ 服务异常"
        ;;
    logs)
        echo "查看服务日志 (Ctrl+C退出):"
        docker-compose -f docker-compose.prod.yml logs -f
        ;;
    update)
        echo "更新应用..."
        ./deploy.sh
        ;;
    *)
        echo "用法: $0 {start|stop|restart|status|logs|update}"
        exit 1
        ;;
esac
EOF

chmod +x "$INSTALL_PATH/manage.sh"

echo -e "${GREEN}✓ 管理脚本创建完成${NC}"

log "部署工具脚本创建完成"

# 安装完成
echo ""
echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}  环境搭建完成！${NC}" 
echo -e "${BLUE}===========================================${NC}"
echo ""

echo -e "${GREEN}✅ 安装完成的软件:${NC}"
echo -e "  • Docker $(docker --version 2>/dev/null | cut -d' ' -f3 | cut -d',' -f1)"
echo -e "  • Docker Compose $(docker-compose --version 2>/dev/null | cut -d' ' -f3 | cut -d',' -f1)"
echo -e "  • Nginx $(nginx -v 2>&1 | cut -d' ' -f3 | cut -d'/' -f2)"
echo -e "  • Node.js $(node -v)"
echo -e "  • Git $(git --version | cut -d' ' -f3)"

echo ""
echo -e "${YELLOW}📁 应用目录: $INSTALL_PATH${NC}"
echo ""

echo -e "${YELLOW}🚀 下一步操作:${NC}"
echo -e "  1. 配置域名DNS解析指向服务器IP"
echo -e "  2. 运行: cd $INSTALL_PATH && ./deploy.sh"
echo -e "  3. 配置SSL证书: ./ssl_setup.sh your-domain.com"
echo -e "  4. 查看服务状态: ./manage.sh status"

echo ""
echo -e "${YELLOW}📝 重要提醒:${NC}"
echo -e "  • 服务器IP: $(curl -s ifconfig.me 2>/dev/null || echo '请手动获取')"
echo -e "  • 防火墙已配置端口: 22(SSH), 80(HTTP), 443(HTTPS)"
echo -e "  • 日志文件: $LOG_FILE"
echo -e "  • 如果添加了用户到docker组，请重新登录生效"

log "环境搭建全部完成"
echo ""
echo -e "${GREEN}🎉 Palu Wiki 服务器环境搭建完成！${NC}"