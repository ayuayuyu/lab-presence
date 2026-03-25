#!/usr/bin/env bash
set -euo pipefail

PI="${1:-ayu@10.0.3.99}"
REMOTE_DIR="~/lab-presence"

echo "==> Deploying to ${PI}"

# 0. .env ファイルの存在確認
if [[ ! -f .env ]]; then
  echo "[ERROR] .env ファイルが見つかりません。"
  echo "  .env.example をコピーして必要な値を設定してください:"
  echo "    cp .env.example .env"
  echo "    vim .env"
  exit 1
fi

# 1. エージェントをクロスコンパイル (arm64)
echo "==> Cross-compiling agent for linux/arm64..."
cd agent
GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o lab-agent-linux-arm64 ./cmd/agent
cd ..

# 2. rsync でプロジェクトを転送
echo "==> Syncing project files to ${PI}:${REMOTE_DIR}..."
rsync -avz --delete \
  --exclude '.git/' \
  --exclude 'node_modules/' \
  --exclude '.next/' \
  --exclude 'out/' \
  --exclude 'agent/lab-agent' \
  --exclude 'agent/lab-agent-linux-arm' \
  --exclude '.env.local' \
  ./ "${PI}:${REMOTE_DIR}/"

# 3. SSH でセットアップスクリプトを実行
echo "==> Running setup on ${PI}..."
ssh -t "${PI}" "cd ${REMOTE_DIR} && sudo bash scripts/setup-pi.sh"

echo ""
echo "=========================================="
echo "  デプロイ完了!"
echo "=========================================="
echo ""
echo "  Cloudflare Tunnel (ホスト上の cloudflared) で外部公開されています。"
echo "  ラズパイ上の cloudflared → localhost:80 → nginx → 各サービス"
echo ""
echo "  アプリ:  https://<your-domain>/"
echo "=========================================="
