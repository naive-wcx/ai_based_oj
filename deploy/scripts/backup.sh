#!/bin/bash

# OJ 系统备份脚本
# 建议添加到 crontab: 0 3 * * * /opt/oj/backup.sh

set -e

BACKUP_DIR="/opt/oj/backups"
OJ_DIR="/opt/oj"
DATE=$(date +%Y%m%d_%H%M%S)
CONFIG_FILE="$OJ_DIR/configs/config.yaml"

# 从配置读取路径（读取失败时使用默认值）
DB_PATH=$(awk '
  $1=="database:" {in_db=1; next}
  in_db && $1=="path:" {print $2; exit}
  in_db && $0 !~ /^[[:space:]]/ {in_db=0}
' "$CONFIG_FILE")
PROBLEMS_PATH=$(awk '
  $1=="paths:" {in_paths=1; next}
  in_paths && $1=="problems:" {print $2; exit}
  in_paths && $0 !~ /^[[:space:]]/ {in_paths=0}
' "$CONFIG_FILE")

DB_PATH=${DB_PATH:-"$OJ_DIR/data/oj.db"}
PROBLEMS_PATH=${PROBLEMS_PATH:-"$OJ_DIR/data/problems"}

# 创建备份目录
mkdir -p $BACKUP_DIR

echo "[$(date)] 开始备份..."

# 备份数据库
echo "备份数据库..."
cp "$DB_PATH" "$BACKUP_DIR/oj_$DATE.db"

# 备份题目数据
echo "备份题目数据..."
tar -czf "$BACKUP_DIR/problems_$DATE.tar.gz" -C "$PROBLEMS_PATH/.." "$(basename "$PROBLEMS_PATH")"

# 备份配置文件
echo "备份配置文件..."
cp $OJ_DIR/configs/config.yaml $BACKUP_DIR/config_$DATE.yaml
cp $OJ_DIR/.env $BACKUP_DIR/env_$DATE

# 删除 7 天前的备份
echo "清理旧备份..."
find $BACKUP_DIR -name "oj_*.db" -mtime +7 -delete
find $BACKUP_DIR -name "problems_*.tar.gz" -mtime +7 -delete
find $BACKUP_DIR -name "config_*.yaml" -mtime +7 -delete
find $BACKUP_DIR -name "env_*" -mtime +7 -delete

echo "[$(date)] 备份完成！"
echo "备份文件保存在: $BACKUP_DIR"
ls -lh $BACKUP_DIR | head -20
