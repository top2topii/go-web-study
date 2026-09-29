VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

.PHONY: run test build fmt vet check

run:
	go run ./cmd/server

test:
	go test ./...

build:
	go build -ldflags "$(LDFLAGS)" -o bin/server ./cmd/server

fmt:
	gofmt -w .

vet:
	go vet ./...

# 커밋 전 확인용: 포맷 검사 + vet + 테스트(race detector 포함)
check:
	@test -z "$$(gofmt -l .)" || (echo "gofmt needed:"; gofmt -l .; exit 1)
	go vet ./...
	go test -race ./...
