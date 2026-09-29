// Package db는 SQLite 연결과 스키마 마이그레이션을 담당한다.
package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"

	// 드라이버는 import만 해도 init()에서 database/sql에 "sqlite"라는 이름으로 등록된다.
	// 이름을 쓰지 않으므로 빈 식별자(_)로 import한다.
	_ "modernc.org/sqlite"
)

// go:embed는 빌드 시점에 파일을 바이너리 안에 넣는다. 배포 시 SQL 파일을 따로 챙길 필요가 없다.
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// Open은 운영용 pragma가 적용된 SQLite 커넥션 풀을 연다.
//
// *sql.DB는 커넥션 하나가 아니라 "풀"이다. 여러 고루틴이 동시에 써도 안전하며
// 앱 전체에서 하나만 만들어 공유한다.
func Open(ctx context.Context, path string) (*sql.DB, error) {
	q := url.Values{}
	// _pragma는 풀의 모든 커넥션이 열릴 때마다 적용된다.
	q.Add("_pragma", "foreign_keys(1)")    // SQLite는 기본적으로 FK 검사를 안 한다!
	q.Add("_pragma", "journal_mode(WAL)")  // 읽기와 쓰기가 서로를 막지 않음
	q.Add("_pragma", "busy_timeout(5000)") // 잠금 충돌 시 즉시 실패 대신 5초 대기
	q.Add("_pragma", "synchronous(NORMAL)")
	// 트랜잭션 시작 시 바로 쓰기 잠금을 잡아 "database is locked" 교착을 피한다.
	q.Set("_txlock", "immediate")
	// time.Time을 SQLite 날짜 함수가 이해하는 형식으로 저장한다.
	q.Set("_time_format", "sqlite")

	db, err := sql.Open("sqlite", "file:"+path+"?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// sql.Open은 실제로 연결하지 않는다. Ping으로 파일 접근 가능 여부를 확인한다.
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite %s: %w", path, err)
	}
	return db, nil
}

// Migrate는 embed된 migrations/*.sql 중 아직 적용 안 된 것을 순서대로 적용한다.
// 적용 이력은 goose_db_version 테이블에 기록된다.
func Migrate(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	sub, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	p, err := goose.NewProvider(database.DialectSQLite3, db, sub)
	if err != nil {
		return fmt.Errorf("goose provider: %w", err)
	}
	results, err := p.Up(ctx)
	if err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}
	for _, r := range results {
		logger.Info("migration applied", "version", r.Source.Version, "file", r.Source.Path, "dur", r.Duration)
	}
	return nil
}
