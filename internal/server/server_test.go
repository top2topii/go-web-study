package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"shopadmin/internal/db"
	"shopadmin/internal/seed"
	"shopadmin/internal/store"
)

// 테스트에서는 로그를 버린다. 디버깅할 땐 io.Discard를 os.Stderr로 바꾸면 된다.
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

const testProducts = 15

// newTestServer는 임시 SQLite 파일에 마이그레이션 + 소량 시드를 넣고 핸들러를 만든다.
func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	ctx := context.Background()
	conn, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := db.Migrate(ctx, conn, testLogger()); err != nil {
		t.Fatal(err)
	}
	if _, err := seed.Run(ctx, conn, seed.Options{
		Customers: 20, Products: testProducts, Days: 10, Now: time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}
	return New(testLogger(), conn, "test")
}

// do는 요청을 보내고 응답 레코더를 돌려주는 작은 헬퍼.
func do(t *testing.T, h http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

// decode는 제네릭 함수: 호출할 때 decode[Page[...]](t, rec)처럼 타입을 지정한다.
func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(rec.Body).Decode(&v); err != nil {
		t.Fatalf("decode: %v (body=%q)", err, rec.Body.String())
	}
	return v
}

func TestHealth(t *testing.T) {
	h := newTestServer(t)
	rec := do(t, h, http.MethodGet, "/api/health")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("X-Request-ID header missing")
	}
	body := decode[healthResponse](t, rec)
	if body.Status != "ok" || body.DB != "ok" || body.Version != "test" {
		t.Errorf("body = %+v", body)
	}
}

// 테이블 주도 테스트: 입력/기대값을 표로 나열하고 t.Run으로 하위 테스트를 돌린다.
func TestRouting(t *testing.T) {
	h := newTestServer(t)

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
		{"products bad page", http.MethodGet, "/api/products?page=0", http.StatusBadRequest},
		{"products page_size too big", http.MethodGet, "/api/products?page_size=1000", http.StatusBadRequest},
		{"product found", http.MethodGet, "/api/products/1", http.StatusOK},
		{"product not found", http.MethodGet, "/api/products/99999", http.StatusNotFound},
		{"product bad id", http.MethodGet, "/api/products/abc", http.StatusBadRequest},
		{"revenue bad days", http.MethodGet, "/api/stats/revenue?days=0", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rec := do(t, h, tt.method, tt.path); rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body=%s)", rec.Code, tt.wantStatus, rec.Body)
			}
		})
	}
}

func TestListProductsPagination(t *testing.T) {
	h := newTestServer(t)

	first := decode[Page[store.ListProductsRow]](t, do(t, h, http.MethodGet, "/api/products?page=1&page_size=5"))
	if len(first.Items) != 5 || first.Total != testProducts {
		t.Fatalf("page1: len=%d total=%d", len(first.Items), first.Total)
	}
	if first.Items[0].CategoryName == nil {
		t.Error("category_name should be joined")
	}

	second := decode[Page[store.ListProductsRow]](t, do(t, h, http.MethodGet, "/api/products?page=2&page_size=5"))
	if second.Items[0].ID == first.Items[0].ID {
		t.Error("page 2 returned same items as page 1")
	}

	// 범위를 벗어난 페이지는 null이 아니라 빈 배열이어야 한다(emit_empty_slices).
	rec := do(t, h, http.MethodGet, "/api/products?page=100")
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	if string(raw["items"]) != "[]" {
		t.Errorf("items = %s, want []", raw["items"])
	}
}

func TestRevenue(t *testing.T) {
	h := newTestServer(t)
	rows := decode[[]store.RevenueByDayRow](t, do(t, h, http.MethodGet, "/api/stats/revenue?days=7"))
	if len(rows) == 0 || len(rows) > 7 {
		t.Fatalf("rows = %d, want 1..7", len(rows))
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -7).Format("2006-01-02")
	for _, r := range rows {
		if r.Day < cutoff {
			t.Errorf("day %s older than cutoff %s", r.Day, cutoff)
		}
		if r.Orders <= 0 || r.Revenue <= 0 {
			t.Errorf("row %+v should have positive orders/revenue", r)
		}
	}
}

func TestRequestIDPropagation(t *testing.T) {
	h := newTestServer(t)
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

	rec := do(t, h, http.MethodGet, "/")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if body := decode[errorBody](t, rec); body.Error == "" {
		t.Error("error message missing")
	}
}
