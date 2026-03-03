BINARY  := lab-agent
SRC     := ./cmd/agent
GOFLAGS := -trimpath -ldflags="-s -w"

# ローカルビルド
.PHONY: build
build:
	go build $(GOFLAGS) -o $(BINARY) $(SRC)

# Raspberry Pi 3B 向けクロスコンパイル (ARM 32-bit)
.PHONY: build-pi
build-pi:
	GOOS=linux GOARCH=arm GOARM=7 go build $(GOFLAGS) -o $(BINARY)-linux-arm $(SRC)

# Raspberry Pi 3B+ / 4 向け (ARM 64-bit)
.PHONY: build-pi64
build-pi64:
	GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -o $(BINARY)-linux-arm64 $(SRC)

# テスト
.PHONY: test
test:
	go test ./...

# クリーン
.PHONY: clean
clean:
	rm -f $(BINARY) $(BINARY)-linux-arm $(BINARY)-linux-arm64
