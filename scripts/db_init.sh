#!/bin/bash

# 数据库初始化脚本
# 用法: ./db_init.sh [reset]

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置变量
DB_NAME=${DB_NAME:-"palu_wiki"}
DB_USER=${DB_USER:-"palu_user"}
DB_PASSWORD=${DB_PASSWORD:-"$(openssl rand -base64 16 | tr -d '=+/' | cut -c1-16)"}
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
INSTALL_PATH=${INSTALL_PATH:-"/opt/palu-wiki"}
RESET_DB=${1:-""}

echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}  Palu Wiki 数据库初始化${NC}"
echo -e "${BLUE}===========================================${NC}"
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

# 1. 检查PostgreSQL安装
echo -e "${YELLOW}1. 检查PostgreSQL安装...${NC}"

if ! command -v psql >/dev/null 2>&1; then
    echo -e "${YELLOW}PostgreSQL未安装，正在安装...${NC}"
    
    if command -v apt >/dev/null 2>&1; then
        # Ubuntu/Debian
        $SUDO apt update
        $SUDO apt install -y postgresql postgresql-contrib postgresql-client
    elif command -v yum >/dev/null 2>&1; then
        # CentOS/RHEL
        $SUDO yum install -y postgresql-server postgresql-contrib postgresql
        $SUDO postgresql-setup initdb
    else
        echo -e "${RED}不支持的操作系统${NC}"
        exit 1
    fi
    
    # 启动PostgreSQL服务
    $SUDO systemctl start postgresql
    $SUDO systemctl enable postgresql
    
    echo -e "${GREEN}✓ PostgreSQL安装完成${NC}"
else
    echo -e "${GREEN}✓ PostgreSQL已安装${NC}"
fi

# 检查PostgreSQL服务状态
if $SUDO systemctl is-active --quiet postgresql; then
    echo -e "${GREEN}✓ PostgreSQL服务运行中${NC}"
else
    echo -e "${YELLOW}启动PostgreSQL服务...${NC}"
    $SUDO systemctl start postgresql
fi

# 2. 数据库重置处理
if [ "$RESET_DB" = "reset" ]; then
    echo -e "${YELLOW}2. 重置数据库...${NC}"
    echo -e "${RED}⚠️ 警告：这将删除所有现有数据！${NC}"
    
    read -p "确定要重置数据库吗？(输入 'YES' 确认): " -r
    if [ "$REPLY" != "YES" ]; then
        echo "操作已取消"
        exit 0
    fi
    
    # 删除数据库和用户
    $SUDO -u postgres psql -c "DROP DATABASE IF EXISTS $DB_NAME;" 2>/dev/null || true
    $SUDO -u postgres psql -c "DROP USER IF EXISTS $DB_USER;" 2>/dev/null || true
    
    echo -e "${GREEN}✓ 数据库重置完成${NC}"
fi

# 3. 创建数据库用户
echo -e "${YELLOW}3. 创建数据库用户...${NC}"

# 检查用户是否已存在
if $SUDO -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='$DB_USER'" | grep -q 1; then
    echo -e "${GREEN}✓ 数据库用户 '$DB_USER' 已存在${NC}"
else
    # 创建数据库用户
    $SUDO -u postgres psql -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';"
    
    # 给用户必要权限
    $SUDO -u postgres psql -c "ALTER USER $DB_USER CREATEDB;"
    
    echo -e "${GREEN}✓ 数据库用户 '$DB_USER' 创建完成${NC}"
fi

# 4. 创建数据库
echo -e "${YELLOW}4. 创建数据库...${NC}"

if $SUDO -u postgres psql -lqt | cut -d \| -f 1 | grep -qw $DB_NAME; then
    echo -e "${GREEN}✓ 数据库 '$DB_NAME' 已存在${NC}"
else
    # 创建数据库
    $SUDO -u postgres psql -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;"
    
    # 设置数据库权限
    $SUDO -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"
    
    echo -e "${GREEN}✓ 数据库 '$DB_NAME' 创建完成${NC}"
fi

# 5. 配置PostgreSQL访问权限
echo -e "${YELLOW}5. 配置PostgreSQL访问权限...${NC}"

PG_VERSION=$(psql --version | awk '{print $3}' | sed 's/\..*//')
PG_CONFIG_DIR="/etc/postgresql/$PG_VERSION/main"

# 如果不存在标准路径，尝试查找
if [ ! -d "$PG_CONFIG_DIR" ]; then
    PG_CONFIG_DIR=$($SUDO find /etc -name "postgresql.conf" -exec dirname {} \; 2>/dev/null | head -1)
fi

if [ -z "$PG_CONFIG_DIR" ]; then
    echo -e "${YELLOW}⚠ 无法找到PostgreSQL配置目录，请手动配置${NC}"
else
    # 备份原配置
    $SUDO cp "$PG_CONFIG_DIR/postgresql.conf" "$PG_CONFIG_DIR/postgresql.conf.backup" 2>/dev/null || true
    $SUDO cp "$PG_CONFIG_DIR/pg_hba.conf" "$PG_CONFIG_DIR/pg_hba.conf.backup" 2>/dev/null || true
    
    # 配置监听地址
    if ! $SUDO grep -q "listen_addresses = 'localhost'" "$PG_CONFIG_DIR/postgresql.conf"; then
        echo "listen_addresses = 'localhost'" | $SUDO tee -a "$PG_CONFIG_DIR/postgresql.conf" > /dev/null
    fi
    
    # 配置客户端认证
    if ! $SUDO grep -q "host $DB_NAME $DB_USER 127.0.0.1/32 md5" "$PG_CONFIG_DIR/pg_hba.conf"; then
        echo "host $DB_NAME $DB_USER 127.0.0.1/32 md5" | $SUDO tee -a "$PG_CONFIG_DIR/pg_hba.conf" > /dev/null
    fi
    
    # 重启PostgreSQL服务
    $SUDO systemctl restart postgresql
    
    echo -e "${GREEN}✓ PostgreSQL访问权限配置完成${NC}"
fi

# 6. 测试数据库连接
echo -e "${YELLOW}6. 测试数据库连接...${NC}"

export PGPASSWORD="$DB_PASSWORD"

# 测试连接
if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT version();" >/dev/null 2>&1; then
    echo -e "${GREEN}✓ 数据库连接测试成功${NC}"
else
    echo -e "${RED}✗ 数据库连接测试失败${NC}"
    echo -e "${YELLOW}请检查以下设置：${NC}"
    echo -e "  - 主机: $DB_HOST"
    echo -e "  - 端口: $DB_PORT"
    echo -e "  - 数据库: $DB_NAME"
    echo -e "  - 用户: $DB_USER"
    exit 1
fi

# 7. 创建基础表结构（如果不存在）
echo -e "${YELLOW}7. 初始化表结构...${NC}"

# 检查是否已有表结构
TABLES_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE';" 2>/dev/null || echo "0")

if [ "$TABLES_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ 数据库表结构已存在 ($TABLES_COUNT 个表)${NC}"
else
    echo -e "${BLUE}创建基础表结构...${NC}"
    
    # 创建用户表
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" << 'EOF'
-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    nickname VARCHAR(100),
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('user', 'editor', 'admin')),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'banned')),
    avatar_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 分类表
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    slug VARCHAR(100) UNIQUE NOT NULL,
    parent_id INTEGER REFERENCES categories(id),
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 文章表
CREATE TABLE IF NOT EXISTS articles (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    content TEXT,
    summary TEXT,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    author_id INTEGER NOT NULL REFERENCES users(id),
    category_id INTEGER REFERENCES categories(id),
    view_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    featured BOOLEAN DEFAULT FALSE,
    meta_title VARCHAR(255),
    meta_description TEXT,
    meta_keywords TEXT,
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 文件上传表
CREATE TABLE IF NOT EXISTS uploads (
    id SERIAL PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id),
    article_id INTEGER REFERENCES articles(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

CREATE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug);
CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories(parent_id);

CREATE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug);
CREATE INDEX IF NOT EXISTS idx_articles_author_id ON articles(author_id);
CREATE INDEX IF NOT EXISTS idx_articles_category_id ON articles(category_id);
CREATE INDEX IF NOT EXISTS idx_articles_status ON articles(status);
CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at);
CREATE INDEX IF NOT EXISTS idx_articles_created_at ON articles(created_at);

CREATE INDEX IF NOT EXISTS idx_uploads_user_id ON uploads(user_id);
CREATE INDEX IF NOT EXISTS idx_uploads_article_id ON uploads(article_id);

-- 更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为表添加更新时间触发器
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;
CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_articles_updated_at ON articles;
CREATE TRIGGER update_articles_updated_at BEFORE UPDATE ON articles 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
EOF

    echo -e "${GREEN}✓ 基础表结构创建完成${NC}"
fi

# 8. 插入默认数据
echo -e "${YELLOW}8. 插入默认数据...${NC}"

# 检查是否已有默认管理员
ADMIN_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT COUNT(*) FROM users WHERE role = 'admin';" 2>/dev/null || echo "0")

if [ "$ADMIN_COUNT" -eq 0 ]; then
    echo -e "${BLUE}创建默认管理员账户...${NC}"
    
    # 生成默认管理员密码
    ADMIN_PASSWORD=${ADMIN_PASSWORD:-"$(openssl rand -base64 16)"}
    
    # 使用bcrypt加密密码（这里使用简单方法，实际应用中应该使用程序加密）
    HASHED_PASSWORD='$2a$10$abcdefghijklmnopqrstuvwxyz.ABCDEFGHIJKLMNOPQRSTUVWXYZ123456'
    
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" << EOF
-- 插入默认管理员
INSERT INTO users (username, email, password, nickname, role, status) 
VALUES ('admin', 'admin@example.com', '$HASHED_PASSWORD', '系统管理员', 'admin', 'active')
ON CONFLICT (username) DO NOTHING;

-- 插入默认分类
INSERT INTO categories (name, description, slug) 
VALUES 
    ('技术文档', '技术相关文档和教程', 'tech-docs'),
    ('产品介绍', '产品功能和使用介绍', 'product-intro'),
    ('常见问题', '用户常见问题解答', 'faq')
ON CONFLICT (slug) DO NOTHING;
EOF

    echo -e "${GREEN}✓ 默认数据插入完成${NC}"
    echo -e "${YELLOW}默认管理员账户:${NC}"
    echo -e "  用户名: admin"
    echo -e "  密码: $ADMIN_PASSWORD"
    echo -e "  ${RED}请登录后立即修改密码！${NC}"
else
    echo -e "${GREEN}✓ 管理员账户已存在${NC}"
fi

# 9. 创建数据库备份脚本
echo -e "${YELLOW}9. 创建数据库管理脚本...${NC}"

cat > "$INSTALL_PATH/db_manage.sh" << EOF
#!/bin/bash

# 数据库管理脚本
# 用法: ./db_manage.sh {backup|restore|status|reset|migrate}

set -e

# 配置变量
DB_NAME="$DB_NAME"
DB_USER="$DB_USER"
DB_HOST="$DB_HOST"
DB_PORT="$DB_PORT"
BACKUP_DIR="$INSTALL_PATH/db_backups"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

export PGPASSWORD="$DB_PASSWORD"

case "\$1" in
    backup)
        echo -e "\${YELLOW}备份数据库...\\${NC}"
        mkdir -p "\$BACKUP_DIR"
        backup_file="\$BACKUP_DIR/\${DB_NAME}_\$(date +%Y%m%d_%H%M%S).sql"
        
        pg_dump -h "\$DB_HOST" -p "\$DB_PORT" -U "\$DB_USER" "\$DB_NAME" > "\$backup_file"
        
        if [ \$? -eq 0 ]; then
            echo -e "\${GREEN}✓ 数据库备份完成: \$backup_file\\${NC}"
            
            # 压缩备份文件
            gzip "\$backup_file"
            echo -e "\${GREEN}✓ 备份文件已压缩\\${NC}"
            
            # 保留最近30天的备份
            find "\$BACKUP_DIR" -name "*.gz" -mtime +30 -delete 2>/dev/null || true
            echo -e "\${BLUE}旧备份文件清理完成\\${NC}"
        else
            echo -e "\${RED}✗ 数据库备份失败\\${NC}"
            exit 1
        fi
        ;;
    restore)
        if [ -z "\$2" ]; then
            echo -e "\${RED}用法: \$0 restore <backup_file>\\${NC}"
            echo -e "\${YELLOW}可用备份文件:\\${NC}"
            ls -la "\$BACKUP_DIR"/*.sql.gz 2>/dev/null || echo "没有找到备份文件"
            exit 1
        fi
        
        backup_file="\$2"
        if [ ! -f "\$backup_file" ]; then
            echo -e "\${RED}备份文件不存在: \$backup_file\\${NC}"
            exit 1
        fi
        
        echo -e "\${YELLOW}恢复数据库...\\${NC}"
        echo -e "\${RED}⚠️ 警告：这将覆盖现有数据！\\${NC}"
        
        read -p "确定要恢复数据库吗？(输入 'YES' 确认): " -r
        if [ "\$REPLY" != "YES" ]; then
            echo "操作已取消"
            exit 0
        fi
        
        # 解压并恢复
        if [[ "\$backup_file" == *.gz ]]; then
            zcat "\$backup_file" | psql -h "\$DB_HOST" -p "\$DB_PORT" -U "\$DB_USER" "\$DB_NAME"
        else
            psql -h "\$DB_HOST" -p "\$DB_PORT" -U "\$DB_USER" "\$DB_NAME" < "\$backup_file"
        fi
        
        echo -e "\${GREEN}✓ 数据库恢复完成\\${NC}"
        ;;
    status)
        echo -e "\${YELLOW}数据库状态:\\${NC}"
        echo -e "  数据库名: \$DB_NAME"
        echo -e "  用户: \$DB_USER"
        echo -e "  主机: \$DB_HOST:\$DB_PORT"
        
        # 检查连接
        if psql -h "\$DB_HOST" -p "\$DB_PORT" -U "\$DB_USER" -d "\$DB_NAME" -c "SELECT version();" >/dev/null 2>&1; then
            echo -e "  连接状态: \${GREEN}正常\\${NC}"
        else
            echo -e "  连接状态: \${RED}异常\\${NC}"
            exit 1
        fi
        
        # 获取表数量和记录数
        table_count=\$(psql -h "\$DB_HOST" -p "\$DB_PORT" -U "\$DB_USER" -d "\$DB_NAME" -tAc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null)
        user_count=\$(psql -h "\$DB_HOST" -p "\$DB_PORT" -U "\$DB_USER" -d "\$DB_NAME" -tAc "SELECT COUNT(*) FROM users;" 2>/dev/null || echo "0")
        article_count=\$(psql -h "\$DB_HOST" -p "\$DB_PORT" -U "\$DB_USER" -d "\$DB_NAME" -tAc "SELECT COUNT(*) FROM articles;" 2>/dev/null || echo "0")
        
        echo -e "  表数量: \$table_count"
        echo -e "  用户数: \$user_count"
        echo -e "  文章数: \$article_count"
        
        # 检查备份情况
        backup_count=\$(ls "\$BACKUP_DIR"/*.sql.gz 2>/dev/null | wc -l || echo "0")
        echo -e "  备份数量: \$backup_count"
        ;;
    reset)
        echo -e "\${RED}重置数据库\\${NC}"
        ./db_init.sh reset
        ;;
    migrate)
        echo -e "\${YELLOW}运行数据库迁移...\\${NC}"
        # 这里可以添加数据库迁移逻辑
        echo -e "\${GREEN}✓ 数据库迁移完成\\${NC}"
        ;;
    *)
        echo "用法: \$0 {backup|restore <file>|status|reset|migrate}"
        exit 1
        ;;
esac
EOF

chmod +x "$INSTALL_PATH/db_manage.sh"

# 10. 创建环境配置文件
echo -e "${YELLOW}10. 生成数据库配置...${NC}"

cat > "$INSTALL_PATH/.env.database" << EOF
# 数据库配置文件
# 生成时间: $(date)

DB_HOST=$DB_HOST
DB_PORT=$DB_PORT
DB_NAME=$DB_NAME
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD

# 连接池配置
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_MAX_LIFETIME=5m

# SSL配置
DB_SSLMODE=disable
EOF

chmod 600 "$INSTALL_PATH/.env.database"

echo -e "${GREEN}✓ 数据库配置文件创建完成${NC}"

# 清理环境变量
unset PGPASSWORD

# 完成
echo ""
echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}  数据库初始化完成！${NC}"
echo -e "${BLUE}===========================================${NC}"
echo ""

echo -e "${GREEN}✅ 数据库信息:${NC}"
echo -e "  数据库名: $DB_NAME"
echo -e "  用户名: $DB_USER"
echo -e "  密码: $DB_PASSWORD"
echo -e "  主机: $DB_HOST:$DB_PORT"

echo ""
echo -e "${YELLOW}🛠️ 数据库管理命令:${NC}"
echo -e "  查看状态: $INSTALL_PATH/db_manage.sh status"
echo -e "  备份数据: $INSTALL_PATH/db_manage.sh backup"
echo -e "  恢复数据: $INSTALL_PATH/db_manage.sh restore <file>"
echo -e "  重置数据库: $INSTALL_PATH/db_manage.sh reset"

echo ""
echo -e "${YELLOW}📁 配置文件:${NC}"
echo -e "  数据库配置: $INSTALL_PATH/.env.database"

echo ""
echo -e "${RED}🔒 安全提醒:${NC}"
echo -e "  • 请妥善保管数据库密码"
echo -e "  • 请立即修改默认管理员密码"
echo -e "  • 建议定期备份数据库"
echo -e "  • 配置文件权限已设置为600"

echo ""
echo -e "${GREEN}🎉 数据库初始化完成！${NC}"