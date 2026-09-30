GO_PKGS := ./cmd/... ./internal/...
GO_DIRS := cmd internal

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)
SQLC := go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
WEB := pnpm -C web

.PHONY: run dev web test build fmt vet check go-check web-check generate seed

run:
	go run ./cmd/server

web:
	$(WEB) dev

# Go API(:8080) + Vite(:5173)를 함께 띄운다. 브라우저는 http://localhost:5173
# Ctrl+C 한 번에 둘 다 종료(kill 0 = 같은 프로세스 그룹 전체).
dev:
	@trap 'kill 0' INT TERM; go run ./cmd/server & $(WEB) dev & wait

# internal/db/queries/*.sql → internal/store/*.go
generate:
	$(SQLC) generate

# DB 파일을 지우고 스키마 + 데모 데이터로 다시 만든다
seed:
	go run ./cmd/seed

test:
	go test $(GO_PKGS)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/server ./cmd/server

fmt:
	gofmt -w $(GO_DIRS)

vet:
	go vet $(GO_PKGS)

# 커밋 전 확인용
check: go-check web-check

go-check:
	@test -z "$$(gofmt -l $(GO_DIRS))" || (echo "gofmt needed:"; gofmt -l $(GO_DIRS); exit 1)
	go vet $(GO_PKGS)
	go test -race $(GO_PKGS)

# tsc 타입 검사 + vite 빌드 + oxlint
web-check:
	$(WEB) lint
	$(WEB) build
