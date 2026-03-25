package handler

import (
	"database/sql"
	"net/http"

	"github.com/ayuayuyu/lab-presence/backend/internal/model"
)

// GET /api/presence/last-seen — 各ユーザーの最終来室日
func HandleLastSeen(db *sql.DB) http.HandlerFunc {
	const query = `
		SELECT u.id, u.name, u.picture, ls.detected_at
		FROM user_last_seen ls
		JOIN users u ON u.id = ls.user_id
		ORDER BY ls.detected_at DESC`

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, "query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var list []model.LastSeen
		for rows.Next() {
			var ls model.LastSeen
			if err := rows.Scan(&ls.UserID, &ls.UserName, &ls.UserPicture, &ls.LastSeenAt); err != nil {
				http.Error(w, "scan failed", http.StatusInternalServerError)
				return
			}
			list = append(list, ls)
		}

		if list == nil {
			list = []model.LastSeen{}
		}
		writeJSON(w, http.StatusOK, list)
	}
}

// GET /api/presence — 現在の在室者一覧
func HandlePresence(db *sql.DB) http.HandlerFunc {
	const query = `SELECT user_id, user_name, user_picture, mac_address, device_label, detected_at FROM current_presence`

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, "query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var list []model.Presence
		for rows.Next() {
			var p model.Presence
			if err := rows.Scan(&p.UserID, &p.UserName, &p.UserPicture, &p.MACAddress, &p.DeviceLabel, &p.DetectedAt); err != nil {
				http.Error(w, "scan failed", http.StatusInternalServerError)
				return
			}
			list = append(list, p)
		}

		if list == nil {
			list = []model.Presence{}
		}
		writeJSON(w, http.StatusOK, list)
	}
}
