package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"shopadmin/internal/auth"
	"shopadmin/internal/db"
	"shopadmin/internal/seed"
	"shopadmin/internal/store"
)

// TestMain은 이 패키지의 모든 테스트 전에 한 번 실행된다.
// bcrypt 비용을 최소로 낮춰 테스트 속도를 높인다(기본값이면 해시 1회에 ~50ms).
func TestMain(m *testing.M) {
	auth.Cost = bcrypt.MinCost
	os.Exit(m.Run())
}

// 테스트에서는 로그를 버린다. 디버깅할 땐 io.Discard를 os.Stderr로 바꾸면 된다.
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

const (
	testProducts = 15
	testPassword = "test-password"
	adminEmail   = "admin@example.com"
	viewerEmail  = "viewer@example.com"
)

type testEnv struct {
	h  http.Handler
	db *sql.DB
}

// newTestServer는 임시 SQLite 파일에 마이그레이션 + 소량 시드를 넣고 핸들러를 만든다.
func newTestServer(t *testing.T) testEnv {
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
		Customers: 20, Products: testProducts, Days: 10, Now: time.Now().UTC(), Password: testPassword,
	}); err != nil {
		t.Fatal(err)
	}
	return testEnv{h: New(testLogger(), conn, Config{Version: "test"}), db: conn}
}

// send는 요청을 보내고 응답 레코더를 돌려준다. body가 ""이면 본문 없음, c가 nil이면 쿠키 없음.
func (e testEnv) send(t *testing.T, method, path, body string, c *http.Cookie, hdr ...string) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	if c != nil {
		req.AddCookie(c)
	}
	// hdr는 가변 인자(...string): "Key", "Value", "Key2", "Value2" 쌍
	for i := 0; i+1 < len(hdr); i += 2 {
		req.Header.Set(hdr[i], hdr[i+1])
	}
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, req)
	return rec
}

func (e testEnv) get(t *testing.T, path string, c *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	return e.send(t, http.MethodGet, path, "", c)
}

// login은 로그인에 성공해야 하며, 세션 쿠키를 반환한다.
func (e testEnv) login(t *testing.T, email string) *http.Cookie {
	t.Helper()
	rec := e.send(t, http.MethodPost, "/api/auth/login",
		`{"email":"`+email+`","password":"`+testPassword+`"}`, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("login %s: status %d body=%s", email, rec.Code, rec.Body)
	}
	c := findCookie(rec, sessionCookie)
	if c == nil {
		t.Fatal("login: no session cookie")
	}
	return c
}

func findCookie(rec *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
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

// ---- 기본 ----

func TestHealth(t *testing.T) {
	e := newTestServer(t)
	rec := e.get(t, "/api/health", nil)

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
	e := newTestServer(t)
	viewer := e.login(t, viewerEmail)
	admin := e.login(t, adminEmail)

	tests := []struct {
		name       string
		method     string
		path       string
		cookie     *http.Cookie
		wantStatus int
	}{
		{"health GET", http.MethodGet, "/api/health", nil, http.StatusOK},
		{"health HEAD", http.MethodHead, "/api/health", nil, http.StatusOK},
		{"health POST is 405", http.MethodPost, "/api/health", nil, http.StatusMethodNotAllowed},
		{"unknown path is 404", http.MethodGet, "/api/nope", nil, http.StatusNotFound},

		{"products anon is 401", http.MethodGet, "/api/products", nil, http.StatusUnauthorized},
		{"products viewer", http.MethodGet, "/api/products", viewer, http.StatusOK},
		{"products bad page", http.MethodGet, "/api/products?page=0", viewer, http.StatusBadRequest},
		{"products page_size too big", http.MethodGet, "/api/products?page_size=1000", viewer, http.StatusBadRequest},
		{"product found", http.MethodGet, "/api/products/1", viewer, http.StatusOK},
		{"product not found", http.MethodGet, "/api/products/99999", viewer, http.StatusNotFound},
		{"product bad id", http.MethodGet, "/api/products/abc", viewer, http.StatusBadRequest},
		{"revenue anon is 401", http.MethodGet, "/api/stats/revenue", nil, http.StatusUnauthorized},
		{"revenue bad days", http.MethodGet, "/api/stats/revenue?days=0", viewer, http.StatusBadRequest},

		{"me anon is 401", http.MethodGet, "/api/auth/me", nil, http.StatusUnauthorized},
		{"users anon is 401", http.MethodGet, "/api/users", nil, http.StatusUnauthorized},
		{"users viewer is 403", http.MethodGet, "/api/users", viewer, http.StatusForbidden},
		{"users admin", http.MethodGet, "/api/users", admin, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rec := e.send(t, tt.method, tt.path, "", tt.cookie); rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body=%s)", rec.Code, tt.wantStatus, rec.Body)
			}
		})
	}
}

// ---- 인증 ----

func TestLogin(t *testing.T) {
	e := newTestServer(t)

	t.Run("success sets secure-by-default cookie", func(t *testing.T) {
		rec := e.send(t, http.MethodPost, "/api/auth/login",
			`{"email":"  ADMIN@example.com ","password":"`+testPassword+`"}`, nil) // 공백·대소문자 무시
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body)
		}
		c := findCookie(rec, sessionCookie)
		if c == nil || !c.HttpOnly || c.SameSite != http.SameSiteLaxMode || c.MaxAge <= 0 {
			t.Errorf("cookie = %+v", c)
		}
		if strings.Contains(rec.Body.String(), "password") {
			t.Errorf("response leaks password field: %s", rec.Body)
		}
		if u := decode[store.User](t, rec); u.Role != "admin" {
			t.Errorf("role = %q", u.Role)
		}
	})

	// 이메일이 없을 때와 비밀번호가 틀릴 때 응답이 완전히 같아야 한다.
	var bodies []string
	for _, body := range []string{
		`{"email":"admin@example.com","password":"wrong-password"}`,
		`{"email":"nobody@example.com","password":"wrong-password"}`,
	} {
		rec := e.send(t, http.MethodPost, "/api/auth/login", body, nil)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s → %d, want 401", body, rec.Code)
		}
		if findCookie(rec, sessionCookie) != nil {
			t.Error("failed login must not set cookie")
		}
		bodies = append(bodies, rec.Body.String())
	}
	if bodies[0] != bodies[1] {
		t.Errorf("responses differ → user enumeration possible:\n%s\n%s", bodies[0], bodies[1])
	}

	for _, bad := range []string{`not json`, `{"email":"a","pasword":"typo"}`, `{"email":"a"}{"email":"b"}`} {
		if rec := e.send(t, http.MethodPost, "/api/auth/login", bad, nil); rec.Code != http.StatusBadRequest {
			t.Errorf("%q → %d, want 400", bad, rec.Code)
		}
	}
}

func TestMeAndLogout(t *testing.T) {
	e := newTestServer(t)
	c := e.login(t, viewerEmail)

	if u := decode[store.User](t, e.get(t, "/api/auth/me", c)); u.Email != viewerEmail {
		t.Fatalf("me = %+v", u)
	}

	rec := e.send(t, http.MethodPost, "/api/auth/logout", "", c)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d", rec.Code)
	}
	if cleared := findCookie(rec, sessionCookie); cleared == nil || cleared.MaxAge >= 0 {
		t.Errorf("logout should clear cookie, got %+v", cleared)
	}
	// 브라우저가 쿠키를 지우지 않고 재사용해도(탈취된 쿠키 등) 서버에서 이미 무효다.
	if rec := e.get(t, "/api/auth/me", c); rec.Code != http.StatusUnauthorized {
		t.Errorf("me after logout = %d, want 401", rec.Code)
	}
	// 로그아웃은 멱등: 다시 해도 204
	if rec := e.send(t, http.MethodPost, "/api/auth/logout", "", c); rec.Code != http.StatusNoContent {
		t.Errorf("second logout = %d", rec.Code)
	}
}

func TestExpiredSession(t *testing.T) {
	e := newTestServer(t)
	ctx := context.Background()
	q := store.New(e.db)

	u, err := q.GetUserByEmail(ctx, adminEmail)
	if err != nil {
		t.Fatal(err)
	}
	token := auth.NewToken()
	if err := q.CreateSession(ctx, store.CreateSessionParams{
		TokenHash: auth.HashToken(token), UserID: u.ID, ExpiresAt: time.Now().UTC().Add(-time.Minute),
	}); err != nil {
		t.Fatal(err)
	}

	rec := e.get(t, "/api/auth/me", &http.Cookie{Name: sessionCookie, Value: token})
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expired session status = %d, want 401", rec.Code)
	}
	if c := findCookie(rec, sessionCookie); c == nil || c.MaxAge >= 0 {
		t.Error("stale cookie should be cleared")
	}

	// 다음 로그인 때 만료 세션이 청소된다.
	e.login(t, adminEmail)
	var n int
	if err := e.db.QueryRow("SELECT count(*) FROM sessions WHERE expires_at <= ?", time.Now().UTC()).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("expired sessions left = %d", n)
	}
}

func TestCrossOriginProtection(t *testing.T) {
	e := newTestServer(t)
	body := `{"email":"admin@example.com","password":"` + testPassword + `"}`

	// 다른 사이트의 페이지가 fetch/form으로 POST하면 브라우저가 Sec-Fetch-Site: cross-site를 붙인다.
	if rec := e.send(t, http.MethodPost, "/api/auth/login", body, nil, "Sec-Fetch-Site", "cross-site"); rec.Code != http.StatusForbidden {
		t.Errorf("cross-site POST = %d, want 403", rec.Code)
	}
	if rec := e.send(t, http.MethodPost, "/api/auth/login", body, nil, "Sec-Fetch-Site", "same-origin"); rec.Code != http.StatusOK {
		t.Errorf("same-origin POST = %d, want 200", rec.Code)
	}
	// GET은 "안전한 메서드"라 교차 출처여도 통과(대신 GET은 상태를 바꾸면 안 된다).
	if rec := e.send(t, http.MethodGet, "/api/health", "", nil, "Sec-Fetch-Site", "cross-site"); rec.Code != http.StatusOK {
		t.Errorf("cross-site GET = %d, want 200", rec.Code)
	}
}

// ---- 데이터 ----

func TestListProductsPagination(t *testing.T) {
	e := newTestServer(t)
	c := e.login(t, viewerEmail)

	first := decode[Page[store.ListProductsRow]](t, e.get(t, "/api/products?page=1&page_size=5", c))
	if len(first.Items) != 5 || first.Total != testProducts {
		t.Fatalf("page1: len=%d total=%d", len(first.Items), first.Total)
	}
	if first.Items[0].CategoryName == nil {
		t.Error("category_name should be joined")
	}

	second := decode[Page[store.ListProductsRow]](t, e.get(t, "/api/products?page=2&page_size=5", c))
	if second.Items[0].ID == first.Items[0].ID {
		t.Error("page 2 returned same items as page 1")
	}

	// 범위를 벗어난 페이지는 null이 아니라 빈 배열이어야 한다(emit_empty_slices).
	raw := decode[map[string]json.RawMessage](t, e.get(t, "/api/products?page=100", c))
	if string(raw["items"]) != "[]" {
		t.Errorf("items = %s, want []", raw["items"])
	}
}

func TestRevenue(t *testing.T) {
	e := newTestServer(t)
	c := e.login(t, viewerEmail)

	rows := decode[[]store.RevenueByDayRow](t, e.get(t, "/api/stats/revenue?days=7", c))
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

// ---- 미들웨어 ----

func TestRequestIDPropagation(t *testing.T) {
	e := newTestServer(t)
	rec := e.send(t, http.MethodGet, "/api/health", "", nil, "X-Request-ID", "abc123")
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
	if body := decode[errorBody](t, rec); body.Error == "" {
		t.Error("error message missing")
	}
}
