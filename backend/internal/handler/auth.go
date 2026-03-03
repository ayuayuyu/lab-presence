package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
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

		if time.Now().Unix() > claims.Exp {
			http.Error(w, `{"error":"token expired"}`, http.StatusUnauthorized)
			return
		}

		if claims.HD != allowedDomain {
			http.Error(w, `{"error":"domain not allowed"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}
