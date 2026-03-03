package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ayuayuyu/lab-presence/backend/internal/model"
)

// GET  /api/devices — デバイス一覧
// POST /api/devices — デバイス登録
func HandleDevices(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listDevices(db, w)
		case http.MethodPost:
			createDevice(db, w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func listDevices(db *sql.DB, w http.ResponseWriter) {
	rows, err := db.Query(`SELECT id, user_id, mac_address::text, label, created_at FROM devices ORDER BY id`)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var devices []model.Device
	for rows.Next() {
		var d model.Device
		if err := rows.Scan(&d.ID, &d.UserID, &d.MACAddress, &d.Label, &d.CreatedAt); err != nil {
			http.Error(w, "scan failed", http.StatusInternalServerError)
			return
		}
		devices = append(devices, d)
	}

	if devices == nil {
		devices = []model.Device{}
	}
	writeJSON(w, http.StatusOK, devices)
}

func createDevice(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var req model.CreateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.MACAddress = strings.TrimSpace(strings.ToLower(req.MACAddress))
	if req.MACAddress == "" || req.UserID == 0 {
		http.Error(w, "user_id and mac_address are required", http.StatusBadRequest)
		return
	}

	var d model.Device
	err := db.QueryRow(
		`INSERT INTO devices (user_id, mac_address, label) VALUES ($1, $2::macaddr, $3)
		 RETURNING id, user_id, mac_address::text, label, created_at`,
		req.UserID, req.MACAddress, req.Label,
	).Scan(&d.ID, &d.UserID, &d.MACAddress, &d.Label, &d.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			http.Error(w, "mac_address already registered", http.StatusConflict)
			return
		}
		if strings.Contains(err.Error(), "foreign") {
			http.Error(w, "user not found", http.StatusBadRequest)
			return
		}
		http.Error(w, "insert failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, d)
}

// PUT    /api/devices/{id} — デバイス更新
// DELETE /api/devices/{id} — デバイス削除
func HandleDevice(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/devices/")
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			http.Error(w, "invalid device id", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodPut:
			updateDevice(db, w, r, id)
		case http.MethodDelete:
			deleteDevice(db, w, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func updateDevice(db *sql.DB, w http.ResponseWriter, r *http.Request, id int) {
	var req model.UpdateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.MACAddress = strings.TrimSpace(strings.ToLower(req.MACAddress))
	if req.MACAddress == "" || req.UserID == 0 {
		http.Error(w, "user_id and mac_address are required", http.StatusBadRequest)
		return
	}

	var d model.Device
	err := db.QueryRow(
		`UPDATE devices SET mac_address=$1::macaddr, label=$2, user_id=$3 WHERE id=$4
		 RETURNING id, user_id, mac_address::text, label, created_at`,
		req.MACAddress, req.Label, req.UserID, id,
	).Scan(&d.ID, &d.UserID, &d.MACAddress, &d.Label, &d.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "device not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "unique") {
			http.Error(w, "mac_address already registered", http.StatusConflict)
			return
		}
		if strings.Contains(err.Error(), "foreign") {
			http.Error(w, "user not found", http.StatusBadRequest)
			return
		}
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, d)
}

func deleteDevice(db *sql.DB, w http.ResponseWriter, id int) {
	result, err := db.Exec(`DELETE FROM devices WHERE id=$1`, id)
	if err != nil {
		http.Error(w, "delete failed", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "device not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
