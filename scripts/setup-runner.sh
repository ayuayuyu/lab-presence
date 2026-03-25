#!/usr/bin/env bash
# ラズパイに GitHub Actions self-hosted runner をセットアップするスクリプト
# 使い方: bash scripts/setup-runner.sh <GITHUB_REPO> <RUNNER_TOKEN>
#
# GITHUB_REPO  : "your-username/lab-presence" の形式
# RUNNER_TOKEN : GitHub Settings → Actions → Runners → New runner で取得したトークン
#
# 例:
#   bash scripts/setup-runner.sh ayuayuyu/lab-presence AAXXXXXXXXXXXXXX

set -euo pipefail

REPO="${1:?Usage: $0 <owner/repo> <runner-token>}"
TOKEN="${2:?Usage: $0 <owner/repo> <runner-token>}"
RUNNER_VERSION="2.333.0"
RUNNER_DIR="${HOME}/actions-runner"

echo "==> GitHub Actions Runner セットアップ開始"
echo "    リポジトリ: ${REPO}"
echo "    ランナーDir: ${RUNNER_DIR}"
echo ""

# --- ランナーのダウンロードと展開 ---
mkdir -p "${RUNNER_DIR}"
cd "${RUNNER_DIR}"

ARCH=$(uname -m)
case "${ARCH}" in
  aarch64) RUNNER_ARCH="arm64" ;;
  armv7l)  RUNNER_ARCH="arm"   ;;
  x86_64)  RUNNER_ARCH="x64"   ;;
  *) echo "Unsupported arch: ${ARCH}"; exit 1 ;;
esac

RUNNER_PKG="actions-runner-linux-${RUNNER_ARCH}-${RUNNER_VERSION}.tar.gz"
RUNNER_URL="https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/${RUNNER_PKG}"

if [ ! -f "${RUNNER_DIR}/config.sh" ]; then
  echo "==> ランナーをダウンロード中... (${RUNNER_ARCH})"
  curl -fsSL -o "${RUNNER_PKG}" "${RUNNER_URL}"

  # ハッシュファイルをダウンロードして検証
  HASH_FILE="${RUNNER_PKG}.sha256"
  curl -fsSL -o "${HASH_FILE}" "${RUNNER_URL}.sha256"
  echo "==> ハッシュを検証中..."
  sha256sum -c "${HASH_FILE}"
  rm -f "${HASH_FILE}"

  tar xzf "${RUNNER_PKG}"
  rm -f "${RUNNER_PKG}"
else
  echo "==> ランナーは既にダウンロード済みです。スキップします。"
fi

# --- ランナーを設定 ---
echo "==> ランナーを設定中..."
./config.sh \
  --url "https://github.com/${REPO}" \
  --token "${TOKEN}" \
  --name "raspberry-pi" \
  --labels "self-hosted,linux,arm64" \
  --work "_work" \
  --unattended \
  --replace

# --- systemd サービスとして登録 ---
echo "==> systemd サービスとして登録中..."
sudo ./svc.sh install
sudo ./svc.sh start

echo ""
echo "=========================================="
echo "  セットアップ完了!"
echo ""
echo "  ステータス確認:"
echo "    sudo ./svc.sh status"
echo "  ログ確認:"
echo "    journalctl -u actions.runner.* -f"
echo "=========================================="
