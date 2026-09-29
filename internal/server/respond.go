package server

import (
	"encoding/json"
	"net/http"
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
