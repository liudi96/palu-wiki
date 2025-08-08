#!/bin/bash

# 数据库备份脚本
# 用法: ./backup.sh [数据库名]

set -e

# 配置
BACKUP_DIR="/opt/palu-wiki/backups"
DB_NAME=${1:-palu_wiki}
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/${DB_NAME}_$DATE.sql"

# 创建备份目录
mkdir -p "$BACKUP_DIR"

echo "开始备份数据库: $DB_NAME"
echo "备份文件: $BACKUP_FILE"

# 执行备份
docker exec palu-wiki-db-prod pg_dump -U palu_user -d "$DB_NAME" > "$BACKUP_FILE"

# 压缩备份文件
gzip "$BACKUP_FILE"

echo "备份完成: ${BACKUP_FILE}.gz"

# 清理旧备份（保留最近7天）
find "$BACKUP_DIR" -name "${DB_NAME}_*.sql.gz" -mtime +7 -delete

echo "清理完成，已删除7天前的备份文件"

# 备份统计
BACKUP_COUNT=$(find "$BACKUP_DIR" -name "${DB_NAME}_*.sql.gz" | wc -l)
echo "当前备份文件数量: $BACKUP_COUNT"