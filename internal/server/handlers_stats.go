package server

import (
	"net/http"
	"strconv"
	"time"
)

// GET /api/stats/revenue?days=30
// 오늘 포함 최근 N일의 일별 주문 수와 매출(취소·환불 제외). UTC 기준.
func (s *Server) handleRevenue(w http.ResponseWriter, r *http.Request) {
	days := 30
	if v := r.URL.Query().Get("days"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 365 {
			writeError(w, http.StatusBadRequest, "days must be between 1 and 365")
			return
		}
		days = n
	}

	// 오늘 0시(UTC)에서 days-1일 전 0시부터
	since := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -(days - 1))

	rows, err := s.q.RevenueByDay(r.Context(), since)
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}
