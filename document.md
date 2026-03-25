# Lab Presence - 仕様書

研究室メンバーの在室状況を自動検知・表示するシステム。
Raspberry Pi 上のエージェントが ARP スキャンで MAC アドレスを検出し、登録済みデバイスと照合してリアルタイムに在室状況を配信する。

---

## 技術スタック

| レイヤー | 技術 |
|----------|------|
| Backend | Go 1.23 (net/http + gorilla/websocket) |
| Frontend | Next.js 16 / React 19 / TypeScript 5 / Tailwind CSS 4 |
| Database | PostgreSQL 16 |
| Agent | Go CLI (arp-scan ラッパー, Linux ARM64) |
| Infra | Docker Compose + Nginx + Cloudflare Tunnel |

---

## アーキテクチャ

```
┌──────────────┐    ARP scan     ┌─────────────┐  POST /api/scan  ┌──────────┐
│ Raspberry Pi │ ──────────────► │   Agent      │ ───────────────► │ Backend  │
│ (Lab Network)│                 │ (Go binary)  │                  │ (Go API) │
└──────────────┘                 └─────────────┘                  └────┬─────┘
                                                                       │
                                                          ┌────────────┼────────────┐
                                                          │ PostgreSQL │ WebSocket  │
                                                          │            │ broadcast  │
                                                          └────────────┘     │
                                                                             ▼
                                                                      ┌──────────┐
                                                                      │ Frontend  │
                                                                      │ (Next.js) │
                                                                      └──────────┘
```

---

## データベース

### テーブル

#### users

| カラム | 型 | 制約 | 説明 |
|--------|-----|------|------|
| id | SERIAL | PK | |
| name | TEXT | NOT NULL | 表示名 |
| email | TEXT | UNIQUE | Google アカウントメール |
| picture | TEXT | DEFAULT '' | プロフィール画像 URL |
| student_id | TEXT | UNIQUE | 学籍番号 |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | |

#### devices

| カラム | 型 | 制約 | 説明 |
|--------|-----|------|------|
| id | SERIAL | PK | |
| user_id | INTEGER | FK → users(id) CASCADE | 所有者 |
| mac_address | MACADDR | UNIQUE, NOT NULL | デバイスの MAC アドレス |
| label | TEXT | DEFAULT '' | デバイス名 (例: "MacBook Pro") |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | |

#### presence_logs

| カラム | 型 | 制約 | 説明 |
|--------|-----|------|------|
| id | SERIAL | PK | |
| device_id | INTEGER | FK → devices(id) CASCADE | 検出されたデバイス |
| detected_at | TIMESTAMPTZ | DEFAULT NOW() | 検出日時 |

**インデックス:** `idx_devices_mac`, `idx_devices_user`, `idx_presence_device`, `idx_presence_detected` (DESC)

### ビュー: current_presence

過去 5 分以内に検出されたユーザーを一覧する。ユーザーごとに最新の検出レコードのみ返す。

---

## API エンドポイント

**ベース:** Nginx が `/api/*` と `/ws/*` を `backend:8080` にプロキシ。

### 認証

- Google OAuth 2.0 ID トークンを `Authorization: Bearer <idToken>` で送信
- Backend は JWT ペイロードを Base64 デコードして `hd` (ドメイン) と `email` を検証
- **署名検証は行わない** (Google トークンを信頼)
- 管理者は `ADMIN_EMAILS` 環境変数で指定

### エンドポイント一覧

| メソッド | パス | 認証 | 説明 |
|----------|------|------|------|
| POST | `/api/auth/me` | 必須 | ログイン時にユーザーを upsert |
| PUT | `/api/users/me` | 必須 | 自分の表示名を更新 |
| GET | `/api/users` | 必須 | ユーザー一覧取得 |
| POST | `/api/users` | 管理者 | ユーザー作成 |
| GET | `/api/presence` | 必須 | 現在の在室者一覧 |
| GET | `/api/presence/last-seen` | 必須 | 全メンバーの最終検出日時 |
| GET | `/api/devices` | 必須 | デバイス一覧 |
| POST | `/api/devices` | 必須 | デバイス登録 (非管理者は自分のみ) |
| PUT | `/api/devices/{id}` | 管理者 | デバイス更新 |
| DELETE | `/api/devices/{id}` | 管理者 | デバイス削除 |
| POST | `/api/scan` | なし | エージェントが MAC アドレスを報告 |
| GET | `/health` | なし | ヘルスチェック |

### リクエスト/レスポンス

**POST /api/auth/me**
```json
// Request
{ "name": "花田歩夢", "picture": "https://..." }
// Response
{ "id": 1, "name": "花田歩夢", "email": "hanadaayumu@pluslab.org", "picture": "https://...", "student_id": "", "created_at": "..." }
```

**POST /api/scan**
```json
// Request
{ "mac_addresses": ["aa:bb:cc:dd:ee:01", "aa:bb:cc:dd:ee:02"] }
// Response
{ "recorded": 1 }
```

**POST /api/devices**
```json
// Request
{ "user_id": 1, "mac_address": "aa:bb:cc:dd:ee:01", "label": "MacBook Pro" }
// Response
{ "id": 1, "user_id": 1, "mac_address": "aa:bb:cc:dd:ee:01", "label": "MacBook Pro", "created_at": "..." }
```

---

## WebSocket

**エンドポイント:** `/ws/presence` (認証不要)

サーバーからクライアントへ一方向のブロードキャスト。`/api/scan` が成功するたびに全クライアントへ配信される。

```json
{
  "presence": [
    {
      "user_id": 1,
      "user_name": "花田歩夢",
      "user_picture": "https://...",
      "mac_address": "aa:bb:cc:dd:ee:01",
      "device_label": "MacBook Pro",
      "detected_at": "2026-03-15T10:30:00Z"
    }
  ],
  "last_seen": [
    {
      "user_id": 1,
      "user_name": "花田歩夢",
      "user_picture": "https://...",
      "last_seen_at": "2026-03-15T10:30:00Z"
    }
  ]
}
```

---

## Agent (Raspberry Pi)

### 動作フロー

1. 起動時に即時実行、その後 2 分間隔で繰り返し
2. `arp-scan --localnet --plain` を実行し、ネットワーク上のデバイスを検出
3. 出力から MAC アドレスを抽出 (小文字正規化・重複除去)
4. `POST /api/scan` で Backend に送信

### 設定

| フラグ / 環境変数 | デフォルト | 説明 |
|---|---|---|
| `-backend` / `BACKEND_URL` | `http://localhost:8080` | Backend URL |
| `-iface` / `SCAN_INTERFACE` | (未指定=全IF) | スキャン対象のネットワークインターフェース |
| `-interval` | `2m` | スキャン間隔 |

### ビルド

```bash
# Makefile で ARM64 向けクロスコンパイル
GOOS=linux GOARCH=arm64 go build -o lab-agent-linux-arm64 ./cmd/agent
```

---

## Frontend

### ページ構成

| パス | 内容 |
|------|------|
| `/` | ダッシュボード: 在室者カード + 最終検出テーブル |
| `/register` | デバイス登録・管理ページ |

### ダッシュボード (`/`)

- WebSocket (`/ws/presence`) でリアルタイム更新
- WebSocket 切断時は 60 秒ポーリングにフォールバック
- ブラウザタブが非表示のとき通信を一時停止 (`visibilitychange`)
- 在室者数カウント表示
- ユーザーカード: アバター、名前、デバイスラベル、検出時刻
- 最終検出テーブル: 全メンバーの最終在室時刻、「在室中」バッジ
- 時刻表示: 相対時間 ("3秒前", "5分前", "2時間前") / 日付 ("今日", "昨日", "3日前")

### デバイス登録 (`/register`)

- 一般ユーザー: 自分のデバイスのみ登録可能
- 管理者: ユーザー作成 + 全デバイスの CRUD
- MAC アドレス + ラベルのフォーム入力
- デバイス一覧テーブル (管理者のみ編集・削除可)

### 認証フロー

1. Google Sign-In ボタンで ID トークンを取得
2. JWT の `hd` クレームで `pluslab.org` ドメインを検証
3. `localStorage` にトークンを保存
4. `POST /api/auth/me` でユーザーを upsert
5. API が 401 を返した場合、トークンを削除してリロード

---

## デプロイ (Docker Compose)

### サービス構成

| サービス | イメージ | ポート | 説明 |
|----------|---------|--------|------|
| db | postgres:16-alpine | 5432 | データベース |
| backend | Go マルチステージビルド | 8080 | API サーバー |
| nginx | Node ビルド → nginx:alpine | 80 | リバースプロキシ + 静的配信 |
| postgres-exporter | prometheuscommunity/postgres-exporter | - | Prometheus メトリクス |

### Nginx 設定

- `/` → Next.js 静的ファイル配信
- `/api/*` → backend:8080 にプロキシ
- `/ws/*` → backend:8080 に WebSocket プロキシ (`Upgrade` ヘッダー付き、タイムアウト 86400s)
- `/health` → backend:8080

---

## 環境変数

| 変数名 | 設定先 | 説明 |
|--------|--------|------|
| `DATABASE_URL` | Backend | PostgreSQL 接続文字列 |
| `PORT` | Backend | リッスンポート (デフォルト: 8080) |
| `ALLOWED_DOMAIN` | Backend | 許可ドメイン (デフォルト: pluslab.org) |
| `ADMIN_EMAILS` | Backend / Frontend | 管理者メール (カンマ区切り) |
| `NEXT_PUBLIC_GOOGLE_CLIENT_ID` | Frontend (ビルド時) | Google OAuth クライアント ID |
| `NEXT_PUBLIC_ADMIN_EMAILS` | Frontend (ビルド時) | フロント側管理者判定用 |
| `NEXT_PUBLIC_API_URL` | Frontend (ビルド時) | API ベース URL (空ならリレーティブ) |
| `BACKEND_URL` | Agent | Backend の URL |
| `SCAN_INTERFACE` | Agent | ARP スキャン対象 IF |

---

## 検出ロジック

### 在室判定

- `current_presence` ビューが **過去 5 分以内** に検出されたユーザーを在室と判定
- Agent のスキャン間隔はデフォルト 2 分
- つまり、2 回連続でスキャンに検出されなかった場合に「不在」となる

### セキュリティ上の注意点

- JWT の署名検証を行っていない (Google 発行を信頼)
- CORS は全オリジン許可 (Cloudflare Tunnel 前提)
- WebSocket は認証なし (フロントエンドのみアクセス制御)
- `/api/scan` は認証なし (内部ネットワークからのアクセスを想定)
