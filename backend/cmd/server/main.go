package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ayuayuyu/lab-presence/backend/internal/db"
	"github.com/ayuayuyu/lab-presence/backend/internal/handler"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	conn, err := db.Connect(databaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer conn.Close()

	hub := handler.NewHub()

	mux := http.NewServeMux()

	// 認証不要（エージェント用・ヘルスチェック）
	mux.HandleFunc("/api/scan", handler.HandleScan(conn, hub))

	// WebSocket（認証不要）
	mux.HandleFunc("/ws/presence", hub.HandleWS())

	// 認証必要
	mux.HandleFunc("/api/presence/last-seen", handler.Auth(handler.HandleLastSeen(conn)))
	mux.HandleFunc("/api/presence", handler.Auth(handler.HandlePresence(conn)))
	mux.HandleFunc("/api/users", handler.Auth(handler.HandleUsers(conn)))
	mux.HandleFunc("/api/devices", handler.Auth(handler.HandleDevices(conn)))
	mux.HandleFunc("/api/devices/", handler.Auth(handler.HandleDevice(conn)))

	// ヘルスチェック
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Printf("server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler.CORS(mux)))
}
