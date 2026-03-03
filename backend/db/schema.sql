-- Lab Presence Management System - Database Schema

BEGIN;

-- ユーザー（研究室メンバー）
CREATE TABLE users (
    id          SERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    student_id  TEXT UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- デバイス（MACアドレスとユーザーの紐付け）
CREATE TABLE devices (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mac_address MACADDR NOT NULL UNIQUE,
    label       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_devices_mac ON devices (mac_address);
CREATE INDEX idx_devices_user ON devices (user_id);

-- 在室ログ（エージェントが検知したMAC → 在室記録）
CREATE TABLE presence_logs (
    id          SERIAL PRIMARY KEY,
    device_id   INTEGER NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_presence_device ON presence_logs (device_id);
CREATE INDEX idx_presence_detected ON presence_logs (detected_at DESC);

-- 最新の在室状態を高速取得するためのビュー
CREATE VIEW current_presence AS
SELECT DISTINCT ON (d.user_id)
    u.id   AS user_id,
    u.name AS user_name,
    d.mac_address,
    d.label AS device_label,
    pl.detected_at
FROM presence_logs pl
JOIN devices d ON d.id = pl.device_id
JOIN users u   ON u.id = d.user_id
WHERE pl.detected_at > NOW() - INTERVAL '5 minutes'
ORDER BY d.user_id, pl.detected_at DESC;

COMMIT;
