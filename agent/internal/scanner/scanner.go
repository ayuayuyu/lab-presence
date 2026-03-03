package scanner

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// MACアドレスの正規表現 (aa:bb:cc:dd:ee:ff)
var macRe = regexp.MustCompile(`([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}`)

// RunCombinedScan は arp-scan でMACアドレスを収集する。
func RunCombinedScan(iface string) ([]string, error) {
	return RunArpScan(iface)
}

// RunArpScan は arp-scan を実行してネットワーク上のMACアドレスを収集する。
// iface が空の場合はデフォルトインターフェースが使われる。
func RunArpScan(iface string) ([]string, error) {
	args := []string{"--localnet", "--plain"}
	if iface != "" {
		args = append(args, "--interface", iface)
	}

	cmd := exec.Command("arp-scan", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("arp-scan: %w", err)
	}

	return parseMACAddresses(string(out)), nil
}

// parseMACAddresses は arp-scan の出力から重複なしのMACアドレス一覧を抽出する。
func parseMACAddresses(output string) []string {
	seen := make(map[string]struct{})
	var macs []string

	s := bufio.NewScanner(strings.NewReader(output))
	for s.Scan() {
		line := s.Text()
		mac := macRe.FindString(line)
		if mac == "" {
			continue
		}
		mac = strings.ToLower(mac)
		if _, ok := seen[mac]; ok {
			continue
		}
		seen[mac] = struct{}{}
		macs = append(macs, mac)
	}

	return macs
}
