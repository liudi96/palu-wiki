#!/bin/bash

# 腾讯云服务器环境检查脚本
# 用法: ./server_check.sh

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}  Palu Wiki 服务器环境检查${NC}"
echo -e "${BLUE}===========================================${NC}"
echo ""

# 系统信息检查
echo -e "${YELLOW}1. 系统信息检查...${NC}"
echo -e "${BLUE}操作系统:${NC} $(lsb_release -d 2>/dev/null | cut -f2 || cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)"
echo -e "${BLUE}内核版本:${NC} $(uname -r)"
echo -e "${BLUE}架构:${NC} $(uname -m)"
echo -e "${BLUE}主机名:${NC} $(hostname)"
echo -e "${BLUE}当前用户:${NC} $(whoami)"
echo ""

# 硬件资源检查
echo -e "${YELLOW}2. 硬件资源检查...${NC}"
echo -e "${BLUE}CPU信息:${NC}"
cpu_info=$(lscpu | grep -E "^CPU\(s\)|^Model name" | sed 's/^/  /')
echo "$cpu_info"

echo -e "${BLUE}内存信息:${NC}"
free -h | grep -E "^Mem:" | awk '{printf "  总计: %s, 已用: %s, 可用: %s (%.1f%%使用率)\n", $2, $3, $7, ($3/$2)*100}'

echo -e "${BLUE}磁盘空间:${NC}"
df -h | grep -E "^/dev/" | awk '{printf "  %s: %s 总计, %s 已用, %s 可用 (%s)\n", $6, $2, $3, $4, $5}'
echo ""

# 网络连接检查
echo -e "${YELLOW}3. 网络连接检查...${NC}"
if ping -c 3 8.8.8.8 >/dev/null 2>&1; then
    echo -e "${GREEN}✓ 外网连接正常${NC}"
else
    echo -e "${RED}✗ 外网连接失败${NC}"
fi

if ping -c 3 mirrors.tencent.com >/dev/null 2>&1; then
    echo -e "${GREEN}✓ 腾讯云镜像源连接正常${NC}"
else
    echo -e "${YELLOW}⚠ 腾讯云镜像源连接失败${NC}"
fi

echo -e "${BLUE}公网IP:${NC} $(curl -s ifconfig.me 2>/dev/null || echo '无法获取')"
echo -e "${BLUE}本地IP:${NC} $(hostname -I | awk '{print $1}')"
echo ""

# 必需软件检查
echo -e "${YELLOW}4. 必需软件检查...${NC}"

# Docker检查
if command -v docker >/dev/null 2>&1; then
    docker_version=$(docker --version | cut -d' ' -f3 | cut -d',' -f1)
    echo -e "${GREEN}✓ Docker已安装 (版本: $docker_version)${NC}"
    
    # 检查Docker服务状态
    if systemctl is-active --quiet docker; then
        echo -e "${GREEN}✓ Docker服务运行中${NC}"
    else
        echo -e "${YELLOW}⚠ Docker服务未运行${NC}"
    fi
    
    # 检查Docker权限
    if docker ps >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Docker权限正常${NC}"
    else
        echo -e "${YELLOW}⚠ 当前用户无Docker权限，需要sudo或加入docker组${NC}"
    fi
else
    echo -e "${RED}✗ Docker未安装${NC}"
fi

# Docker Compose检查
if command -v docker-compose >/dev/null 2>&1; then
    compose_version=$(docker-compose --version | cut -d' ' -f3 | cut -d',' -f1)
    echo -e "${GREEN}✓ Docker Compose已安装 (版本: $compose_version)${NC}"
elif docker compose version >/dev/null 2>&1; then
    compose_version=$(docker compose version --short)
    echo -e "${GREEN}✓ Docker Compose (Plugin)已安装 (版本: $compose_version)${NC}"
else
    echo -e "${RED}✗ Docker Compose未安装${NC}"
fi

# Git检查
if command -v git >/dev/null 2>&1; then
    git_version=$(git --version | cut -d' ' -f3)
    echo -e "${GREEN}✓ Git已安装 (版本: $git_version)${NC}"
else
    echo -e "${RED}✗ Git未安装${NC}"
fi

# Nginx检查
if command -v nginx >/dev/null 2>&1; then
    nginx_version=$(nginx -v 2>&1 | cut -d' ' -f3 | cut -d'/' -f2)
    echo -e "${GREEN}✓ Nginx已安装 (版本: $nginx_version)${NC}"
    
    if systemctl is-active --quiet nginx; then
        echo -e "${GREEN}✓ Nginx服务运行中${NC}"
    else
        echo -e "${YELLOW}⚠ Nginx服务未运行${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Nginx未安装 (可选，但推荐用作反向代理)${NC}"
fi

# 防火墙检查
echo ""
echo -e "${YELLOW}5. 防火墙和端口检查...${NC}"

if command -v ufw >/dev/null 2>&1; then
    ufw_status=$(ufw status | grep -o "Status: [a-z]*" | cut -d' ' -f2)
    echo -e "${BLUE}UFW状态:${NC} $ufw_status"
    
    if [ "$ufw_status" = "active" ]; then
        echo -e "${BLUE}UFW规则:${NC}"
        ufw status numbered | grep -E "^\[\s*[0-9]" | sed 's/^/  /'
    fi
elif systemctl list-unit-files | grep -q firewalld; then
    if systemctl is-active --quiet firewalld; then
        echo -e "${BLUE}Firewalld:${NC} active"
        echo -e "${BLUE}开放端口:${NC}"
        firewall-cmd --list-ports | sed 's/^/  /'
    else
        echo -e "${BLUE}Firewalld:${NC} inactive"
    fi
else
    echo -e "${YELLOW}⚠ 未检测到常见防火墙(ufw/firewalld)${NC}"
fi

# 检查关键端口
echo -e "${BLUE}端口占用检查:${NC}"
ports_to_check=(22 80 443 3000 8080)
for port in "${ports_to_check[@]}"; do
    if netstat -tln 2>/dev/null | grep -q ":$port "; then
        echo -e "  ${YELLOW}⚠ 端口$port已被占用${NC}"
    else
        echo -e "  ${GREEN}✓ 端口$port可用${NC}"
    fi
done

echo ""

# 系统服务检查
echo -e "${YELLOW}6. 系统服务检查...${NC}"

services_to_check=("ssh" "cron")
for service in "${services_to_check[@]}"; do
    if systemctl is-active --quiet "$service"; then
        echo -e "${GREEN}✓ $service服务运行中${NC}"
    else
        echo -e "${YELLOW}⚠ $service服务未运行${NC}"
    fi
done

echo ""

# 存储空间检查
echo -e "${YELLOW}7. 存储空间检查...${NC}"
total_space=$(df / | tail -1 | awk '{print $2}')
available_space=$(df / | tail -1 | awk '{print $4}')
used_percentage=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')

echo -e "${BLUE}根分区使用情况:${NC} $used_percentage% 已使用"

if [ "$used_percentage" -lt 80 ]; then
    echo -e "${GREEN}✓ 磁盘空间充足${NC}"
elif [ "$used_percentage" -lt 90 ]; then
    echo -e "${YELLOW}⚠ 磁盘空间使用率较高${NC}"
else
    echo -e "${RED}✗ 磁盘空间不足${NC}"
fi

# 推荐至少保留5GB空间用于Docker镜像和日志
available_gb=$((available_space / 1024 / 1024))
if [ "$available_gb" -gt 5 ]; then
    echo -e "${GREEN}✓ 可用空间充足 (${available_gb}GB)${NC}"
else
    echo -e "${YELLOW}⚠ 可用空间较少 (${available_gb}GB)，建议清理或扩容${NC}"
fi

echo ""

# 性能基准测试
echo -e "${YELLOW}8. 简单性能测试...${NC}"

# CPU测试
echo -e "${BLUE}CPU性能 (计算质数):${NC}"
start_time=$(date +%s.%N)
seq 1 10000 | factor | wc -l >/dev/null
end_time=$(date +%s.%N)
cpu_time=$(echo "$end_time - $start_time" | bc -l)
echo -e "  质数计算耗时: ${cpu_time}秒"

# 磁盘写入测试
echo -e "${BLUE}磁盘写入性能:${NC}"
if [ -w /tmp ]; then
    write_speed=$(dd if=/dev/zero of=/tmp/test_write bs=1M count=100 2>&1 | grep -o '[0-9.]*[ ]*MB/s' | tail -1)
    rm -f /tmp/test_write
    echo -e "  写入速度: $write_speed"
fi

echo ""

# 安全检查
echo -e "${YELLOW}9. 基础安全检查...${NC}"

# SSH配置检查
if [ -f /etc/ssh/sshd_config ]; then
    if grep -q "PermitRootLogin yes" /etc/ssh/sshd_config 2>/dev/null; then
        echo -e "${YELLOW}⚠ SSH允许root直接登录，建议禁用${NC}"
    else
        echo -e "${GREEN}✓ SSH已禁用root直接登录${NC}"
    fi
    
    if grep -q "PasswordAuthentication no" /etc/ssh/sshd_config 2>/dev/null; then
        echo -e "${GREEN}✓ SSH已禁用密码认证${NC}"
    else
        echo -e "${YELLOW}⚠ SSH仍允许密码认证，建议使用密钥认证${NC}"
    fi
fi

# 检查是否有fail2ban
if command -v fail2ban-client >/dev/null 2>&1; then
    echo -e "${GREEN}✓ Fail2ban已安装${NC}"
else
    echo -e "${YELLOW}⚠ 建议安装fail2ban防止暴力破解${NC}"
fi

echo ""

# 总结报告
echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}  环境检查完成 - 总结报告${NC}"
echo -e "${BLUE}===========================================${NC}"
echo ""

echo -e "${YELLOW}必需软件安装建议:${NC}"
! command -v docker >/dev/null 2>&1 && echo -e "  ${RED}• 安装Docker${NC}"
! command -v docker-compose >/dev/null 2>&1 && ! docker compose version >/dev/null 2>&1 && echo -e "  ${RED}• 安装Docker Compose${NC}"
! command -v git >/dev/null 2>&1 && echo -e "  ${RED}• 安装Git${NC}"
! command -v nginx >/dev/null 2>&1 && echo -e "  ${YELLOW}• 安装Nginx (推荐)${NC}"

echo ""
echo -e "${YELLOW}安全加固建议:${NC}"
echo -e "  • 配置防火墙规则 (开放22, 80, 443端口)"
echo -e "  • 禁用SSH root登录"
echo -e "  • 使用SSH密钥认证"
echo -e "  • 安装fail2ban防暴力破解"
echo -e "  • 定期更新系统补丁"

echo ""
echo -e "${YELLOW}部署准备:${NC}"
echo -e "  • 确保至少有5GB可用磁盘空间"
echo -e "  • 配置域名A记录指向服务器IP"
echo -e "  • 准备SSL证书 (推荐使用Let's Encrypt)"
echo -e "  • 检查服务器提供商的安全组设置"

echo ""
if command -v docker >/dev/null 2>&1 && (command -v docker-compose >/dev/null 2>&1 || docker compose version >/dev/null 2>&1) && command -v git >/dev/null 2>&1; then
    echo -e "${GREEN}🎉 服务器基本满足部署要求！${NC}"
else
    echo -e "${YELLOW}⚠️ 服务器需要安装必需软件后才能部署${NC}"
fi

echo ""
echo -e "${BLUE}下一步: 运行 ./server_setup.sh 进行自动化环境搭建${NC}"