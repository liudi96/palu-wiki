#!/bin/bash

# SSL证书自动配置脚本
# 用法: ./ssl_setup.sh <domain> [email]

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 参数检查
if [ $# -lt 1 ]; then
    echo -e "${RED}用法: $0 <domain> [email]${NC}"
    echo -e "${YELLOW}例如: $0 wiki.example.com admin@example.com${NC}"
    exit 1
fi

DOMAIN=$1
EMAIL=${2:-"admin@${DOMAIN}"}
NGINX_CONFIG_PATH="/etc/nginx/sites-available"
NGINX_ENABLED_PATH="/etc/nginx/sites-enabled"
SSL_PATH="/etc/letsencrypt/live/${DOMAIN}"
INSTALL_PATH=${INSTALL_PATH:-"/opt/palu-wiki"}

echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}  Palu Wiki SSL证书配置${NC}"
echo -e "${BLUE}===========================================${NC}"
echo ""
echo -e "${YELLOW}域名: ${DOMAIN}${NC}"
echo -e "${YELLOW}邮箱: ${EMAIL}${NC}"
echo ""

# 检查是否为root用户或有sudo权限
if [[ $EUID -eq 0 ]]; then
    SUDO=""
elif sudo -n true 2>/dev/null; then
    SUDO="sudo"
else
    echo -e "${RED}错误: 需要root权限或sudo权限${NC}"
    exit 1
fi

# 1. 验证域名解析
echo -e "${YELLOW}1. 验证域名解析...${NC}"

# 获取服务器公网IP
SERVER_IP=$(curl -s ifconfig.me 2>/dev/null || curl -s icanhazip.com 2>/dev/null || echo "")

if [ -z "$SERVER_IP" ]; then
    echo -e "${RED}无法获取服务器公网IP${NC}"
    exit 1
fi

echo -e "${BLUE}服务器IP: ${SERVER_IP}${NC}"

# 检查域名解析
DOMAIN_IP=$(dig +short $DOMAIN 2>/dev/null || nslookup $DOMAIN 2>/dev/null | grep -A1 "Name:" | tail -n1 | awk '{print $2}' || echo "")

if [ "$DOMAIN_IP" = "$SERVER_IP" ]; then
    echo -e "${GREEN}✓ 域名解析正确${NC}"
else
    echo -e "${YELLOW}⚠ 域名解析检查:${NC}"
    echo -e "  域名 ${DOMAIN} 解析到: ${DOMAIN_IP:-"未解析"}"
    echo -e "  服务器IP: ${SERVER_IP}"
    echo -e "${YELLOW}请确保域名A记录指向服务器IP${NC}"
    
    read -p "是否继续配置？(y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# 2. 安装Certbot
echo -e "${YELLOW}2. 安装Certbot...${NC}"

if ! command -v certbot >/dev/null 2>&1; then
    if command -v apt >/dev/null 2>&1; then
        # Ubuntu/Debian
        $SUDO apt update
        $SUDO apt install -y snapd
        $SUDO snap install core
        $SUDO snap refresh core
        $SUDO snap install --classic certbot
        $SUDO ln -sf /snap/bin/certbot /usr/bin/certbot
    elif command -v yum >/dev/null 2>&1; then
        # CentOS/RHEL
        $SUDO yum install -y epel-release
        $SUDO yum install -y certbot python3-certbot-nginx
    fi
    
    echo -e "${GREEN}✓ Certbot安装完成${NC}"
else
    echo -e "${GREEN}✓ Certbot已安装${NC}"
fi

# 3. 创建临时Nginx配置
echo -e "${YELLOW}3. 创建临时Nginx配置...${NC}"

# 创建临时配置用于证书验证
cat > /tmp/${DOMAIN}.conf << EOF
server {
    listen 80;
    server_name ${DOMAIN} www.${DOMAIN};
    
    # Let's Encrypt 验证
    location /.well-known/acme-challenge/ {
        root /var/www/html;
        allow all;
    }
    
    # 其他请求暂时返回维护页面
    location / {
        root /var/www/html;
        index maintenance.html;
        try_files /maintenance.html =503;
    }
}
EOF

# 创建维护页面
$SUDO mkdir -p /var/www/html
cat > /tmp/maintenance.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Palu Wiki - 正在部署</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
        .container { max-width: 600px; margin: 0 auto; }
        h1 { color: #333; }
        p { color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Palu Wiki</h1>
        <p>网站正在部署中，请稍后访问...</p>
        <p>Site is being deployed, please try again later...</p>
    </div>
</body>
</html>
EOF

$SUDO mv /tmp/maintenance.html /var/www/html/maintenance.html
$SUDO mv /tmp/${DOMAIN}.conf ${NGINX_CONFIG_PATH}/${DOMAIN}

# 启用站点
$SUDO ln -sf ${NGINX_CONFIG_PATH}/${DOMAIN} ${NGINX_ENABLED_PATH}/${DOMAIN}

# 删除默认配置（如果存在）
$SUDO rm -f ${NGINX_ENABLED_PATH}/default

# 测试并重载Nginx
$SUDO nginx -t
$SUDO systemctl reload nginx

echo -e "${GREEN}✓ 临时Nginx配置完成${NC}"

# 4. 申请SSL证书
echo -e "${YELLOW}4. 申请SSL证书...${NC}"

if [ ! -d "$SSL_PATH" ]; then
    echo -e "${BLUE}正在申请 ${DOMAIN} 的SSL证书...${NC}"
    
    # 使用webroot方式申请证书
    $SUDO certbot certonly \
        --webroot \
        --webroot-path=/var/www/html \
        --email "$EMAIL" \
        --agree-tos \
        --non-interactive \
        --domains "$DOMAIN"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ SSL证书申请成功${NC}"
    else
        echo -e "${RED}✗ SSL证书申请失败${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}✓ SSL证书已存在${NC}"
fi

# 5. 创建生产环境Nginx配置
echo -e "${YELLOW}5. 创建生产环境Nginx配置...${NC}"

cat > /tmp/${DOMAIN}_prod.conf << EOF
# Palu Wiki Nginx配置
upstream backend {
    server 127.0.0.1:8080;
    keepalive 32;
}

upstream frontend {
    server 127.0.0.1:3000;
    keepalive 32;
}

# HTTP重定向到HTTPS
server {
    listen 80;
    server_name ${DOMAIN} www.${DOMAIN};
    
    # Let's Encrypt 验证
    location /.well-known/acme-challenge/ {
        root /var/www/html;
        allow all;
    }
    
    # 重定向到HTTPS
    location / {
        return 301 https://\$server_name\$request_uri;
    }
}

# HTTPS主配置
server {
    listen 443 ssl http2;
    server_name ${DOMAIN} www.${DOMAIN};
    
    # SSL证书配置
    ssl_certificate ${SSL_PATH}/fullchain.pem;
    ssl_certificate_key ${SSL_PATH}/privkey.pem;
    
    # SSL安全配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA384;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    ssl_stapling on;
    ssl_stapling_verify on;
    
    # 安全头
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:;" always;
    
    # Gzip压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1000;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml;
    
    # 客户端缓存
    location ~* \.(jpg|jpeg|png|gif|ico|css|js|woff|woff2)$ {
        expires 1y;
        add_header Cache-Control "public, no-transform";
        access_log off;
    }
    
    # API代理到后端
    location /api/ {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }
    
    # 健康检查和监控
    location ~ ^/(health|metrics)$ {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        access_log off;
    }
    
    # 静态文件服务
    location /uploads/ {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        expires 30d;
        add_header Cache-Control "public, no-transform";
    }
    
    # 前端应用
    location / {
        proxy_pass http://frontend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
        
        # 处理前端路由
        try_files \$uri \$uri/ @fallback;
    }
    
    # 前端路由回退
    location @fallback {
        proxy_pass http://frontend;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    # 访问日志
    access_log /var/log/nginx/${DOMAIN}_access.log combined;
    error_log /var/log/nginx/${DOMAIN}_error.log warn;
}
EOF

$SUDO mv /tmp/${DOMAIN}_prod.conf ${NGINX_CONFIG_PATH}/${DOMAIN}

# 测试配置
$SUDO nginx -t

if [ $? -eq 0 ]; then
    $SUDO systemctl reload nginx
    echo -e "${GREEN}✓ 生产环境Nginx配置完成${NC}"
else
    echo -e "${RED}✗ Nginx配置错误${NC}"
    exit 1
fi

# 6. 配置SSL证书自动续期
echo -e "${YELLOW}6. 配置SSL证书自动续期...${NC}"

# 检查certbot自动续期
if $SUDO crontab -l 2>/dev/null | grep -q certbot; then
    echo -e "${GREEN}✓ SSL证书自动续期已配置${NC}"
else
    # 添加自动续期任务
    (
        $SUDO crontab -l 2>/dev/null
        echo "0 12 * * * /usr/bin/certbot renew --quiet --nginx && /usr/bin/systemctl reload nginx"
    ) | $SUDO crontab -
    
    echo -e "${GREEN}✓ SSL证书自动续期配置完成${NC}"
fi

# 7. 创建SSL管理脚本
cat > "$INSTALL_PATH/ssl_manage.sh" << 'EOF'
#!/bin/bash

# SSL证书管理脚本
# 用法: ./ssl_manage.sh {status|renew|test|info}

set -e

DOMAIN=${1:-$(hostname -f)}

case "$1" in
    status)
        echo "SSL证书状态:"
        certbot certificates
        ;;
    renew)
        echo "续期SSL证书..."
        sudo certbot renew --nginx
        sudo systemctl reload nginx
        echo "证书续期完成"
        ;;
    test)
        echo "测试SSL证书自动续期..."
        sudo certbot renew --dry-run
        ;;
    info)
        if [ -z "$2" ]; then
            echo "用法: $0 info <domain>"
            exit 1
        fi
        DOMAIN=$2
        echo "SSL证书信息 ($DOMAIN):"
        echo | openssl s_client -servername $DOMAIN -connect $DOMAIN:443 2>/dev/null | openssl x509 -noout -dates -subject -issuer
        ;;
    *)
        echo "用法: $0 {status|renew|test|info <domain>}"
        exit 1
        ;;
esac
EOF

chmod +x "$INSTALL_PATH/ssl_manage.sh"

echo -e "${GREEN}✓ SSL管理脚本创建完成${NC}"

# 8. 验证SSL配置
echo -e "${YELLOW}8. 验证SSL配置...${NC}"

sleep 5

# 检查HTTPS访问
if curl -ksI https://${DOMAIN} >/dev/null 2>&1; then
    echo -e "${GREEN}✓ HTTPS访问正常${NC}"
else
    echo -e "${YELLOW}⚠ HTTPS访问测试失败，请检查应用是否已启动${NC}"
fi

# 检查SSL证书
ssl_info=$(echo | openssl s_client -servername $DOMAIN -connect $DOMAIN:443 2>/dev/null | openssl x509 -noout -dates 2>/dev/null || echo "无法获取证书信息")
echo -e "${BLUE}SSL证书信息:${NC}"
echo "$ssl_info" | sed 's/^/  /'

# 完成
echo ""
echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}  SSL配置完成！${NC}"
echo -e "${BLUE}===========================================${NC}"
echo ""

echo -e "${GREEN}✅ 配置完成:${NC}"
echo -e "  • 域名: ${DOMAIN}"
echo -e "  • SSL证书: Let's Encrypt"
echo -e "  • 自动续期: 已配置"
echo -e "  • HTTPS强制跳转: 已启用"
echo -e "  • 安全头: 已配置"

echo ""
echo -e "${YELLOW}🔗 访问地址:${NC}"
echo -e "  • HTTPS: https://${DOMAIN}"
echo -e "  • HTTP: http://${DOMAIN} (自动跳转到HTTPS)"

echo ""
echo -e "${YELLOW}🛠️ SSL管理命令:${NC}"
echo -e "  • 查看证书状态: ${INSTALL_PATH}/ssl_manage.sh status"
echo -e "  • 手动续期: ${INSTALL_PATH}/ssl_manage.sh renew"  
echo -e "  • 测试自动续期: ${INSTALL_PATH}/ssl_manage.sh test"
echo -e "  • 查看证书信息: ${INSTALL_PATH}/ssl_manage.sh info ${DOMAIN}"

echo ""
echo -e "${BLUE}🔒 SSL配置完成！现在可以安全访问 https://${DOMAIN}${NC}"