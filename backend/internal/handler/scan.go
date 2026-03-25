package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/ayuayuyu/lab-presence/backend/internal/model"
	"github.com/lib/pq"
)

// scanCount はスキャン回数をカウントし、定期的な古いログ削除に使用する。
var scanCount atomic.Int64

// POST /api/scan — エージェントからのスキャン結果を受け取る
func HandleScan(db *sql.DB, hub *Hub) http.HandlerFunc {
	// バッチINSERT: 登録済みMACのみ presence_logs に記録
	const insertPresence = `
		INSERT INTO presence_logs (device_id, detected_at)
		SELECT d.id, NOW()
		FROM unnest($1::macaddr[]) AS mac(addr)
		JOIN devices d ON d.mac_address = mac.addr`

	// user_last_seen を1クエリでUPSERT
	const upsertLastSeen = `
		INSERT INTO user_last_seen (user_id, detected_at)
		SELECT d.user_id, NOW()
		FROM unnest($1::macaddr[]) AS mac(addr)
		JOIN devices d ON d.mac_address = mac.addr
		ON CONFLICT (user_id) DO UPDATE SET detected_at = EXCLUDED.detected_at`

	// 古いログを削除（30日超過）
	const cleanupOldLogs = `DELETE FROM presence_logs WHERE detected_at < NOW() - INTERVAL '30 days'`

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

		if len(req.MACAddresses) == 0 {
			writeJSON(w, http.StatusOK, map[string]int{"recorded": 0})
			return
		}

		macArray := pq.Array(req.MACAddresses)

		// presence_logs にバッチINSERT
		res, err := db.Exec(insertPresence, macArray)
		if err != nil {
			log.Printf("scan batch insert error: %v", err)
			http.Error(w, "insert failed", http.StatusInternalServerError)
			return
		}
		inserted, _ := res.RowsAffected()

		// user_last_seen をUPSERT
		if _, err := db.Exec(upsertLastSeen, macArray); err != nil {
			log.Printf("scan upsert last_seen error: %v", err)
		}

		// 100回に1回、古いログを削除
		if scanCount.Add(1)%100 == 0 {
			go func() {
				if _, err := db.Exec(cleanupOldLogs); err != nil {
					log.Printf("cleanup old logs error: %v", err)
				}
			}()
		}

		// スキャンデータ記録後、WebSocketクライアントに即座にpush
		if inserted > 0 {
			hub.BroadcastPresence(db)
		}

		writeJSON(w, http.StatusOK, map[string]int{"recorded": int(inserted)})
	}
}
