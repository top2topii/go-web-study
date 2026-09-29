// Command server는 shopadmin의 HTTP 서버 진입점이다.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"shopadmin/internal/db"
	"shopadmin/internal/server"
)

// version은 빌드 시 -ldflags "-X main.version=..." 으로 덮어쓴다.
var version = "dev"

func main() {
	// main은 최대한 얇게 두고 실제 로직은 run에 둔다.
	// run은 error를 반환하므로 테스트하기 쉽고, 종료 코드 처리는 여기 한 곳에서만 한다.
	if err := run(context.Background(), os.Args[1:], os.Getenv, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

type config struct {
	addr     string
	dbPath   string
	logLevel slog.Level
	logJSON  bool
}

// loadConfig는 환경 변수를 기본값으로, 명령줄 플래그를 우선으로 설정을 읽는다.
// os.Getenv 대신 getenv 함수를 주입받아 테스트에서 가짜 환경을 넣을 수 있다.
func loadConfig(args []string, getenv func(string) string) (config, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	addr := fs.String("addr", envOr(getenv, "ADDR", "127.0.0.1:8080"), "listen address")
	dbPath := fs.String("db", envOr(getenv, "DB_PATH", "data/shopadmin.db"), "SQLite 파일 경로")
	level := fs.String("log-level", envOr(getenv, "LOG_LEVEL", "info"), "debug|info|warn|error")
	logJSON := fs.Bool("log-json", getenv("LOG_JSON") == "1", "JSON 로그 출력")

	if err := fs.Parse(args); err != nil {
		return config{}, err
	}

	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(*level)); err != nil {
		return config{}, fmt.Errorf("invalid log level %q: %w", *level, err)
	}
	return config{addr: *addr, dbPath: *dbPath, logLevel: lvl, logJSON: *logJSON}, nil
}

func envOr(getenv func(string) string, key, fallback string) string {
	if v := getenv(key); v != "" {
		return v
	}
	return fallback
}

func newLogger(w io.Writer, cfg config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.logLevel}
	if cfg.logJSON {
		return slog.New(slog.NewJSONHandler(w, opts))
	}
	return slog.New(slog.NewTextHandler(w, opts))
}

func run(ctx context.Context, args []string, getenv func(string) string, stdout io.Writer) error {
	// Ctrl+C(SIGINT)나 SIGTERM을 받으면 ctx가 취소된다.
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig(args, getenv)
	if err != nil {
		return err
	}
	logger := newLogger(stdout, cfg)

	if err := os.MkdirAll(filepath.Dir(cfg.dbPath), 0o755); err != nil {
		return err
	}
	conn, err := db.Open(ctx, cfg.dbPath)
	if err != nil {
		return err
	}
	// defer는 LIFO(나중에 등록한 것이 먼저 실행). Shutdown이 끝난 뒤 DB가 닫힌다.
	defer conn.Close()

	// 서버 시작 전에 스키마를 최신으로 맞춘다.
	if err := db.Migrate(ctx, conn, logger); err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    cfg.addr,
		Handler: server.New(logger, conn, version),
		// 타임아웃이 없으면 느린 클라이언트가 커넥션을 무한정 붙잡을 수 있다(Slowloris).
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// ListenAndServe는 블로킹이므로 고루틴에서 돌리고, 에러는 채널로 받는다.
	// 버퍼 1: 아무도 받지 않아도 고루틴이 막히지 않게 한다.
	errCh := make(chan error, 1)
	go func() {
		logger.Info("server listening", "addr", "http://"+cfg.addr, "version", version)
		// Shutdown이 호출되면 ErrServerClosed가 반환되는데, 이는 정상 종료다.
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		// 포트 충돌 등으로 서버가 시작조차 못 한 경우
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	// 진행 중인 요청이 끝날 때까지 최대 10초 기다린다.
	// 부모 ctx는 이미 취소됐으므로 새 context를 쓴다.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	logger.Info("server stopped")
	return nil
}
