package server

import (
	"context"
	"net/http"
	"time"
)

// 구조체 태그 `json:"..."`로 JSON 필드 이름을 정한다.
type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
	DB      string `json:"db"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := healthResponse{
		Status:  "ok",
		Version: s.version,
		Uptime:  time.Since(s.started).Round(time.Second).String(),
		DB:      "ok",
	}
	status := http.StatusOK

	// 요청 context에서 파생: 클라이언트가 끊으면 Ping도 취소된다.
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.db.PingContext(ctx); err != nil {
		s.logger.Error("health: db ping failed", "err", err)
		resp.Status, resp.DB = "degraded", "unreachable"
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, resp)
}
