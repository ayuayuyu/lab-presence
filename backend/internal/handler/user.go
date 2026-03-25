package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ayuayuyu/lab-presence/backend/internal/model"
)

// GET /api/users  — ユーザー一覧
// POST /api/users — ユーザー登録
func HandleUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listUsers(db, w)
		case http.MethodPost:
			createUser(db, w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func listUsers(db *sql.DB, w http.ResponseWriter) {
	rows, err := db.Query(`SELECT id, name, COALESCE(email, ''), picture, COALESCE(student_id, ''), created_at FROM users ORDER BY id`)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Picture, &u.StudentID, &u.CreatedAt); err != nil {
			http.Error(w, "scan failed", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	if users == nil {
		users = []model.User{}
	}
	writeJSON(w, http.StatusOK, users)
}

func createUser(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
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
	err := db.QueryRow(
		`INSERT INTO users (name, student_id) VALUES ($1, $2) RETURNING id, name, COALESCE(email, ''), picture, COALESCE(student_id, ''), created_at`,
		req.Name, req.StudentID,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Picture, &u.StudentID, &u.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			http.Error(w, "student_id already exists", http.StatusConflict)
			return
		}
		http.Error(w, "insert failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, u)
}

// PUT /api/users/me — 表示名変更
func HandleUserMe(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		email := extractEmail(r)
		if email == "" {
			http.Error(w, `{"error":"email not found in token"}`, http.StatusBadRequest)
			return
		}

		var req model.UpdateUserNameRequest
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
		err := db.QueryRow(
			`UPDATE users SET name = $1 WHERE email = $2 RETURNING id, name, COALESCE(email, ''), picture, COALESCE(student_id, ''), created_at`,
			req.Name, email,
		).Scan(&u.ID, &u.Name, &u.Email, &u.Picture, &u.StudentID, &u.CreatedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "user not found", http.StatusNotFound)
				return
			}
			http.Error(w, "update failed", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, u)
	}
}
