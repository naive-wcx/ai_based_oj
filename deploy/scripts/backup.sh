#!/bin/bash

# OJ 系统备份脚本
# 建议添加到 crontab: 0 3 * * * /opt/oj/backup.sh

set -e

BACKUP_DIR="/opt/oj/backups"
OJ_DIR="/opt/oj"
DATE=$(date +%Y%m%d_%H%M%S)

# 创建备份目录
mkdir -p $BACKUP_DIR

echo "[$(date)] 开始备份..."

# 备份数据库
echo "备份数据库..."
cp $OJ_DIR/data/db/oj.db $BACKUP_DIR/oj_$DATE.db

# 备份题目数据
echo "备份题目数据..."
tar -czf $BACKUP_DIR/problems_$DATE.tar.gz -C $OJ_DIR/data problems

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
