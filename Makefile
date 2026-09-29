VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

SQLC := go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1

.PHONY: run test build fmt vet check generate seed

run:
	go run ./cmd/server

# internal/db/queries/*.sql → internal/store/*.go
generate:
	$(SQLC) generate

# DB 파일을 지우고 스키마 + 데모 데이터로 다시 만든다
seed:
	go run ./cmd/seed

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
