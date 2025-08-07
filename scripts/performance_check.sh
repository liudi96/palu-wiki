#!/bin/bash

# 性能监控脚本
# 用法: ./performance_check.sh [URL]

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
BASE_URL=${1:-"http://localhost"}
ENDPOINTS=(
    "/health"
    "/health/detailed"  
    "/metrics"
    "/"
    "/api/v1/articles"
    "/api/v1/categories"
)

echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}  Palu Wiki 性能监控报告${NC}"
echo -e "${BLUE}===========================================${NC}"
echo -e "${YELLOW}目标URL: $BASE_URL${NC}"
echo ""

# 整体性能测试
echo -e "${YELLOW}端点性能测试:${NC}"
printf "%-30s %-10s %-10s %-10s %s\n" "端点" "状态码" "响应时间" "大小" "状态"

for endpoint in "${ENDPOINTS[@]}"; do
    url="$BASE_URL$endpoint"
    
    # 使用curl测试性能
    response=$(curl -s -o /dev/null -w "%{http_code},%{time_total},%{size_download}" "$url" 2>/dev/null || echo "000,0,0")
    
    IFS=',' read -r status_code response_time size <<< "$response"
    
    # 确定状态颜色
    if [ "$status_code" = "200" ]; then
        status_color="${GREEN}✓${NC}"
    elif [ "$status_code" = "000" ]; then
        status_color="${RED}✗${NC}"
        status_code="ERR"
        response_time="N/A"
        size="N/A"
    else
        status_color="${YELLOW}!${NC}"
    fi
    
    printf "%-30s %-10s %-10ss %-10s %s\n" "$endpoint" "$status_code" "$response_time" "$size" "$status_color"
done

echo ""

# 负载测试（轻量）
echo -e "${YELLOW}并发测试 (10个请求):${NC}"
health_url="$BASE_URL/health"

if command -v curl >/dev/null 2>&1; then
    start_time=$(date +%s.%N)
    
    # 并行执行10个请求
    for i in {1..10}; do
        curl -s "$health_url" > /dev/null &
    done
    wait
    
    end_time=$(date +%s.%N)
    total_time=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "N/A")
    
    echo -e "${BLUE}10个并发请求总耗时: ${total_time}秒${NC}"
    
    if [ "$total_time" != "N/A" ]; then
        avg_time=$(echo "scale=3; $total_time / 10" | bc -l)
        echo -e "${BLUE}平均响应时间: ${avg_time}秒${NC}"
    fi
else
    echo -e "${YELLOW}未找到curl命令，跳过并发测试${NC}"
fi

echo ""

# 获取应用指标
echo -e "${YELLOW}应用性能指标:${NC}"
metrics_response=$(curl -s "$BASE_URL/metrics" 2>/dev/null || echo "{}")

if [ "$metrics_response" != "{}" ] && [ "$metrics_response" != "" ]; then
    # 使用简单的方式解析JSON (如果有jq会更好)
    if command -v jq >/dev/null 2>&1; then
        echo -e "${BLUE}内存使用:${NC}"
        echo "$metrics_response" | jq -r '.memory | to_entries[] | "  \(.key): \(.value)"'
        
        echo -e "${BLUE}Goroutines:${NC} $(echo "$metrics_response" | jq -r '.goroutines')"
        
        echo -e "${BLUE}运行时间:${NC} $(echo "$metrics_response" | jq -r '.uptime_seconds') 秒"
        
        echo -e "${BLUE}数据库连接:${NC}"
        echo "$metrics_response" | jq -r '.database | to_entries[] | "  \(.key): \(.value)"' 2>/dev/null || echo "  数据库指标不可用"
    else
        echo -e "${YELLOW}安装jq以获得更好的指标显示${NC}"
        echo "$metrics_response"
    fi
else
    echo -e "${RED}无法获取应用指标${NC}"
fi

echo ""

# 系统资源使用
echo -e "${YELLOW}系统资源使用:${NC}"
if [ -f /proc/meminfo ]; then
    echo -e "${BLUE}内存使用情况:${NC}"
    free -h | grep -E "^Mem:" | awk '{printf "  总计: %s, 已用: %s, 可用: %s (%.1f%%)\n", $2, $3, $7, ($3/$2)*100}'
fi

if [ -f /proc/loadavg ]; then
    echo -e "${BLUE}系统负载:${NC}"
    loadavg=$(cat /proc/loadavg | cut -d' ' -f1-3)
    echo "  $loadavg"
fi

echo ""

# Docker容器状态（如果在Docker环境中）
if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
    echo -e "${YELLOW}Docker容器状态:${NC}"
    container_names=("palu-wiki-backend-prod" "palu-wiki-frontend-prod" "palu-wiki-nginx-prod")
    
    for container in "${container_names[@]}"; do
        if docker ps --filter "name=$container" --format "table {{.Names}}" | grep -q "$container"; then
            stats=$(docker stats --no-stream --format "{{.CPUPerc}} {{.MemUsage}}" "$container" 2>/dev/null || echo "N/A N/A")
            echo -e "  ${GREEN}✓${NC} $container: CPU: $stats"
        else
            echo -e "  ${RED}✗${NC} $container: 未运行"
        fi
    done
fi

echo ""

# 建议和警告
echo -e "${YELLOW}性能建议:${NC}"

# 检查响应时间
slow_endpoints=()
for endpoint in "${ENDPOINTS[@]}"; do
    url="$BASE_URL$endpoint"
    response_time=$(curl -s -o /dev/null -w "%{time_total}" "$url" 2>/dev/null || echo "0")
    
    # 如果响应时间超过1秒，标记为慢
    if (( $(echo "$response_time > 1.0" | bc -l 2>/dev/null || echo 0) )); then
        slow_endpoints+=("$endpoint ($response_time s)")
    fi
done

if [ ${#slow_endpoints[@]} -gt 0 ]; then
    echo -e "  ${YELLOW}⚠ 以下端点响应较慢:${NC}"
    for slow in "${slow_endpoints[@]}"; do
        echo -e "    - $slow"
    done
    echo -e "  ${YELLOW}建议检查数据库查询和网络连接${NC}"
else
    echo -e "  ${GREEN}✓ 所有端点响应时间正常${NC}"
fi

echo ""
echo -e "${GREEN}性能监控完成！${NC}"