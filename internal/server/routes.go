package server

// routes는 모든 엔드포인트를 한곳에서 등록한다. 파일 하나만 보면 API 전체와 권한이 보인다.
//
// Go 1.22+ ServeMux 패턴 문법: "[METHOD ][HOST]/path/{wildcard}"
//   - "GET /api/health"   → GET(과 HEAD)만 매칭, 다른 메서드는 자동으로 405
//   - "/api/users/{id}"   → r.PathValue("id")로 값 추출
//   - "/static/"          → 끝이 /이면 하위 경로 전체 매칭
//
// 권한: 기본은 공개. s.requireAuth(로그인 필요), s.requireRole("admin", ...)으로 감싼다.
func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/health", s.handleHealth)

	s.mux.HandleFunc("POST /api/auth/login", s.handleLogin)
	s.mux.HandleFunc("POST /api/auth/logout", s.handleLogout)
	s.mux.HandleFunc("GET /api/auth/me", s.requireAuth(s.handleMe))

	s.mux.HandleFunc("GET /api/products", s.requireAuth(s.handleListProducts))
	s.mux.HandleFunc("GET /api/products/{id}", s.requireAuth(s.handleGetProduct))

	s.mux.HandleFunc("GET /api/stats/revenue", s.requireAuth(s.handleRevenue))

	s.mux.HandleFunc("GET /api/users", s.requireRole("admin", s.handleListUsers))
}
