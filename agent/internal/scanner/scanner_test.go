package scanner

import "testing"

func TestParseMACAddresses(t *testing.T) {
	// arp-scan --localnet --plain の典型的な出力
	output := `192.168.1.1	aa:bb:cc:dd:ee:01	BUFFALO INC.
192.168.1.10	aa:bb:cc:dd:ee:02	Apple, Inc.
192.168.1.20	AA:BB:CC:DD:EE:03	Intel Corporate
192.168.1.10	aa:bb:cc:dd:ee:02	Apple, Inc. (DUP: 1)

3 packets received by filter, 0 packets dropped by kernel
`

	macs := parseMACAddresses(output)

	if len(macs) != 3 {
		t.Fatalf("expected 3 MACs, got %d: %v", len(macs), macs)
	}

	// すべて小文字に正規化されていること
	for _, mac := range macs {
		if mac != "aa:bb:cc:dd:ee:01" && mac != "aa:bb:cc:dd:ee:02" && mac != "aa:bb:cc:dd:ee:03" {
			t.Errorf("unexpected MAC: %s", mac)
		}
	}
}

func TestParseMACAddresses_Empty(t *testing.T) {
	macs := parseMACAddresses("")
	if len(macs) != 0 {
		t.Fatalf("expected 0 MACs, got %d", len(macs))
	}
}

func TestParseIPNeighMACs(t *testing.T) {
	// "ip neigh show" の典型的な出力
	output := `10.0.2.94 dev eth0 lladdr dc:93:96:1c:5f:fd REACHABLE
10.0.3.1 dev eth0 lladdr aa:bb:cc:dd:ee:01 STALE
10.0.2.50 dev eth0 lladdr AA:BB:CC:DD:EE:02 DELAY
10.0.2.100 dev eth0  FAILED
10.0.2.200 dev eth0  INCOMPLETE
10.0.3.1 dev eth0 lladdr aa:bb:cc:dd:ee:01 STALE
`

	macs := parseIPNeighMACs(output)

	if len(macs) != 3 {
		t.Fatalf("expected 3 MACs, got %d: %v", len(macs), macs)
	}

	expected := map[string]bool{
		"dc:93:96:1c:5f:fd": true,
		"aa:bb:cc:dd:ee:01": true,
		"aa:bb:cc:dd:ee:02": true,
	}
	for _, mac := range macs {
		if !expected[mac] {
			t.Errorf("unexpected MAC: %s", mac)
		}
	}
}

func TestParseIPNeighMACs_ExcludesFailed(t *testing.T) {
	output := `10.0.2.100 dev eth0  FAILED
10.0.2.200 dev eth0  INCOMPLETE
`
	macs := parseIPNeighMACs(output)
	if len(macs) != 0 {
		t.Fatalf("expected 0 MACs, got %d: %v", len(macs), macs)
	}
}

func TestParseCIDRFromIPAddr(t *testing.T) {
	output := "2: eth0    inet 10.0.3.99/22 brd 10.0.3.255 scope global noprefixroute eth0"
	cidr, err := parseCIDRFromIPAddr(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cidr != "10.0.3.99/22" {
		t.Errorf("expected 10.0.3.99/22, got %s", cidr)
	}
}

func TestParseDefaultInterface(t *testing.T) {
	output := "default via 10.0.0.1 dev eth0 proto dhcp src 10.0.3.99 metric 100"
	iface := parseDefaultInterface(output)
	if iface != "eth0" {
		t.Errorf("expected eth0, got %s", iface)
	}
}

func TestParseDefaultInterface_Empty(t *testing.T) {
	iface := parseDefaultInterface("")
	if iface != "" {
		t.Errorf("expected empty, got %s", iface)
	}
}
