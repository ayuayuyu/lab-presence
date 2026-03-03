package model

import "time"

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	StudentID string    `json:"student_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Device struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	MACAddress string    `json:"mac_address"`
	Label      string    `json:"label"`
	CreatedAt  time.Time `json:"created_at"`
}

// エージェントからのスキャン結果
type ScanRequest struct {
	MACAddresses []string `json:"mac_addresses"`
}

// 在室状態（current_presence ビューの行）
type Presence struct {
	UserID      int       `json:"user_id"`
	UserName    string    `json:"user_name"`
	MACAddress  string    `json:"mac_address"`
	DeviceLabel string    `json:"device_label"`
	DetectedAt  time.Time `json:"detected_at"`
}

// 最終来室記録
type LastSeen struct {
	UserID     int       `json:"user_id"`
	UserName   string    `json:"user_name"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

// ユーザー作成リクエスト
type CreateUserRequest struct {
	Name      string `json:"name"`
	StudentID string `json:"student_id"`
}

// デバイス登録リクエスト
type CreateDeviceRequest struct {
	UserID     int    `json:"user_id"`
	MACAddress string `json:"mac_address"`
	Label      string `json:"label"`
}

// デバイス更新リクエスト
type UpdateDeviceRequest struct {
	UserID     int    `json:"user_id"`
	MACAddress string `json:"mac_address"`
	Label      string `json:"label"`
}
