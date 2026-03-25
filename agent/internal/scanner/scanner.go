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

// CIDR表記の正規表現 (例: 10.0.3.99/22)
var cidrRe = regexp.MustCompile(`\d+\.\d+\.\d+\.\d+/\d+`)

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

// parseIPNeighMACs は "ip neigh show" の出力からMACアドレスを抽出する。
// FAILED / INCOMPLETE 状態の行は lladdr を持たないため自動的にスキップされる。
func parseIPNeighMACs(output string) []string {
	seen := make(map[string]struct{})
	var macs []string

	s := bufio.NewScanner(strings.NewReader(output))
	for s.Scan() {
		line := s.Text()
		if !strings.Contains(line, "lladdr") {
			continue
		}
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

// parseCIDRFromIPAddr は "ip addr show <iface>" の出力から inet の CIDR を抽出する。
func parseCIDRFromIPAddr(output string) (string, error) {
	s := bufio.NewScanner(strings.NewReader(output))
	for s.Scan() {
		line := s.Text()
		if !strings.Contains(line, "inet ") {
			continue
		}
		cidr := cidrRe.FindString(line)
		if cidr != "" {
			return cidr, nil
		}
	}
	return "", fmt.Errorf("no CIDR found in output")
}

// parseDefaultInterface は "ip route show default" の出力からデフォルトインターフェース名を抽出する。
func parseDefaultInterface(output string) string {
	fields := strings.Fields(output)
	for i, f := range fields {
		if f == "dev" && i+1 < len(fields) {
			return fields[i+1]
		}
	}
	return ""
}
