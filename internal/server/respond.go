package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// writeJSON은 v를 JSON으로 직렬화해 응답한다.
// any는 interface{}의 별칭으로, 어떤 타입이든 받는다.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// 헤더는 WriteHeader 이전에 설정해야 한다. 이후 설정은 무시된다.
	w.WriteHeader(status)
	// 이미 상태 코드를 보낸 뒤라 인코딩 에러가 나도 응답을 바꿀 수 없다.
	_ = json.NewEncoder(w).Encode(v)
}

type errorBody struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorBody{Error: msg})
}

// serverError는 내부 에러를 로그에 남기고 클라이언트에는 일반 메시지만 준다.
// DB 에러 문자열 같은 내부 정보를 응답에 노출하지 않기 위해서다.
func (s *Server) serverError(w http.ResponseWriter, r *http.Request, err error) {
	s.logger.Error("internal error",
		"id", RequestIDFrom(r.Context()),
		"method", r.Method,
		"path", r.URL.Path,
		"err", err,
	)
	writeError(w, http.StatusInternalServerError, "internal server error")
}

// Page는 목록 응답의 공통 모양이다.
// [T any]는 제네릭 타입 파라미터: Page[store.Product], Page[store.Customer] 등으로 재사용.
type Page[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int64 `json:"page"`
	PageSize int64 `json:"page_size"`
}

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// parsePage는 ?page=&page_size= 를 읽어 검증한다. 없으면 기본값.
func parsePage(q url.Values) (page, size int64, err error) {
	page, size = 1, defaultPageSize
	if v := q.Get("page"); v != "" {
		if page, err = strconv.ParseInt(v, 10, 64); err != nil || page < 1 {
			return 0, 0, fmt.Errorf("page must be a positive integer")
		}
	}
	if v := q.Get("page_size"); v != "" {
		if size, err = strconv.ParseInt(v, 10, 64); err != nil || size < 1 || size > maxPageSize {
			return 0, 0, fmt.Errorf("page_size must be between 1 and %d", maxPageSize)
		}
	}
	return page, size, nil
}

// pathID는 경로 와일드카드 {name}을 양의 정수로 읽는다.
func pathID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return id, nil
}
