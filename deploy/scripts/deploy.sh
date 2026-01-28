#!/bin/bash

# OJ 系统部署脚本
# 使用方法: ./deploy.sh [server_ip] [domain]

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查参数
if [ -z "$1" ]; then
    echo "使用方法: $0 <server_ip> [domain]"
    echo "示例: $0 192.168.1.100 oj.example.com"
    exit 1
fi

SERVER_IP=$1
DOMAIN=${2:-$SERVER_IP}
DEPLOY_DIR="/opt/oj"

log_info "开始部署 OJ 系统到 $SERVER_IP"

# 1. 构建后端
log_info "构建后端..."
cd backend

# 检查是否在 Windows 上（使用 WSL 或 Git Bash）
if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "win32" ]]; then
    log_warn "检测到 Windows 环境，请在 WSL 或 Linux 环境中构建"
    log_info "或者使用以下命令在 Windows 上交叉编译:"
    log_info "set GOOS=linux && set GOARCH=amd64 && go build -o oj-server ./cmd/server"
    
    # 尝试使用 PowerShell 设置环境变量并编译
    export GOOS=linux
    export GOARCH=amd64
    export CGO_ENABLED=0
fi

go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o oj-server ./cmd/server
cd ..

# 2. 构建前端
log_info "构建前端..."
cd frontend
npm install
npm run build
cd ..

# 3. 准备部署文件
log_info "准备部署文件..."
mkdir -p deploy/package
cp backend/oj-server deploy/package/
cp -r backend/configs deploy/package/
cp -r frontend/dist deploy/package/static
cp deploy/systemd/oj.service deploy/package/
cp deploy/nginx/oj.conf deploy/package/

# 创建 .env 模板
cat > deploy/package/.env.template << 'EOF'
# DeepSeek API Key（用于 AI 判题）
DEEPSEEK_API_KEY=your_api_key_here

# JWT 密钥（请修改为随机字符串）
JWT_SECRET=your_jwt_secret_here_please_change_it
EOF

# 4. 上传到服务器
log_info "上传文件到服务器..."
ssh root@$SERVER_IP "mkdir -p $DEPLOY_DIR"
scp -r deploy/package/* root@$SERVER_IP:$DEPLOY_DIR/

# 5. 在服务器上执行安装
log_info "在服务器上配置服务..."
ssh root@$SERVER_IP << ENDSSH
set -e

# 安装依赖
apt-get update
apt-get install -y nginx gcc g++ python3

# 创建必要目录
mkdir -p $DEPLOY_DIR/data/{problems,submissions,db,sandbox}
mkdir -p /var/log/oj

# 设置权限
chown -R www-data:www-data $DEPLOY_DIR
chmod +x $DEPLOY_DIR/oj-server

# 配置环境变量
if [ ! -f $DEPLOY_DIR/.env ]; then
    cp $DEPLOY_DIR/.env.template $DEPLOY_DIR/.env
    echo "请编辑 $DEPLOY_DIR/.env 文件设置 API Key 和 JWT 密钥"
fi

# 更新配置中的路径
sed -i 's|./data|/opt/oj/data|g' $DEPLOY_DIR/configs/config.yaml
sed -i 's|mode: debug|mode: release|g' $DEPLOY_DIR/configs/config.yaml

# 安装 systemd 服务
cp $DEPLOY_DIR/oj.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable oj
systemctl restart oj

# 配置 Nginx
cp $DEPLOY_DIR/oj.conf /etc/nginx/sites-available/oj
sed -i "s|your-domain.com|$DOMAIN|g" /etc/nginx/sites-available/oj
ln -sf /etc/nginx/sites-available/oj /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default
nginx -t && systemctl reload nginx

echo "部署完成！"
ENDSSH

log_info "部署完成！"
log_info "请执行以下步骤完成配置："
echo ""
echo "1. SSH 登录服务器: ssh root@$SERVER_IP"
echo "2. 编辑配置文件: nano /opt/oj/.env"
echo "3. 设置 DEEPSEEK_API_KEY 和 JWT_SECRET"
echo "4. 重启服务: systemctl restart oj"
echo ""
echo "5. (可选) 配置 HTTPS:"
echo "   apt install certbot python3-certbot-nginx"
echo "   certbot --nginx -d $DOMAIN"
echo ""
log_info "访问地址: http://$DOMAIN"
log_info "默认管理员账号: admin / admin123"
