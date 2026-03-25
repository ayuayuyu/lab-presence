package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ayuayuyu/lab-presence/backend/internal/model"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub は接続中のWebSocketクライアントを管理する。
type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

// NewHub は新しいHubを作成する。
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]struct{}),
	}
}

// HandleWS は /ws/presence のWebSocket接続を処理する。
func (h *Hub) HandleWS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws upgrade error: %v", err)
			return
		}

		h.mu.Lock()
		h.clients[conn] = struct{}{}
		h.mu.Unlock()

		// クライアントの切断を検知するためにreadループ
		go func() {
			defer func() {
				h.mu.Lock()
				delete(h.clients, conn)
				h.mu.Unlock()
				conn.Close()
			}()
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					break
				}
			}
		}()
	}
}

// wsPresencePayload はWebSocketで送信する在室データ。
type wsPresencePayload struct {
	Presence []model.Presence `json:"presence"`
	LastSeen []model.LastSeen `json:"last_seen"`
}

// BroadcastPresence はDBから最新の在室データを取得し全クライアントに送信する。
func (h *Hub) BroadcastPresence(db *sql.DB) {
	h.mu.Lock()
	count := len(h.clients)
	h.mu.Unlock()

	if count == 0 {
		return
	}

	presence, err := queryPresence(db)
	if err != nil {
		log.Printf("ws broadcast: presence query error: %v", err)
		return
	}

	lastSeen, err := queryLastSeen(db)
	if err != nil {
		log.Printf("ws broadcast: last_seen query error: %v", err)
		return
	}

	payload := wsPresencePayload{
		Presence: presence,
		LastSeen: lastSeen,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("ws broadcast: marshal error: %v", err)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for conn := range h.clients {
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("ws write error: %v", err)
			conn.Close()
			delete(h.clients, conn)
		}
	}
}

func queryPresence(db *sql.DB) ([]model.Presence, error) {
	const query = `SELECT user_id, user_name, user_picture, mac_address, device_label, detected_at FROM current_presence`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Presence
	for rows.Next() {
		var p model.Presence
		if err := rows.Scan(&p.UserID, &p.UserName, &p.UserPicture, &p.MACAddress, &p.DeviceLabel, &p.DetectedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	if list == nil {
		list = []model.Presence{}
	}
	return list, nil
}

func queryLastSeen(db *sql.DB) ([]model.LastSeen, error) {
	const query = `
		SELECT u.id, u.name, u.picture, ls.detected_at
		FROM user_last_seen ls
		JOIN users u ON u.id = ls.user_id
		ORDER BY ls.detected_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.LastSeen
	for rows.Next() {
		var ls model.LastSeen
		if err := rows.Scan(&ls.UserID, &ls.UserName, &ls.UserPicture, &ls.LastSeenAt); err != nil {
			return nil, err
		}
		list = append(list, ls)
	}
	if list == nil {
		list = []model.LastSeen{}
	}
	return list, nil
}
