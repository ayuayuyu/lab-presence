#!/usr/bin/env bash
# ============================================================
# setup-monitoring.sh
# ラズパイに Grafana + Prometheus 監視スタックをセットアップ
#
# 使い方 (ローカルから):
#   ssh ayu@10.0.3.99 "cd ~/lab-presence && sudo bash scripts/setup-monitoring.sh"
#
# または deploy.sh 後に手動で:
#   cd ~/lab-presence && sudo bash scripts/setup-monitoring.sh
# ============================================================
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "${PROJECT_DIR}"

echo "=========================================="
echo "  Lab Presence 監視スタック セットアップ"
echo "=========================================="

# --- 前提チェック ---
if ! command -v docker &>/dev/null; then
  echo "[ERROR] Docker がインストールされていません。先に setup-pi.sh を実行してください。"
  exit 1
fi

if ! docker compose version &>/dev/null; then
  echo "[ERROR] Docker Compose プラグインがありません。先に setup-pi.sh を実行してください。"
  exit 1
fi

# --- 既存のアプリが起動しているか確認 ---
if ! docker compose ps --status running 2>/dev/null | grep -q "db"; then
  echo "[INFO] メインサービスが起動していません。先に起動します..."
  docker compose up -d --build
  echo "[INFO] DB の起動を待機中..."
  sleep 10
fi

# --- 監視スタックを起動 ---
echo ""
echo "==> 監視スタックを起動中..."
docker compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d --build

echo ""
echo "==> 起動完了を待機中..."
sleep 5

# --- 状態確認 ---
echo ""
echo "=========================================="
echo "  セットアップ完了!"
echo "=========================================="
echo ""
echo "サービス状態:"
docker compose -f docker-compose.yml -f docker-compose.monitoring.yml ps
echo ""
echo "----------------------------------------"
echo "  アクセス情報"
echo "----------------------------------------"
IP=$(hostname -I | awk '{print $1}')
echo ""
echo "  Grafana:    http://${IP}:3000/grafana/"
echo "  Prometheus: http://${IP}:9090"
echo ""
echo "  Grafana ログイン:"
echo "    ユーザー名: admin"
echo "    パスワード: lab-grafana"
echo ""
echo "  ※ 初回ログイン後にパスワードを変更してください"
echo "  ※ ダッシュボード「Lab Presence - Raspberry Pi Monitor」が"
echo "    自動でプロビジョニングされています"
echo "----------------------------------------"
