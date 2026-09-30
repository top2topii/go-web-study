package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"shopadmin/internal/auth"
	"shopadmin/internal/store"
)

const sessionCookie = "sid"

// ---- context에 로그인 사용자 보관 ----

func withUser(ctx context.Context, u *store.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

// userFrom은 로그인 사용자를 꺼낸다. 두 번째 반환값은 "있음" 여부(comma-ok 관용구).
func userFrom(ctx context.Context) (*store.User, bool) {
	u, ok := ctx.Value(userKey).(*store.User)
	return u, ok
}

// ---- 미들웨어 ----

// loadSession은 쿠키가 있으면 세션을 조회해 사용자를 context에 넣는다.
// 쿠키가 없거나 무효여도 요청을 막지 않는다. 막는 건 requireAuth의 몫이다(관심사 분리).
func (s *Server) loadSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(sessionCookie)
		if err != nil || c.Value == "" {
			next.ServeHTTP(w, r)
			return
		}
		row, err := s.q.GetSessionUser(r.Context(), store.GetSessionUserParams{
			TokenHash: auth.HashToken(c.Value),
			Now:       time.Now().UTC(),
		})
		if errors.Is(err, sql.ErrNoRows) {
			// 만료됐거나 로그아웃된 세션: 브라우저의 낡은 쿠키도 지운다.
			s.clearSessionCookie(w)
			next.ServeHTTP(w, r)
			return
		}
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(withUser(r.Context(), &row.User)))
	})
}

// requireAuth는 로그인하지 않은 요청을 401로 거부한다.
// 라우트 단위로 감싸기 위해 HandlerFunc → HandlerFunc 형태로 만든다.
func (s *Server) requireAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := userFrom(r.Context()); !ok {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		h(w, r)
	}
}

// requireRole은 로그인 + 역할을 검사한다. 로그인 안 함 → 401, 권한 부족 → 403.
func (s *Server) requireRole(role string, h http.HandlerFunc) http.HandlerFunc {
	return s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		u, _ := userFrom(r.Context())
		if u.Role != role {
			writeError(w, http.StatusForbidden, "insufficient permissions")
			return
		}
		h(w, r)
	})
}

// ---- 쿠키 ----

func (s *Server) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   int(s.cfg.SessionTTL.Seconds()),
		HttpOnly: true,                 // JS(document.cookie)에서 읽을 수 없음 → XSS로 탈취 불가
		Secure:   s.cfg.CookieSecure,   // HTTPS(와 localhost)에서만 전송
		SameSite: http.SameSiteLaxMode, // 다른 사이트에서 시작된 POST에는 쿠키를 붙이지 않음
	})
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: s.cfg.CookieSecure, SameSite: http.SameSiteLaxMode,
	})
}

// ---- 핸들러 ----

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// POST /api/auth/login  {"email": "...", "password": "..."}
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx := r.Context()

	u, err := s.q.GetUserByEmail(ctx, strings.TrimSpace(req.Email))
	var hash string // 사용자가 없으면 빈 문자열 → CheckPassword가 더미 비교로 시간을 맞춘다
	switch {
	case err == nil:
		hash = u.PasswordHash
	case !errors.Is(err, sql.ErrNoRows):
		s.serverError(w, r, err)
		return
	}
	// "이메일이 없음"과 "비밀번호 틀림"을 구분하지 않는다(계정 존재 여부 노출 방지).
	if !auth.CheckPassword(hash, req.Password) {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	// 세션 고정(session fixation) 방지: 기존 세션은 지우고 항상 새 토큰을 발급한다.
	if c, err := r.Cookie(sessionCookie); err == nil {
		_ = s.q.DeleteSession(ctx, auth.HashToken(c.Value))
	}

	now := time.Now().UTC()
	token := auth.NewToken()
	if err := s.q.CreateSession(ctx, store.CreateSessionParams{
		TokenHash: auth.HashToken(token),
		UserID:    u.ID,
		ExpiresAt: now.Add(s.cfg.SessionTTL),
	}); err != nil {
		s.serverError(w, r, err)
		return
	}
	// 만료 세션 청소는 로그인 김에 한다. 실패해도 로그인은 성공이므로 로그만 남긴다.
	if n, err := s.q.DeleteExpiredSessions(ctx, now); err != nil {
		s.logger.Warn("cleanup expired sessions", "err", err)
	} else if n > 0 {
		s.logger.Info("expired sessions removed", "count", n)
	}

	s.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, u) // User.PasswordHash는 json:"-"라 응답에 포함되지 않는다
}

// POST /api/auth/logout
// 로그인 상태가 아니어도 204. 로그아웃은 멱등해야 한다.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		if err := s.q.DeleteSession(r.Context(), auth.HashToken(c.Value)); err != nil {
			s.serverError(w, r, err)
			return
		}
	}
	s.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/auth/me
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	u, _ := userFrom(r.Context())
	writeJSON(w, http.StatusOK, u)
}

// GET /api/users (admin 전용)
func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.q.ListUsers(r.Context())
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
}
