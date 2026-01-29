#!/usr/bin/env bash
set -euo pipefail

# Local deployment script: build on local machine and deploy to a fresh server.
# Usage: ./deploy/scripts/deploy_fresh_local.sh <server_ip> <domain> [ssh_user] [ssh_port]

if [ $# -lt 2 ]; then
  echo "Usage: $0 <server_ip> <domain> [ssh_user] [ssh_port]"
  exit 1
fi

SERVER_IP="$1"
DOMAIN="$2"
SSH_USER="${3:-root}"
SSH_PORT="${4:-22}"
DEPLOY_DIR="/opt/oj"
PKG_DIR="deploy/package"

echo "[INFO] Build backend..."
(
  cd backend
  go mod tidy
  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o oj-server ./cmd/server
)

echo "[INFO] Build frontend..."
(
  cd frontend
  npm install
  npm run build
)

echo "[INFO] Prepare package..."
rm -rf "$PKG_DIR"
mkdir -p "$PKG_DIR"
cp backend/oj-server "$PKG_DIR"/
cp -r backend/configs "$PKG_DIR"/
cp -r frontend/dist "$PKG_DIR"/static
cp deploy/systemd/oj.service "$PKG_DIR"/
cp deploy/nginx/oj.conf "$PKG_DIR"/
cp deploy/scripts/setup_fresh_server.sh "$PKG_DIR"/

echo "[INFO] Upload to server..."
ssh -p "$SSH_PORT" "$SSH_USER@$SERVER_IP" "sudo mkdir -p $DEPLOY_DIR"
scp -P "$SSH_PORT" -r "$PKG_DIR"/* "$SSH_USER@$SERVER_IP:$DEPLOY_DIR/"

echo "[INFO] Run remote setup (wipe data)..."
ssh -p "$SSH_PORT" "$SSH_USER@$SERVER_IP" "sudo bash $DEPLOY_DIR/setup_fresh_server.sh --domain '$DOMAIN' --wipe"

echo "[INFO] Done. Access: http://$DOMAIN"
