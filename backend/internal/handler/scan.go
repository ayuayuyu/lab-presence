package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ayuayuyu/lab-presence/backend/internal/model"
)

// POST /api/scan — エージェントからのスキャン結果を受け取る
func HandleScan(db *sql.DB, hub *Hub) http.HandlerFunc {
	// 登録済みMACのみ presence_logs に記録するクエリ
	const query = `
		INSERT INTO presence_logs (device_id, detected_at)
		SELECT id, NOW()
		FROM devices
		WHERE mac_address = $1::macaddr
	`

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req model.ScanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		inserted := 0
		for _, mac := range req.MACAddresses {
			res, err := db.Exec(query, mac)
			if err != nil {
				log.Printf("scan insert error (mac=%s): %v", mac, err)
				continue
			}
			if n, _ := res.RowsAffected(); n > 0 {
				inserted++
			}
		}

		// スキャンデータ記録後、WebSocketクライアントに即座にpush
		if inserted > 0 {
			hub.BroadcastPresence(db)
		}

		writeJSON(w, http.StatusOK, map[string]int{"recorded": inserted})
	}
}
