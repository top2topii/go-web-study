// Package server는 HTTP 라우팅, 미들웨어, 핸들러를 담는다.
package server

import (
	"log/slog"
	"net/http"
	"time"
)

// Server는 핸들러들이 공유하는 의존성(로거, DB 등)을 보관한다.
// 핸들러를 Server의 메서드로 만들면 전역 변수 없이 의존성에 접근할 수 있다.
type Server struct {
	logger  *slog.Logger
	version string
	started time.Time
	mux     *http.ServeMux
}

// New는 미들웨어까지 감싼 최종 http.Handler를 반환한다.
func New(logger *slog.Logger, version string) http.Handler {
	s := &Server{
		logger:  logger,
		version: version,
		started: time.Now(),
		mux:     http.NewServeMux(),
	}
	s.routes()

	// 미들웨어는 안쪽부터 감싼다. 요청은 바깥 → 안쪽 순서로 통과한다:
	// requestID → logRequests → recoverPanic → mux
	// recoverPanic이 logRequests 안쪽에 있어야 패닉으로 생긴 500도 로그에 남는다.
	var h http.Handler = s.mux
	h = recoverPanic(logger, h)
	h = logRequests(logger, h)
	h = requestID(h)
	return h
}
