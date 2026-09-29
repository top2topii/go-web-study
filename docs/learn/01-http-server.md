# 01. 표준 라이브러리 HTTP 서버

1단계 코드: `cmd/server/`, `internal/server/`

## 1. 디렉토리 규칙

| 경로 | 의미 |
|---|---|
| `cmd/<이름>/` | 실행 파일 하나당 디렉토리 하나. `package main` |
| `internal/` | **컴파일러가 강제하는** 비공개 영역. 이 모듈 밖에서는 import 불가 |
| `go.mod` | 모듈 이름(`shopadmin`)과 Go 버전. import 경로는 `shopadmin/internal/server`처럼 모듈 이름으로 시작 |

패키지 이름은 디렉토리 이름과 같게, 짧은 소문자로 짓는다. `http`처럼 표준 패키지와 겹치는 이름은 피한다(그래서 `server`).

## 2. `main`은 얇게, `run`은 두껍게 — `cmd/server/main.go`

```go
func main() {
    if err := run(ctx, os.Args[1:], os.Getenv, os.Stdout); err != nil { ... os.Exit(1) }
}
```

- `os.Exit`, `os.Args`, `os.Getenv` 같은 **전역 상태를 run의 인자로 주입**한다. 그러면 테스트에서 가짜 값을 넣을 수 있다(`main_test.go`의 `env(map)` 참고).
- Go는 예외가 없다. 실패는 마지막 반환값 `error`로 알린다. `if err != nil { return err }`가 기본 패턴이다.
- 에러에 맥락을 붙일 때는 `fmt.Errorf("graceful shutdown: %w", err)`를 쓴다. `%w`로 감싸면 나중에 `errors.Is`나 `errors.As`로 원래 에러를 꺼낼 수 있다.

## 3. Go 1.22+ ServeMux 패턴 — `internal/server/routes.go`

```go
mux.HandleFunc("GET /api/health", s.handleHealth)
mux.HandleFunc("GET /api/users/{id}", h) // r.PathValue("id")
```

- 메서드를 지정하면 다른 메서드로 온 요청은 **자동으로 405 + `Allow` 헤더**를 받는다. `GET`을 지정하면 `HEAD`도 함께 허용된다.
- 여러 패턴이 겹치면 **더 구체적인 패턴이 이긴다**. 등록 순서와는 무관하다.
- 예전에는 이 기능 때문에 gorilla/mux나 chi를 썼지만, 이제는 표준만으로 충분하다.

## 4. 핸들러와 의존성

```go
type Server struct { logger *slog.Logger; ... }
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request)
```

- 핸들러의 시그니처는 항상 `(http.ResponseWriter, *http.Request)`이다.
- 핸들러를 `*Server`의 **메서드**로 만들면 전역 변수 없이 로거나 DB에 접근할 수 있다. 메서드 값 `s.handleHealth`는 그대로 `HandlerFunc`로 쓸 수 있다.
- `ResponseWriter`의 순서 규칙: **Header 설정 → WriteHeader(status) → Write(body)**. WriteHeader 이후에 설정한 헤더는 무시된다(`respond.go`).

## 5. 미들웨어 — `internal/server/middleware.go`

```go
func(next http.Handler) http.Handler
```

- 다음 핸들러를 받아 앞뒤로 일을 덧붙인 새 핸들러를 돌려준다. 감싸는 순서가 곧 실행 순서다(`server.go` 주석 참고).
- **임베딩**: `statusRecorder`는 `http.ResponseWriter`를 필드 이름 없이 품고, `WriteHeader`와 `Write`만 재정의해서 상태 코드를 가로챈다. 나머지 메서드는 자동으로 위임된다.
- **context 값**: 요청 범위 데이터(요청 ID, 나중엔 로그인 사용자)를 담는다. 키는 비공개 타입(`ctxKey`)으로 만들어 다른 패키지와의 충돌을 막는다.
- **defer + recover**: panic은 스택을 거슬러 올라가며 defer를 실행한다. `recover()`는 defer 안에서만 panic을 멈출 수 있다.

## 6. 로깅 — `log/slog`

표준 구조화 로거다. `key=value` 쌍으로 기록한다. `-log-json` 플래그를 주면 JSON으로 출력한다.

```
level=INFO msg=request id=60668985fe35a0bb method=GET path=/api/health status=200 bytes=50 dur=170µs
```

## 7. Graceful shutdown

1. `signal.NotifyContext`: Ctrl+C나 SIGTERM을 받으면 ctx가 취소된다.
2. `ListenAndServe`는 블로킹이므로 **고루틴**에서 실행하고, 에러는 **채널**로 받는다.
3. `select`로 "서버 에러"와 "종료 신호" 중 먼저 오는 쪽을 기다린다.
4. `srv.Shutdown(ctx)`는 새 연결을 받지 않고, 진행 중인 요청이 끝날 때까지 기다린다(최대 10초).

**실측 관찰**: Chrome으로 접속한 직후 종료하면 약 4초가 걸렸다. Chrome이 요청 없이 미리 열어두는 연결(preconnect) 때문이다. `Shutdown`은 요청이 한 번도 오지 않은 새 연결을 **5초가 지나야** 유휴 상태로 보고 닫는다. 연결이 열린 지 6초 뒤에 종료하면 0.004초 만에 끝났다.

## 8. 서버 타임아웃

`http.Server{}`의 기본값에는 **타임아웃이 없다**. `ReadHeaderTimeout` 등을 설정하지 않으면 헤더를 아주 느리게 보내는 클라이언트가 연결을 계속 붙잡을 수 있다(Slowloris 공격). 그래서 `http.ListenAndServe(addr, h)` 단축형은 운영에서 쓰지 않는다.

## 9. 테스트 — `*_test.go`

- `httptest.NewRecorder()`와 `httptest.NewRequest()`로 **네트워크 없이** 핸들러를 호출한다.
- **테이블 주도 테스트**: 케이스를 struct 슬라이스로 나열하고 `t.Run(name, ...)`으로 하위 테스트를 돌린다. 실패하면 어느 케이스인지 이름이 나온다.
- `t.Fatalf`는 즉시 중단하고, `t.Errorf`는 기록만 하고 계속 진행한다.
- 같은 패키지(`package server`)의 테스트는 비공개 함수(`recoverPanic`)도 테스트할 수 있다.

## 명령어

```sh
make run     # go run ./cmd/server  (기본 127.0.0.1:8080)
make test    # go test ./...
make check   # gofmt 검사 + vet + race detector 테스트
make build   # bin/server, 버전은 git describe로 주입(-ldflags -X)
./bin/server -addr :9000 -log-level debug -log-json
```
