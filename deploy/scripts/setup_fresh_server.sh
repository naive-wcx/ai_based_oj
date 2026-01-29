#!/usr/bin/env bash
set -euo pipefail

# Server-side setup script.
# Usage: sudo bash /opt/oj/setup_fresh_server.sh --domain your-domain.com [--wipe] [--with-java] [--with-isolate]

DOMAIN=""
WIPE_DATA="0"
WITH_JAVA="0"
WITH_ISOLATE="0"
DEPLOY_DIR="/opt/oj"

while [ $# -gt 0 ]; do
  case "$1" in
    --domain)
      DOMAIN="$2"
      shift 2
      ;;
    --wipe)
      WIPE_DATA="1"
      shift
      ;;
    --with-java)
      WITH_JAVA="1"
      shift
      ;;
    --with-isolate)
      WITH_ISOLATE="1"
      shift
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

if [ -z "$DOMAIN" ]; then
  echo "Missing --domain"
  exit 1
fi

echo "[INFO] Install dependencies..."
apt update
apt install -y nginx sqlite3 gcc g++ make python3

if [ "$WITH_JAVA" = "1" ]; then
  apt install -y default-jdk
fi

if [ "$WITH_ISOLATE" = "1" ]; then
  apt install -y isolate
fi

echo "[INFO] Prepare directories..."
mkdir -p "$DEPLOY_DIR"
mkdir -p "$DEPLOY_DIR/data"/{problems,submissions,db,sandbox} "$DEPLOY_DIR/static"
mkdir -p /var/log/oj

if [ "$WIPE_DATA" = "1" ]; then
  echo "[WARN] Wiping data and database..."
  rm -rf "$DEPLOY_DIR/data"
  mkdir -p "$DEPLOY_DIR/data"/{problems,submissions,db,sandbox}
fi

echo "[INFO] Configure config.yaml..."
sed -i 's|mode: debug|mode: release|g' "$DEPLOY_DIR/configs/config.yaml"
sed -i 's|./data|/opt/oj/data|g' "$DEPLOY_DIR/configs/config.yaml"
sed -i 's|workers: 2|workers: 1|g' "$DEPLOY_DIR/configs/config.yaml"

if command -v isolate >/dev/null 2>&1; then
  sed -i 's|sandbox: simple|sandbox: isolate|g' "$DEPLOY_DIR/configs/config.yaml"
fi

echo "[INFO] Ensure .env..."
if [ ! -f "$DEPLOY_DIR/.env" ]; then
  if command -v openssl >/dev/null 2>&1; then
    JWT_SECRET="$(openssl rand -hex 32)"
  else
    JWT_SECRET="$(head -c 48 /dev/urandom | base64 | tr -d '\n' | head -c 64)"
  fi
  cat > "$DEPLOY_DIR/.env" <<EOF
JWT_SECRET=$JWT_SECRET
EOF
fi

echo "[INFO] Set permissions..."
chown -R www-data:www-data "$DEPLOY_DIR" /var/log/oj
chmod +x "$DEPLOY_DIR/oj-server"

echo "[INFO] Install systemd service..."
cp "$DEPLOY_DIR/oj.service" /etc/systemd/system/oj.service
systemctl daemon-reload
systemctl enable oj
systemctl restart oj

echo "[INFO] Configure Nginx..."
cp "$DEPLOY_DIR/oj.conf" /etc/nginx/sites-available/oj
sed -i "s|your-domain.com|$DOMAIN|g" /etc/nginx/sites-available/oj
ln -sf /etc/nginx/sites-available/oj /etc/nginx/sites-enabled/oj
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl reload nginx

echo "[INFO] Done."
echo "Health check: curl http://127.0.0.1:8080/health"
