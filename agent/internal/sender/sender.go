package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ScanPayload struct {
	MACAddresses []string `json:"mac_addresses"`
}

// Send はMACアドレス一覧をバックエンドの /api/scan に送信する。
func Send(backendURL string, macs []string) error {
	payload := ScanPayload{MACAddresses: macs}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(backendURL+"/api/scan", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
