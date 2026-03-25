#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
AGENT_BINARY="${PROJECT_DIR}/agent/lab-agent-linux-arm64"
SERVICE_FILE="${PROJECT_DIR}/agent/lab-agent.service"

echo "=========================================="
echo "  Lab Presence セットアップ (Raspberry Pi)"
echo "=========================================="

# 1. Docker のインストール確認
if ! command -v docker &>/dev/null; then
  echo "==> Installing Docker..."
  curl -fsSL https://get.docker.com | sh
  usermod -aG docker "$(logname)"
fi

# docker compose (plugin) の確認
if ! docker compose version &>/dev/null; then
  echo "==> Installing Docker Compose plugin..."
  apt-get update -qq
  apt-get install -y -qq docker-compose-plugin
fi

# 2. arp-scan のインストール確認
if ! command -v arp-scan &>/dev/null; then
  echo "==> Installing arp-scan..."
  apt-get update -qq
  apt-get install -y -qq arp-scan
fi

# 2.5. fping のインストール確認（ping スキャン用）
if ! command -v fping &>/dev/null; then
  echo "==> Installing fping..."
  apt-get update -qq
  apt-get install -y -qq fping
fi

# 3. ポート80を占有するホスト側サービスを停止
if systemctl is-active --quiet nginx 2>/dev/null; then
  echo "==> Stopping host nginx (port 80 conflict)..."
  systemctl stop nginx
  systemctl disable nginx
fi
if systemctl is-active --quiet apache2 2>/dev/null; then
  echo "==> Stopping host apache2 (port 80 conflict)..."
  systemctl stop apache2
  systemctl disable apache2
fi

# 4. .env 確認
if [[ ! -f "${PROJECT_DIR}/.env" ]]; then
  echo "[ERROR] .env ファイルが見つかりません。deploy.sh 経由で転送されているはずです。"
  exit 1
fi

# 5. Docker Compose でサービス起動 (--remove-orphans で不要コンテナも削除)
echo "==> Starting services..."
cd "${PROJECT_DIR}"
docker compose up -d --build --remove-orphans

# 6. DB起動待機
echo "==> Waiting for DB to be ready..."
sleep 5

# 6.1. DBマイグレーション実行
echo "==> Running database migrations..."
for f in "${PROJECT_DIR}"/backend/db/migrations/*.sql; do
  [ -f "$f" ] || continue
  echo "  applying $(basename "$f")..."
  docker compose exec -T db psql -U lab -d lab_presence -f - < "$f"
done

# 6.5. Cloudflare Tunnel の設定確認・更新
CF_CONFIG="/home/$(logname)/.cloudflared/config.yml"
if [[ -f "${CF_CONFIG}" ]]; then
  echo "==> Updating cloudflared config to point to localhost:80..."
  # url を localhost:80 に書き換え
  if grep -q '^url:' "${CF_CONFIG}"; then
    sed -i 's|^url:.*|url: http://localhost:80|' "${CF_CONFIG}"
  else
    echo 'url: http://localhost:80' >> "${CF_CONFIG}"
  fi
  echo "==> Restarting cloudflared service..."
  systemctl restart cloudflared
else
  echo "[WARN] cloudflared config not found at ${CF_CONFIG}"
  echo "  手動で url: http://localhost:80 に設定してください。"
fi

# 7. エージェントバイナリを配置（実行中なら先に停止）
echo "==> Installing agent binary..."
if systemctl is-active --quiet lab-agent 2>/dev/null; then
  echo "==> Stopping running agent..."
  systemctl stop lab-agent
fi
cp "${AGENT_BINARY}" /usr/local/bin/lab-agent
chmod +x /usr/local/bin/lab-agent

# 8. systemd サービスを登録・起動
echo "==> Configuring systemd service..."
cp "${SERVICE_FILE}" /etc/systemd/system/lab-agent.service
systemctl daemon-reload
systemctl enable lab-agent
systemctl restart lab-agent

# 9. 状態確認
echo ""
echo "=========================================="
echo "  セットアップ完了!"
echo "=========================================="
echo ""
echo "サービス状態:"
docker compose ps
echo ""
echo "エージェント状態:"
systemctl status lab-agent --no-pager || true
echo ""
echo "----------------------------------------"
echo "  アクセス情報"
echo "----------------------------------------"
IP=$(hostname -I | awk '{print $1}')
echo ""
echo "  LAN:"
echo "    アプリ:   http://${IP}/"
echo ""
echo "  外部 (Cloudflare Tunnel):"
echo "    cloudflared (systemd) → localhost:80 → nginx"
echo "    Cloudflare で設定したドメインからアクセスできます。"
echo "----------------------------------------"
