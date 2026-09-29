package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 테스트에서는 로그를 버린다. 디버깅할 땐 io.Discard를 os.Stderr로 바꾸면 된다.
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHealth(t *testing.T) {
	h := New(testLogger(), "test")

	// httptest.NewRecorder는 실제 네트워크 없이 응답을 메모리에 기록한다.
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("X-Request-ID header missing")
	}

	var body healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Status != "ok" || body.Version != "test" {
		t.Errorf("body = %+v", body)
	}
}

// 테이블 주도 테스트: 입력/기대값을 표로 나열하고 t.Run으로 하위 테스트를 돌린다.
func TestRouting(t *testing.T) {
	h := New(testLogger(), "test")

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"health GET", http.MethodGet, "/api/health", http.StatusOK},
		{"health HEAD", http.MethodHead, "/api/health", http.StatusOK},
		{"health POST is 405", http.MethodPost, "/api/health", http.StatusMethodNotAllowed},
		{"unknown path is 404", http.MethodGet, "/api/nope", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestRequestIDPropagation(t *testing.T) {
	h := New(testLogger(), "test")
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	req.Header.Set("X-Request-ID", "abc123")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != "abc123" {
		t.Errorf("X-Request-ID = %q, want abc123", got)
	}
}

func TestRecoverPanic(t *testing.T) {
	boom := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	h := recoverPanic(testLogger(), boom)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	var body errorBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error == "" {
		t.Error("error message missing")
	}
}
