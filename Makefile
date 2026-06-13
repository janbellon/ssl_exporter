APP := ssl_exporter
BUILD_DIR := build
GO_ENTRYPOINT := ./cmd/ssl_exporter

.PHONY: build clean

build: clean \
	linux-amd64 \
	linux-arm64 \
	darwin-amd64 \
	darwin-arm64 \
	windows-amd64

clean:
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)

linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP)-linux-amd64 $(GO_ENTRYPOINT)

linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP)-linux-arm64 $(GO_ENTRYPOINT)

darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP)-darwin-amd64 $(GO_ENTRYPOINT)

darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP)-darwin-arm64 $(GO_ENTRYPOINT)

windows-amd64:
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP)-windows-amd64.exe $(GO_ENTRYPOINT)