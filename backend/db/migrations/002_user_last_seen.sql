-- 002: user_last_seen テーブル追加
-- 冪等: IF NOT EXISTS / ON CONFLICT で何度実行してもエラーにならない

BEGIN;

-- 1. テーブル作成
CREATE TABLE IF NOT EXISTS user_last_seen (
    user_id     INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    detected_at TIMESTAMPTZ NOT NULL
);

-- 2. 既存の presence_logs から初期データを移行
INSERT INTO user_last_seen (user_id, detected_at)
SELECT d.user_id, MAX(pl.detected_at)
FROM presence_logs pl
JOIN devices d ON d.id = pl.device_id
GROUP BY d.user_id
ON CONFLICT (user_id) DO NOTHING;

COMMIT;
