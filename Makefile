APP=ssl_exporter

build:
	go build -o build/$(APP) ./cmd/ssl_exporter

run:
	go run ./cmd/ssl_exporter

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run

clean:
	rm -rf build

.PHONY: build run test fmt lint docker-build docker-run compose-up compose-down clean