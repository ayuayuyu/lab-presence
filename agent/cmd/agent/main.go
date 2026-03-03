package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayuayuyu/lab-presence/agent/internal/scanner"
	"github.com/ayuayuyu/lab-presence/agent/internal/sender"
)

func main() {
	backendURL := flag.String("backend", "http://localhost:8080", "backend API base URL")
	iface := flag.String("iface", "", "network interface for arp-scan (default: auto)")
	interval := flag.Duration("interval", 2*time.Minute, "scan interval")
	flag.Parse()

	// 環境変数でも上書き可能
	if env := os.Getenv("BACKEND_URL"); env != "" {
		*backendURL = env
	}
	if env := os.Getenv("SCAN_INTERFACE"); env != "" {
		*iface = env
	}

	log.Printf("agent started: backend=%s iface=%q interval=%s", *backendURL, *iface, *interval)

	// 初回は即実行
	runOnce(*backendURL, *iface)

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			runOnce(*backendURL, *iface)
		case s := <-sig:
			log.Printf("received %s, shutting down", s)
			return
		}
	}
}

func runOnce(backendURL, iface string) {
	macs, err := scanner.RunCombinedScan(iface)
	if err != nil {
		log.Printf("scan error: %v", err)
		return
	}

	log.Printf("detected %d MAC address(es)", len(macs))

	if len(macs) == 0 {
		return
	}

	if err := sender.Send(backendURL, macs); err != nil {
		log.Printf("send error: %v", err)
		return
	}

	log.Printf("sent %d MACs to backend", len(macs))
}
