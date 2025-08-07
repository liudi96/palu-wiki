#!/bin/bash

# 服务器安全加固脚本
# 用法: ./security_hardening.sh

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置变量
SSH_PORT=${SSH_PORT:-22}
INSTALL_PATH=${INSTALL_PATH:-"/opt/palu-wiki"}
LOG_FILE="/var/log/palu-wiki-security.log"

echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}  Palu Wiki 服务器安全加固${NC}"
echo -e "${BLUE}===========================================${NC}"
echo ""

# 记录日志函数
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') [SECURITY] $1" | sudo tee -a "$LOG_FILE" > /dev/null
}

# 检查是否为root用户或有sudo权限
if [[ $EUID -eq 0 ]]; then
    SUDO=""
    log "以root用户运行安全加固脚本"
elif sudo -n true 2>/dev/null; then
    SUDO="sudo"
    log "以sudo用户运行安全加固脚本"
else
    echo -e "${RED}错误: 需要root权限或sudo权限${NC}"
    exit 1
fi

# 1. 系统更新和基础安全包
echo -e "${YELLOW}1. 更新系统并安装安全工具...${NC}"
log "开始系统更新和安全工具安装"

if command -v apt >/dev/null 2>&1; then
    $SUDO apt update
    $SUDO apt upgrade -y
    $SUDO apt install -y fail2ban ufw lynis chkrootkit rkhunter unattended-upgrades apt-listchanges
elif command -v yum >/dev/null 2>&1; then
    $SUDO yum update -y
    $SUDO yum install -y fail2ban firewalld lynis chkrootkit rkhunter yum-cron
fi

echo -e "${GREEN}✓ 系统更新和安全工具安装完成${NC}"
log "系统更新和安全工具安装完成"

# 2. SSH安全加固
echo -e "${YELLOW}2. SSH安全加固...${NC}"
log "开始SSH安全加固"

# 备份原始SSH配置
$SUDO cp /etc/ssh/sshd_config /etc/ssh/sshd_config.backup.$(date +%Y%m%d_%H%M%S)

# 创建安全的SSH配置
cat > /tmp/sshd_config_secure << EOF
# Palu Wiki SSH安全配置

# 基础设置
Port $SSH_PORT
Protocol 2
HostKey /etc/ssh/ssh_host_rsa_key
HostKey /etc/ssh/ssh_host_ecdsa_key
HostKey /etc/ssh/ssh_host_ed25519_key

# 安全设置
PermitRootLogin no
MaxAuthTries 3
MaxSessions 2
MaxStartups 10:30:100

# 认证设置
PubkeyAuthentication yes
PasswordAuthentication no
ChallengeResponseAuthentication no
KerberosAuthentication no
GSSAPIAuthentication no
UsePAM yes

# 会话设置
ClientAliveInterval 300
ClientAliveCountMax 2
LoginGraceTime 30
TCPKeepAlive yes

# 访问控制
# AllowUsers your-username
# DenyUsers root

# 协议设置
X11Forwarding no
PrintMotd no
PrintLastLog yes
Compression delayed

# 日志设置
SyslogFacility AUTH
LogLevel INFO

# 其他安全设置
PermitEmptyPasswords no
PermitUserEnvironment no
AllowAgentForwarding no
AllowTcpForwarding no
GatewayPorts no
PermitTunnel no

# Banner
Banner /etc/ssh/ssh_banner
EOF

$SUDO mv /tmp/sshd_config_secure /etc/ssh/sshd_config

# 创建SSH banner
cat > /tmp/ssh_banner << 'EOF'
******************************************************************
*                       Palu Wiki 服务器                        *
*                                                                *
*               仅限授权用户访问！                                *
*            Authorized users only!                             *
*                                                                *
*         所有活动都会被记录和监控                                *
*      All activities are logged and monitored                  *
******************************************************************
EOF

$SUDO mv /tmp/ssh_banner /etc/ssh/ssh_banner

# 测试SSH配置
if $SUDO sshd -t; then
    echo -e "${GREEN}✓ SSH配置测试通过${NC}"
    log "SSH配置测试通过"
else
    echo -e "${RED}✗ SSH配置错误，恢复备份配置${NC}"
    $SUDO mv /etc/ssh/sshd_config.backup.* /etc/ssh/sshd_config
    log "SSH配置错误，已恢复备份"
    exit 1
fi

# 重启SSH服务（小心操作）
echo -e "${YELLOW}注意：即将重启SSH服务，请确保您有其他访问方式${NC}"
read -p "按Enter继续，或Ctrl+C取消..." -r

$SUDO systemctl restart sshd

echo -e "${GREEN}✓ SSH安全加固完成${NC}"
log "SSH安全加固完成"

# 3. 防火墙配置
echo -e "${YELLOW}3. 配置防火墙...${NC}"
log "开始防火墙配置"

if command -v ufw >/dev/null 2>&1; then
    # Ubuntu/Debian - UFW
    $SUDO ufw --force reset
    $SUDO ufw default deny incoming
    $SUDO ufw default allow outgoing
    
    # 允许必要端口
    $SUDO ufw allow $SSH_PORT/tcp comment 'SSH'
    $SUDO ufw allow 80/tcp comment 'HTTP'
    $SUDO ufw allow 443/tcp comment 'HTTPS'
    
    # 限制SSH连接频率
    $SUDO ufw limit $SSH_PORT/tcp
    
    # 启用防火墙
    $SUDO ufw --force enable
    
    echo -e "${GREEN}✓ UFW防火墙配置完成${NC}"
    
elif command -v firewall-cmd >/dev/null 2>&1; then
    # CentOS/RHEL - Firewalld
    $SUDO systemctl start firewalld
    $SUDO systemctl enable firewalld
    
    # 配置防火墙规则
    $SUDO firewall-cmd --permanent --remove-service=ssh
    $SUDO firewall-cmd --permanent --add-port=$SSH_PORT/tcp
    $SUDO firewall-cmd --permanent --add-service=http
    $SUDO firewall-cmd --permanent --add-service=https
    
    # 限制SSH连接（需要rich rules）
    $SUDO firewall-cmd --permanent --add-rich-rule="rule family='ipv4' service name='ssh' accept limit value='5/m'"
    
    $SUDO firewall-cmd --reload
    
    echo -e "${GREEN}✓ Firewalld防火墙配置完成${NC}"
fi

log "防火墙配置完成"

# 4. Fail2ban配置
echo -e "${YELLOW}4. 配置Fail2ban入侵防护...${NC}"
log "开始Fail2ban配置"

# 创建自定义jail配置
cat > /tmp/jail.local << EOF
[DEFAULT]
# 基础配置
bantime = 3600
findtime = 600
maxretry = 3
backend = systemd

# 邮件通知（可选）
# destemail = admin@yourdomain.com
# sendername = Fail2Ban
# action = %(action_mw)s

# SSH防护
[sshd]
enabled = true
port = $SSH_PORT
logpath = /var/log/auth.log
backend = systemd
maxretry = 3
bantime = 3600

# Nginx防护
[nginx-http-auth]
enabled = true
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 6

[nginx-noscript]
enabled = true
port = http,https
logpath = /var/log/nginx/access.log
maxretry = 6

[nginx-badbots]
enabled = true
port = http,https
logpath = /var/log/nginx/access.log
maxretry = 2

[nginx-noproxy]
enabled = true
port = http,https
logpath = /var/log/nginx/access.log
maxretry = 2

# 应用级防护
[palu-wiki-auth]
enabled = true
port = http,https
logpath = /var/log/palu-wiki/access.log
filter = palu-wiki-auth
maxretry = 5
bantime = 1800

# 递归ban
[recidive]
enabled = true
logpath = /var/log/fail2ban.log
action = iptables-allports[name=recidive]
bantime = 86400
findtime = 86400
maxretry = 5
EOF

$SUDO mv /tmp/jail.local /etc/fail2ban/jail.local

# 创建自定义过滤器
$SUDO mkdir -p /etc/fail2ban/filter.d

cat > /tmp/palu-wiki-auth.conf << 'EOF'
[Definition]
failregex = ^.*"level":"ERROR".*"path":"/api/v1/auth/login".*"client_ip":"<HOST>".*$
            ^.*Failed login attempt.*from <HOST>.*$
ignoreregex =
EOF

$SUDO mv /tmp/palu-wiki-auth.conf /etc/fail2ban/filter.d/palu-wiki-auth.conf

# 启动fail2ban
$SUDO systemctl start fail2ban
$SUDO systemctl enable fail2ban

echo -e "${GREEN}✓ Fail2ban配置完成${NC}"
log "Fail2ban配置完成"

# 5. 内核安全参数优化
echo -e "${YELLOW}5. 内核安全参数优化...${NC}"
log "开始内核安全参数优化"

cat > /tmp/99-security.conf << 'EOF'
# 网络安全参数
net.ipv4.ip_forward = 0
net.ipv4.conf.all.send_redirects = 0
net.ipv4.conf.default.send_redirects = 0
net.ipv4.conf.all.accept_source_route = 0
net.ipv4.conf.default.accept_source_route = 0
net.ipv4.conf.all.accept_redirects = 0
net.ipv4.conf.default.accept_redirects = 0
net.ipv4.conf.all.secure_redirects = 0
net.ipv4.conf.default.secure_redirects = 0
net.ipv4.conf.all.log_martians = 1
net.ipv4.conf.default.log_martians = 1
net.ipv4.icmp_echo_ignore_broadcasts = 1
net.ipv4.icmp_ignore_bogus_error_responses = 1
net.ipv4.conf.all.rp_filter = 1
net.ipv4.conf.default.rp_filter = 1
net.ipv4.tcp_syncookies = 1

# IPv6安全参数
net.ipv6.conf.all.accept_source_route = 0
net.ipv6.conf.default.accept_source_route = 0
net.ipv6.conf.all.accept_redirects = 0
net.ipv6.conf.default.accept_redirects = 0
net.ipv6.conf.all.disable_ipv6 = 1
net.ipv6.conf.default.disable_ipv6 = 1

# 内存保护
kernel.dmesg_restrict = 1
kernel.kptr_restrict = 2
kernel.yama.ptrace_scope = 1

# 文件系统安全
fs.suid_dumpable = 0
fs.protected_hardlinks = 1
fs.protected_symlinks = 1
EOF

$SUDO mv /tmp/99-security.conf /etc/sysctl.d/99-security.conf
$SUDO sysctl --system

echo -e "${GREEN}✓ 内核安全参数优化完成${NC}"
log "内核安全参数优化完成"

# 6. 文件系统安全加固
echo -e "${YELLOW}6. 文件系统安全加固...${NC}"
log "开始文件系统安全加固"

# 设置重要文件权限
$SUDO chmod 600 /etc/ssh/sshd_config
$SUDO chmod 644 /etc/passwd
$SUDO chmod 644 /etc/group
$SUDO chmod 600 /etc/shadow
$SUDO chmod 600 /etc/gshadow

# 查找并修复危险权限文件
echo -e "${BLUE}查找危险权限文件...${NC}"
find / -type f -perm -4000 2>/dev/null | head -20 | while read file; do
    echo "  发现SUID文件: $file"
    log "发现SUID文件: $file"
done

# 设置用户目录权限
$SUDO chmod 755 /home/*/ 2>/dev/null || true

echo -e "${GREEN}✓ 文件系统安全加固完成${NC}"
log "文件系统安全加固完成"

# 7. 日志监控配置
echo -e "${YELLOW}7. 配置安全日志监控...${NC}"
log "开始日志监控配置"

# 配置rsyslog安全日志
cat > /tmp/50-security.conf << 'EOF'
# Palu Wiki安全日志配置
# SSH日志
auth,authpriv.*                 /var/log/auth.log

# Fail2ban日志
local0.*                        /var/log/fail2ban.log

# 应用安全日志
local1.*                        /var/log/palu-wiki/security.log

# 内核安全日志
kern.*                          /var/log/kern.log

# 系统日志轮转
$WorkDirectory /var/lib/rsyslog
$ActionFileDefaultTemplate RSYSLOG_TraditionalFileFormat
$FileOwner syslog
$FileGroup adm
$FileCreateMode 0640
$DirCreateMode 0755
$Umask 0022
EOF

$SUDO mv /tmp/50-security.conf /etc/rsyslog.d/50-security.conf
$SUDO systemctl restart rsyslog

# 配置日志轮转
cat > /tmp/palu-wiki-security << 'EOF'
/var/log/palu-wiki/*.log {
    daily
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    create 640 palu-wiki adm
    postrotate
        /bin/kill -HUP `cat /var/run/rsyslogd.pid 2> /dev/null` 2> /dev/null || true
    endscript
}

/var/log/auth.log {
    weekly
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    postrotate
        /bin/kill -HUP `cat /var/run/rsyslogd.pid 2> /dev/null` 2> /dev/null || true
    endscript
}
EOF

$SUDO mv /tmp/palu-wiki-security /etc/logrotate.d/palu-wiki-security

echo -e "${GREEN}✓ 安全日志监控配置完成${NC}"
log "日志监控配置完成"

# 8. 创建安全监控脚本
echo -e "${YELLOW}8. 创建安全监控工具...${NC}"
log "创建安全监控工具"

cat > "$INSTALL_PATH/security_monitor.sh" << 'EOF'
#!/bin/bash

# 安全监控脚本
# 用法: ./security_monitor.sh {check|status|report|ban|unban}

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

case "$1" in
    check)
        echo -e "${YELLOW}安全状态检查:${NC}"
        echo ""
        
        # SSH状态
        echo -e "${BLUE}SSH服务状态:${NC}"
        if systemctl is-active --quiet ssh || systemctl is-active --quiet sshd; then
            echo -e "  ${GREEN}✓ SSH服务运行正常${NC}"
            echo "  监听端口: $(ss -tlnp | grep :22 || ss -tlnp | grep :$SSH_PORT)"
        else
            echo -e "  ${RED}✗ SSH服务异常${NC}"
        fi
        
        # 防火墙状态
        echo -e "${BLUE}防火墙状态:${NC}"
        if command -v ufw >/dev/null 2>&1; then
            ufw status | grep -E "Status|To"
        elif command -v firewall-cmd >/dev/null 2>&1; then
            firewall-cmd --state
            firewall-cmd --list-ports
        fi
        
        # Fail2ban状态
        echo -e "${BLUE}Fail2ban状态:${NC}"
        if systemctl is-active --quiet fail2ban; then
            echo -e "  ${GREEN}✓ Fail2ban运行正常${NC}"
            fail2ban-client status | head -20
        else
            echo -e "  ${RED}✗ Fail2ban未运行${NC}"
        fi
        ;;
    status)
        echo -e "${YELLOW}系统安全状态:${NC}"
        
        # 登录用户
        echo -e "${BLUE}当前登录用户:${NC}"
        who
        
        # 最近登录
        echo -e "${BLUE}最近登录记录:${NC}"
        last -10
        
        # 失败登录
        echo -e "${BLUE}最近失败登录:${NC}"
        lastb -10 2>/dev/null | head -10 || echo "  没有失败登录记录"
        
        # 系统进程
        echo -e "${BLUE}可疑进程检查:${NC}"
        ps aux --sort=-%cpu | head -10
        ;;
    report)
        echo -e "${YELLOW}生成安全报告...${NC}"
        report_file="/tmp/security_report_$(date +%Y%m%d_%H%M%S).txt"
        
        {
            echo "Palu Wiki 安全报告"
            echo "生成时间: $(date)"
            echo "=========================="
            echo ""
            
            echo "系统信息:"
            uname -a
            echo ""
            
            echo "磁盘使用:"
            df -h
            echo ""
            
            echo "内存使用:"
            free -h
            echo ""
            
            echo "网络连接:"
            ss -tuln
            echo ""
            
            echo "防火墙状态:"
            if command -v ufw >/dev/null 2>&1; then
                ufw status numbered
            elif command -v firewall-cmd >/dev/null 2>&1; then
                firewall-cmd --list-all
            fi
            echo ""
            
            echo "Fail2ban状态:"
            fail2ban-client status 2>/dev/null || echo "Fail2ban未运行"
            echo ""
            
            echo "最近安全事件:"
            tail -50 /var/log/auth.log | grep -i "failed\|error\|authentication failure" || echo "没有发现安全事件"
            
        } > "$report_file"
        
        echo -e "${GREEN}✓ 安全报告已生成: $report_file${NC}"
        ;;
    ban)
        if [ -z "$2" ]; then
            echo -e "${RED}用法: $0 ban <IP地址>${NC}"
            exit 1
        fi
        
        IP="$2"
        echo -e "${YELLOW}封禁IP地址: $IP${NC}"
        
        sudo fail2ban-client set sshd banip "$IP"
        echo -e "${GREEN}✓ IP $IP 已被封禁${NC}"
        ;;
    unban)
        if [ -z "$2" ]; then
            echo -e "${RED}用法: $0 unban <IP地址>${NC}"
            exit 1
        fi
        
        IP="$2"
        echo -e "${YELLOW}解封IP地址: $IP${NC}"
        
        sudo fail2ban-client set sshd unbanip "$IP"
        echo -e "${GREEN}✓ IP $IP 已被解封${NC}"
        ;;
    *)
        echo "用法: $0 {check|status|report|ban <ip>|unban <ip>}"
        exit 1
        ;;
esac
EOF

chmod +x "$INSTALL_PATH/security_monitor.sh"

# 9. 自动安全检查定时任务
echo -e "${YELLOW}9. 配置定时安全检查...${NC}"
log "配置定时安全检查"

# 创建每日安全检查脚本
cat > "$INSTALL_PATH/daily_security_check.sh" << 'EOF'
#!/bin/bash

# 每日安全检查脚本

# 生成日志文件名
LOG_FILE="/var/log/palu-wiki/daily-security-$(date +%Y%m%d).log"

{
    echo "=========================================="
    echo "每日安全检查报告 - $(date)"
    echo "=========================================="
    echo ""
    
    # 检查异常登录
    echo "异常登录检查:"
    lastb -10 2>/dev/null | head -5 || echo "没有失败登录记录"
    echo ""
    
    # 检查系统负载
    echo "系统负载检查:"
    uptime
    echo ""
    
    # 检查磁盘使用
    echo "磁盘使用检查:"
    df -h | grep -E "^/dev/" | awk '$5 > 80 {print "警告: "$6" 使用率 "$5}'
    echo ""
    
    # 检查服务状态
    echo "关键服务状态:"
    for service in ssh nginx postgresql fail2ban; do
        if systemctl is-active --quiet $service 2>/dev/null; then
            echo "  ✓ $service: 正常"
        else
            echo "  ✗ $service: 异常"
        fi
    done
    echo ""
    
    # Fail2ban状态
    echo "Fail2ban封禁统计:"
    fail2ban-client status 2>/dev/null | grep "Currently banned" || echo "无法获取状态"
    echo ""
    
} >> "$LOG_FILE" 2>&1

# 如果有异常，发送邮件通知（需要配置邮件服务）
# mail -s "Palu Wiki 每日安全报告" admin@yourdomain.com < "$LOG_FILE"
EOF

chmod +x "$INSTALL_PATH/daily_security_check.sh"

# 添加到crontab
(
    crontab -l 2>/dev/null || echo ""
    echo "# Palu Wiki 每日安全检查"
    echo "0 6 * * * $INSTALL_PATH/daily_security_check.sh"
) | crontab -

echo -e "${GREEN}✓ 定时安全检查配置完成${NC}"
log "定时安全检查配置完成"

# 10. 创建安全恢复脚本
cat > "$INSTALL_PATH/security_recovery.sh" << 'EOF'
#!/bin/bash

# 安全恢复脚本 - 紧急情况下使用

echo "紧急安全恢复模式"
echo "=================="

# 解除所有fail2ban封禁
echo "1. 清除所有fail2ban封禁..."
sudo fail2ban-client unban --all

# 重置防火墙到基本规则
echo "2. 重置防火墙规则..."
if command -v ufw >/dev/null 2>&1; then
    sudo ufw --force reset
    sudo ufw default deny incoming
    sudo ufw default allow outgoing
    sudo ufw allow ssh
    sudo ufw allow 80
    sudo ufw allow 443
    sudo ufw --force enable
fi

# 恢复SSH配置到安全状态
echo "3. 检查SSH配置..."
sudo sshd -t && echo "SSH配置正常" || echo "SSH配置异常，请检查"

# 重启关键服务
echo "4. 重启关键服务..."
for service in ssh nginx fail2ban; do
    sudo systemctl restart $service && echo "  ✓ $service 重启成功" || echo "  ✗ $service 重启失败"
done

echo "安全恢复操作完成！"
EOF

chmod +x "$INSTALL_PATH/security_recovery.sh"

# 完成
echo ""
echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}  服务器安全加固完成！${NC}"
echo -e "${BLUE}===========================================${NC}"
echo ""

echo -e "${GREEN}✅ 安全加固项目:${NC}"
echo -e "  • SSH安全配置 (端口: $SSH_PORT)"
echo -e "  • 防火墙配置 (UFW/Firewalld)"
echo -e "  • Fail2ban入侵防护"
echo -e "  • 内核安全参数优化"
echo -e "  • 文件系统权限加固"
echo -e "  • 安全日志监控"
echo -e "  • 定时安全检查"

echo ""
echo -e "${YELLOW}🛠️ 安全管理工具:${NC}"
echo -e "  • 安全监控: $INSTALL_PATH/security_monitor.sh check"
echo -e "  • 系统状态: $INSTALL_PATH/security_monitor.sh status"
echo -e "  • 生成报告: $INSTALL_PATH/security_monitor.sh report"
echo -e "  • 封禁IP: $INSTALL_PATH/security_monitor.sh ban <IP>"
echo -e "  • 紧急恢复: $INSTALL_PATH/security_recovery.sh"

echo ""
echo -e "${YELLOW}📝 重要提醒:${NC}"
echo -e "  • SSH端口已更改为: $SSH_PORT"
echo -e "  • Root用户登录已禁用"
echo -e "  • 密码认证已禁用，仅允许密钥认证"
echo -e "  • 防火墙已启用，仅开放必要端口"
echo -e "  • Fail2ban已启用入侵检测"

echo ""
echo -e "${RED}🔒 安全注意事项:${NC}"
echo -e "  • 请确保您有SSH密钥访问权限"
echo -e "  • 请记住SSH端口号: $SSH_PORT"
echo -e "  • 建议定期运行安全检查"
echo -e "  • 保持系统和软件更新"
echo -e "  • 定期审查日志文件"

echo ""
echo -e "${BLUE}日志位置:${NC}"
echo -e "  • 安全加固日志: $LOG_FILE"
echo -e "  • SSH日志: /var/log/auth.log"
echo -e "  • Fail2ban日志: /var/log/fail2ban.log"
echo -e "  • 每日安全检查: /var/log/palu-wiki/daily-security-*.log"

log "服务器安全加固全部完成"

echo ""
echo -e "${GREEN}🎉 服务器安全加固完成！${NC}"