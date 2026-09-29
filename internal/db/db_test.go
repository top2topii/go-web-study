package db

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
)

// openTest는 테스트마다 격리된 임시 DB를 만든다.
// t.TempDir()은 테스트가 끝나면 자동 삭제되고, t.Cleanup은 종료 시 실행할 정리 함수를 등록한다.
func openTest(t *testing.T) *sql.DB {
	t.Helper() // 실패 시 이 함수가 아니라 호출한 줄을 가리키게 한다
	ctx := context.Background()
	conn, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := Migrate(ctx, conn, slog.New(slog.NewTextHandler(io.Discard, nil))); err != nil {
		t.Fatal(err)
	}
	return conn
}

func TestMigrateIsIdempotent(t *testing.T) {
	conn := openTest(t)
	// 두 번째 실행은 적용할 것이 없으므로 에러 없이 끝나야 한다.
	if err := Migrate(context.Background(), conn, slog.New(slog.NewTextHandler(io.Discard, nil))); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
}

func TestPragmasApplied(t *testing.T) {
	conn := openTest(t)
	var fk int
	var mode string
	if err := conn.QueryRow("PRAGMA foreign_keys").Scan(&fk); err != nil {
		t.Fatal(err)
	}
	if err := conn.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if fk != 1 || mode != "wal" {
		t.Errorf("foreign_keys=%d journal_mode=%s", fk, mode)
	}
}

func TestConstraints(t *testing.T) {
	conn := openTest(t)
	tests := []struct {
		name string
		sql  string
	}{
		{"FK: 존재하지 않는 고객의 주문", `INSERT INTO orders (customer_id) VALUES (999)`},
		{"CHECK: 음수 가격", `INSERT INTO products (sku, name, price) VALUES ('X', 'x', -1)`},
		{"CHECK: 잘못된 상태", `INSERT INTO customers (email, name, status) VALUES ('a@b.c', 'a', 'weird')`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := conn.Exec(tt.sql); err == nil {
				t.Error("expected constraint error, got nil")
			}
		})
	}
}
