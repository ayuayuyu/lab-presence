package handler

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/ayuayuyu/lab-presence/backend/internal/model"
)

type idTokenPayload struct {
	Email string `json:"email"`
	HD    string `json:"hd"`
	Exp   int64  `json:"exp"`
}

func parseIDToken(token string) (*idTokenPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, http.ErrNotSupported
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims idTokenPayload
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	return &claims, nil
}

// isAdminEmail は指定のメールアドレスが ADMIN_EMAILS に含まれるかを返す
func isAdminEmail(email string) bool {
	for _, e := range strings.Split(os.Getenv("ADMIN_EMAILS"), ",") {
		if strings.TrimSpace(e) == email {
			return true
		}
	}
	return false
}

// extractEmail は Authorization ヘッダーの JWT からメールアドレスを取得する
func extractEmail(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := parseIDToken(token)
	if err != nil {
		return ""
	}
	return claims.Email
}

// AuthWithAdminWrite は GET には Auth を、書き込み系メソッド (POST/PUT/DELETE) には AdminAuth を適用する
func AuthWithAdminWrite(next http.HandlerFunc) http.HandlerFunc {
	authHandler := Auth(next)
	adminHandler := AdminAuth(next)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			authHandler.ServeHTTP(w, r)
		} else {
			adminHandler.ServeHTTP(w, r)
		}
	}
}

// AdminAuth は Auth に加えて、メールアドレスが ADMIN_EMAILS に含まれるかを確認するミドルウェア
func AdminAuth(next http.HandlerFunc) http.HandlerFunc {
	allowedDomain := os.Getenv("ALLOWED_DOMAIN")
	if allowedDomain == "" {
		allowedDomain = "pluslab.org"
	}

	adminEmails := make(map[string]bool)
	for _, email := range strings.Split(os.Getenv("ADMIN_EMAILS"), ",") {
		email = strings.TrimSpace(email)
		if email != "" {
			adminEmails[email] = true
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"authorization required"}`, http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := parseIDToken(token)
		if err != nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		if claims.HD != allowedDomain {
			http.Error(w, `{"error":"domain not allowed"}`, http.StatusForbidden)
			return
		}

		if !adminEmails[claims.Email] {
			http.Error(w, `{"error":"admin access required"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// Auth は JWT の payload をデコードして hd ドメインを確認するミドルウェア
func Auth(next http.HandlerFunc) http.HandlerFunc {
	allowedDomain := os.Getenv("ALLOWED_DOMAIN")
	if allowedDomain == "" {
		allowedDomain = "pluslab.org"
	}

	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"authorization required"}`, http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := parseIDToken(token)
		if err != nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		if claims.HD != allowedDomain {
			http.Error(w, `{"error":"domain not allowed"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// POST /api/auth/me — Googleログイン時にユーザーレコードを upsert
func HandleAuthMe(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		email := extractEmail(r)
		if email == "" {
			http.Error(w, `{"error":"email not found in token"}`, http.StatusBadRequest)
			return
		}

		var req model.SyncAuthMeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}

		var u model.User
		err := db.QueryRow(`
			INSERT INTO users (name, email, picture)
			VALUES ($1, $2, $3)
			ON CONFLICT (email) DO UPDATE SET picture = EXCLUDED.picture
			RETURNING id, name, email, picture, COALESCE(student_id, ''), created_at
		`, req.Name, email, req.Picture).Scan(&u.ID, &u.Name, &u.Email, &u.Picture, &u.StudentID, &u.CreatedAt)
		if err != nil {
			http.Error(w, "upsert failed", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, u)
	}
}
