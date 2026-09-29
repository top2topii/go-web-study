package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

// 미들웨어의 모양은 항상 func(http.Handler) http.Handler 이다.
// 다음 핸들러(next)를 받아, 앞뒤로 일을 하는 새 핸들러를 돌려준다.

// context 키는 충돌을 막기 위해 외부에서 만들 수 없는 비공개 타입으로 정의한다.
type ctxKey int

const requestIDKey ctxKey = iota

// requestID는 요청마다 고유 ID를 만들어 context와 응답 헤더에 넣는다.
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFrom은 context에서 요청 ID를 꺼낸다. 없으면 빈 문자열.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string) // 타입 단언(type assertion), 실패해도 panic 없음
	return id
}

// statusRecorder는 핸들러가 쓴 상태 코드와 바이트 수를 기록한다.
// http.ResponseWriter를 임베딩해서 나머지 메서드는 그대로 위임한다.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) Write(b []byte) (int, error) {
	if sr.status == 0 {
		sr.status = http.StatusOK // WriteHeader 없이 Write하면 암묵적으로 200
	}
	n, err := sr.ResponseWriter.Write(b)
	sr.bytes += n
	return n, err
}

// Unwrap은 http.ResponseController가 원래 ResponseWriter의 기능(Flush 등)에
// 접근할 수 있게 한다.
func (sr *statusRecorder) Unwrap() http.ResponseWriter { return sr.ResponseWriter }

// logRequests는 요청마다 메서드, 경로, 상태, 소요 시간을 한 줄로 남긴다.
func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		if rec.status == 0 {
			rec.status = http.StatusOK
		}
		logger.LogAttrs(r.Context(), slog.LevelInfo, "request",
			slog.String("id", RequestIDFrom(r.Context())),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.status),
			slog.Int("bytes", rec.bytes),
			slog.Duration("dur", time.Since(start)),
		)
	})
}

// recoverPanic은 핸들러의 panic을 잡아 500으로 바꾼다.
// net/http도 panic을 잡긴 하지만 커넥션을 끊어버리므로 클라이언트는 응답을 못 받는다.
func recoverPanic(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// defer는 함수가 끝날 때(정상 종료든 panic이든) 실행된다.
		defer func() {
			if v := recover(); v != nil {
				// http.ErrAbortHandler는 의도된 중단이므로 다시 던진다.
				if v == http.ErrAbortHandler {
					panic(v)
				}
				logger.Error("panic recovered",
					"id", RequestIDFrom(r.Context()),
					"panic", v,
					"stack", string(debug.Stack()),
				)
				w.Header().Set("Connection", "close")
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
