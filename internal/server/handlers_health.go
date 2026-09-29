package server

import (
	"net/http"
	"time"
)

// 구조체 태그 `json:"..."`로 JSON 필드 이름을 정한다.
type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:  "ok",
		Version: s.version,
		Uptime:  time.Since(s.started).Round(time.Second).String(),
	})
}
