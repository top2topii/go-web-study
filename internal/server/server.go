// Package server는 HTTP 라우팅, 미들웨어, 핸들러를 담는다.
package server

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"shopadmin/internal/store"
)

// Config는 서버 동작 옵션이다. 필드가 늘어나도 New의 시그니처는 그대로다.
type Config struct {
	Version string
	// CookieSecure: 세션 쿠키에 Secure 속성 부여(HTTPS에서만 전송).
	// 브라우저는 http://localhost를 보안 컨텍스트로 취급하므로 로컬 개발에서도 true로 동작한다.
	CookieSecure bool
	SessionTTL   time.Duration
}

// Server는 핸들러들이 공유하는 의존성(로거, DB 등)을 보관한다.
// 핸들러를 Server의 메서드로 만들면 전역 변수 없이 의존성에 접근할 수 있다.
type Server struct {
	logger  *slog.Logger
	db      *sql.DB        // 트랜잭션, Ping 등 저수준 작업용
	q       *store.Queries // sqlc가 생성한 쿼리 메서드
	cfg     Config
	started time.Time
	mux     *http.ServeMux
}

// New는 미들웨어까지 감싼 최종 http.Handler를 반환한다.
func New(logger *slog.Logger, db *sql.DB, cfg Config) http.Handler {
	if cfg.SessionTTL == 0 {
		cfg.SessionTTL = 7 * 24 * time.Hour
	}
	s := &Server{
		logger:  logger,
		db:      db,
		q:       store.New(db),
		cfg:     cfg,
		started: time.Now(),
		mux:     http.NewServeMux(),
	}
	s.routes()

	// 미들웨어는 안쪽부터 감싼다. 요청은 바깥 → 안쪽 순서로 통과한다:
	// requestID → logRequests → recoverPanic → CSRF 방어 → 세션 로드 → mux
	// recoverPanic이 logRequests 안쪽에 있어야 패닉으로 생긴 500도 로그에 남는다.
	var h http.Handler = s.mux
	h = s.loadSession(h)
	// Go 1.25+ 표준 CSRF 방어: 다른 사이트에서 온 POST/PUT/DELETE 등을 403으로 거부한다.
	// 브라우저가 붙이는 Sec-Fetch-Site / Origin 헤더를 본다.
	h = http.NewCrossOriginProtection().Handler(h)
	h = recoverPanic(logger, h)
	h = logRequests(logger, h)
	h = requestID(h)
	return h
}
