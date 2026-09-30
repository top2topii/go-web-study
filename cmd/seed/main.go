// Command seed는 DB를 초기화하고 데모 데이터를 채운다.
//
//	go run ./cmd/seed -db data/shopadmin.db
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"shopadmin/internal/auth"
	"shopadmin/internal/db"
	"shopadmin/internal/seed"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	path := flag.String("db", "data/shopadmin.db", "SQLite 파일 경로")
	flag.Parse()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// 기존 DB 파일 삭제(WAL 모드의 -wal, -shm 보조 파일 포함)
	for _, suffix := range []string{"", "-wal", "-shm"} {
		if err := os.Remove(*path + suffix); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(*path), 0o755); err != nil {
		return err
	}

	conn, err := db.Open(ctx, *path)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := db.Migrate(ctx, conn, logger); err != nil {
		return err
	}
	opt := seed.DefaultOptions()
	// 비밀번호는 환경 변수로 받는다. 없으면 무작위로 만들어 한 번만 출력한다.
	opt.Password = os.Getenv("SEED_PASSWORD")
	generated := opt.Password == ""
	if generated {
		opt.Password = auth.NewToken()[:16]
	}

	res, err := seed.Run(ctx, conn, opt)
	if err != nil {
		return err
	}
	logger.Info("seeded", "db", *path, "users", res.Users,
		"categories", res.Categories, "products", res.Products,
		"customers", res.Customers, "orders", res.Orders, "items", res.Items)

	fmt.Println("\n로그인 계정:")
	for _, u := range seed.SeedUsers {
		fmt.Printf("  %-20s (%s)\n", u.Email, u.Role)
	}
	if generated {
		fmt.Printf("  비밀번호: %s   ← 무작위 생성. 고정하려면 SEED_PASSWORD=... make seed\n", opt.Password)
	} else {
		fmt.Println("  비밀번호: $SEED_PASSWORD")
	}
	return nil
}
