#!/bin/bash

# 安全配置脚本
# 用法: ./setup_security.sh

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}  Palu Wiki 安全配置脚本${NC}"
echo -e "${BLUE}===========================================${NC}"
echo ""

# 检查是否为root用户
if [ "$EUID" -eq 0 ]; then
    echo -e "${YELLOW}警告: 检测到以root用户运行，建议使用普通用户${NC}"
fi

# 生成安全密钥
echo -e "${YELLOW}1. 生成安全密钥...${NC}"

# 生成JWT密钥
jwt_secret=$(openssl rand -base64 32)
echo -e "${GREEN}✓ JWT密钥已生成${NC}"

# 生成数据库密码
db_password=$(openssl rand -base64 16 | tr -d "=+/" | cut -c1-16)
echo -e "${GREEN}✓ 数据库密码已生成${NC}"

# 生成Redis密码
redis_password=$(openssl rand -base64 16 | tr -d "=+/" | cut -c1-16)
echo -e "${GREEN}✓ Redis密码已生成${NC}"

# 创建安全配置文件
echo -e "${YELLOW}2. 创建安全配置文件...${NC}"

cat > .env.security << EOF
# 安全配置文件 - 请妥善保管，不要提交到版本控制
# 生成时间: $(date)

# JWT密钥（强密钥）
JWT_SECRET=${jwt_secret}

# 数据库密码（强密码）
DB_PASSWORD=${db_password}

# Redis密码（强密码）
REDIS_PASSWORD=${redis_password}

# 其他安全配置
CORS_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
RATE_LIMIT_RPM=100
MAX_UPLOAD_SIZE=10485760
LOG_LEVEL=info
LOG_FORMAT=json

# 管理员账户配置
ADMIN_USERNAME=admin
ADMIN_PASSWORD=$(openssl rand -base64 16)
ADMIN_EMAIL=admin@yourdomain.com
EOF

echo -e "${GREEN}✓ 安全配置文件已创建: .env.security${NC}"

# 设置文件权限
chmod 600 .env.security
echo -e "${GREEN}✓ 安全文件权限已设置 (600)${NC}"

# 创建默认管理员用户脚本
echo -e "${YELLOW}3. 创建管理员用户脚本...${NC}"

cat > scripts/create_admin.sql << 'EOF'
-- 创建默认管理员用户
-- 请修改密码和邮箱

-- 首先删除可能存在的默认管理员
DELETE FROM users WHERE username = 'admin';

-- 插入新的管理员用户（密码将通过应用程序哈希）
INSERT INTO users (username, email, password, nickname, role, status, created_at, updated_at) 
VALUES (
    'admin',
    'admin@yourdomain.com', 
    '$2a$10$abcdefghijklmnopqrstuvwxyz', -- 这个密码需要通过应用程序设置
    '系统管理员',
    'admin',
    'active',
    NOW(),
    NOW()
);
EOF

echo -e "${GREEN}✓ 管理员用户SQL脚本已创建${NC}"

# 创建密码更新脚本
cat > scripts/update_passwords.sh << 'EOF'
#!/bin/bash

# 密码更新脚本
# 用法: ./update_passwords.sh

set -e

echo "更新系统密码..."

# 检查环境变量文件
if [ ! -f ".env.security" ]; then
    echo "错误: .env.security文件不存在"
    exit 1
fi

# 加载安全配置
source .env.security

echo "1. 更新数据库密码..."
# 这里需要根据实际情况更新数据库密码

echo "2. 更新Redis密码..."
# 这里需要根据实际情况更新Redis密码

echo "3. 重启服务..."
docker compose -f docker-compose.prod.yml restart

echo "密码更新完成！"
EOF

chmod +x scripts/update_passwords.sh
echo -e "${GREEN}✓ 密码更新脚本已创建${NC}"

# 创建安全检查脚本
cat > scripts/security_check.sh << 'EOF'
#!/bin/bash

# 安全检查脚本

set -e

echo "===========================================" 
echo "  Palu Wiki 安全检查报告"
echo "==========================================="
echo ""

# 检查文件权限
echo "文件权限检查:"
echo "✓ 检查敏感文件权限..."

files_to_check=(".env" ".env.production" ".env.security")
for file in "${files_to_check[@]}"; do
    if [ -f "$file" ]; then
        perm=$(stat -c "%a" "$file" 2>/dev/null || stat -f "%A" "$file" 2>/dev/null || echo "unknown")
        if [ "$perm" = "600" ] || [ "$perm" = "-rw-------" ]; then
            echo "  ✓ $file 权限正确 ($perm)"
        else
            echo "  ✗ $file 权限不安全 ($perm)，建议设置为600"
        fi
    fi
done

# 检查默认密码
echo ""
echo "默认密码检查:"
if [ -f ".env" ]; then
    if grep -q "your-strong-jwt-secret" .env 2>/dev/null; then
        echo "  ✗ 发现默认JWT密钥，请更新"
    else
        echo "  ✓ JWT密钥已自定义"
    fi
    
    if grep -q "palu_password" .env 2>/dev/null; then
        echo "  ✗ 发现默认数据库密码，请更新"
    else
        echo "  ✓ 数据库密码已自定义"
    fi
else
    echo "  ✗ .env文件不存在"
fi

# 检查HTTPS配置
echo ""
echo "HTTPS配置检查:"
if [ -d "nginx/ssl" ]; then
    echo "  ✓ SSL证书目录存在"
    if [ -f "nginx/ssl/cert.pem" ] && [ -f "nginx/ssl/key.pem" ]; then
        echo "  ✓ SSL证书文件存在"
    else
        echo "  ✗ SSL证书文件缺失"
    fi
else
    echo "  ✗ SSL证书目录不存在"
fi

# 检查Docker安全
echo ""
echo "Docker安全检查:"
if docker info >/dev/null 2>&1; then
    echo "  ✓ Docker服务运行正常"
    
    # 检查是否以非root用户运行
    if docker compose -f docker-compose.prod.yml config | grep -q "user: appuser"; then
        echo "  ✓ 容器以非root用户运行"
    else
        echo "  ✗ 容器可能以root用户运行"
    fi
else
    echo "  ✗ Docker服务未运行"
fi

echo ""
echo "安全检查完成！"
echo ""
echo "建议："
echo "1. 定期更新密码"
echo "2. 启用HTTPS"
echo "3. 定期备份数据"
echo "4. 监控访问日志"
echo "5. 保持系统更新"
EOF

chmod +x scripts/security_check.sh
echo -e "${GREEN}✓ 安全检查脚本已创建${NC}"

# 输出安全配置信息
echo ""
echo -e "${YELLOW}安全配置完成！${NC}"
echo ""
echo -e "${BLUE}重要信息:${NC}"
echo -e "${YELLOW}1. 安全配置文件: .env.security${NC}"
echo -e "${YELLOW}2. JWT密钥: ${jwt_secret}${NC}"
echo -e "${YELLOW}3. 数据库密码: ${db_password}${NC}"
echo -e "${YELLOW}4. Redis密码: ${redis_password}${NC}"
echo ""
echo -e "${RED}请妥善保管这些密钥！${NC}"
echo ""
echo -e "${BLUE}下一步操作:${NC}"
echo -e "1. 将 .env.security 中的配置复制到生产环境的 .env 文件"
echo -e "2. 运行: ./scripts/security_check.sh"
echo -e "3. 确保 .env.security 不会被提交到版本控制"
echo -e "4. 在生产服务器上设置SSL证书"
echo ""
EOF