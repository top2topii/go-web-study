# 03. 인증: 세션 쿠키 + bcrypt + 역할

3단계 코드: `internal/auth/`, `internal/server/handlers_auth.go`, `migrations/00002_sessions.sql`

## 전체 흐름

```
POST /api/auth/login {email, password}
  → users에서 조회 → bcrypt 비교
  → token = crypto/rand 32바이트 (쿠키로 전송)
  → sessions에 sha256(token) 저장 (DB에는 원문 없음)
  → Set-Cookie: sid=token; HttpOnly; Secure; SameSite=Lax

이후 모든 요청
  → loadSession 미들웨어: 쿠키 → sha256 → sessions JOIN users (만료 확인)
  → context에 *store.User 저장
  → requireAuth / requireRole이 context를 보고 401 / 403 결정
```

## 1. 세션 방식 vs JWT

| | 서버 세션(선택) | JWT |
|---|---|---|
| 로그아웃·강제 만료 | DB 행 삭제로 **즉시** | 만료 전까지 유효(블랙리스트가 필요) |
| 저장소 | DB 조회 1회/요청 | 서명 검증만 |
| 적합 | 단일 서버 admin 앱 | 여러 서비스 간 인증 위임 |

SQLite를 쓰는 단일 서버 admin에는 서버 세션이 더 단순하고 안전하다. `TestMeAndLogout`에서 로그아웃 후 **같은 쿠키를 재사용해도 401**이 나오는 것을 확인한다.

## 2. 비밀번호 — `internal/auth/auth.go`

- **bcrypt**: 솔트가 해시 안에 포함되어 같은 비밀번호도 해시가 매번 다르다. `Cost`를 1 올릴 때마다 계산 시간이 2배가 된다(기본 10, 약 50ms).
- **72바이트 제한**: bcrypt는 72바이트까지만 쓴다. 그래서 `ValidatePassword`가 초과분을 거부한다.
- **사용자 열거(user enumeration) 방지**, 두 가지:
  1. 응답 문구를 같게 한다: 이메일이 없을 때와 비밀번호가 틀릴 때 모두 `invalid email or password`.
  2. **시간도 같게 한다**: 사용자가 없어도 더미 해시와 bcrypt 비교를 수행한다.

  실측(Chrome → 서버 로그):
  ```
  존재하는 계정 + 틀린 비번   : 53ms, 42ms
  없는 계정                    : 55ms, 42ms
  ```
  이 처리가 없으면 없는 계정은 약 0.03ms 만에 응답해서, 가입 여부가 시간으로 드러난다.

## 3. 세션 토큰

- `crypto/rand`를 쓴다. `math/rand`는 예측 가능하므로 보안 용도로 쓰면 안 된다.
- DB에는 `sha256(token)`만 저장한다. DB가 유출돼도 해시로는 쿠키를 만들 수 없다. 토큰 자체가 256비트 난수라서 bcrypt 같은 느린 해시는 필요 없다.
- **세션 고정(session fixation) 방지**: 로그인할 때 기존 세션을 지우고 항상 새 토큰을 발급한다.
- 만료된 세션은 로그인할 때 함께 청소한다(`DeleteExpiredSessions :execrows`는 삭제된 행 수를 반환한다).

## 4. 쿠키 속성

| 속성 | 막는 것 |
|---|---|
| `HttpOnly` | JS의 `document.cookie`로 읽을 수 없다. XSS가 생겨도 세션 탈취가 안 된다. **실측: Chrome에서 `document.cookie` = `""`** |
| `Secure` | HTTPS에서만 전송한다. 브라우저는 `http://localhost`를 예외로 허용한다. **실측: Chrome에서 localhost HTTP로 로그인 동작 확인** |
| `SameSite=Lax` | 다른 사이트에서 시작된 POST에는 쿠키를 붙이지 않는다 |
| `MaxAge` | 7일. 로그아웃이나 무효 쿠키일 때는 `MaxAge=-1`로 삭제를 지시한다 |

## 5. CSRF — Go 1.25 `http.CrossOriginProtection`

```go
h = http.NewCrossOriginProtection().Handler(h)
```

- 브라우저가 자동으로 붙이는 `Sec-Fetch-Site` 헤더(없으면 `Origin`과 `Host` 비교)를 보고, **다른 출처에서 온 POST/PUT/PATCH/DELETE를 403으로 거부**한다.
- GET, HEAD, OPTIONS는 통과시킨다. 그러니 **GET 핸들러는 절대 상태를 바꾸면 안 된다.**
- 예전에는 CSRF 토큰을 발급하고 폼에 심어야 했지만, 이제 표준 라이브러리만으로 충분하다(`TestCrossOriginProtection`).

## 6. 미들웨어 설계: 로드와 거부를 분리

| 단계 | 역할 | 실패 시 |
|---|---|---|
| `loadSession` (전역) | 쿠키가 있으면 사용자를 context에 넣는다 | 막지 않는다(익명으로 통과), 낡은 쿠키는 삭제 |
| `requireAuth` (라우트별) | context에 사용자가 있는지 확인 | 401 |
| `requireRole("admin", …)` | 역할 확인 | 403 |

권한이 `routes.go` 한 파일에 모두 보인다:
```go
s.mux.HandleFunc("GET /api/users", s.requireRole("admin", s.handleListUsers))
```

**401 vs 403**: 401은 "누구인지 모름(로그인 필요)", 403은 "누구인지는 알지만 권한 없음"이다.

## 7. 요청 본문 파싱 — `readJSON`

- `http.MaxBytesReader`: 1MB를 넘는 본문을 거부한다(메모리 고갈 공격 방지).
- `DisallowUnknownFields`: `"pasword"` 같은 오타 필드를 조용히 무시하지 않고 400으로 돌려준다.
- `dec.More()`: JSON 객체가 두 개 이어 붙은 본문을 거부한다.

## 8. Go 문법 포인트

- **comma-ok 관용구**: `u, ok := ctx.Value(userKey).(*store.User)`. 타입이 맞지 않아도 panic 대신 `ok=false`가 된다.
- **iota**: `const ( requestIDKey ctxKey = iota; userKey )`는 0, 1로 자동 증가한다.
- **가변 인자**: `send(t, ..., hdr ...string)`. 호출할 때 `"Key", "Value"` 쌍을 넘긴다.
- **패키지 변수로 테스트 설정 바꾸기**: `TestMain`에서 `auth.Cost = bcrypt.MinCost`. 서버 패키지 테스트 전체가 4초대로 끝난다.
- **sqlc.embed(users)**: JOIN 결과를 `GetSessionUserRow{User: store.User{...}}`처럼 struct째로 받는다.
- **HandlerFunc를 감싸는 함수**: `requireAuth(h http.HandlerFunc) http.HandlerFunc`. 전역 미들웨어(`http.Handler`)와 달리 라우트마다 선택적으로 적용한다.

## 9. 실측 중 발견: 실제 브라우저 프로필

BrowserSkill은 **실제 Chrome 프로필**을 쓴다. 한 bsk 세션에서 로그인한 쿠키가 다음 bsk 세션에도 남아 있어서, "익명 요청"이 200으로 나와 잠시 헷갈렸다. 브라우저로 인증을 검증할 때는 **먼저 로그아웃하고 시작**한다.

## 10. 아직 없는 것 (이후 과제)

- 로그인 시도 횟수 제한(rate limit): 무차별 대입 방어
- 비밀번호 변경 시 해당 사용자의 다른 세션 모두 삭제
- 슬라이딩 만료(활동이 있으면 세션 연장)

## API (3단계)

| 엔드포인트 | 권한 |
|---|---|
| `POST /api/auth/login` | 공개 |
| `POST /api/auth/logout` | 공개(멱등) |
| `GET /api/auth/me` | 로그인 |
| `GET /api/products`, `/api/products/{id}`, `/api/stats/revenue` | 로그인 |
| `GET /api/users` | admin |

## 명령어

```sh
SEED_PASSWORD=원하는비번 make seed   # admin@example.com / viewer@example.com
make seed                            # 비밀번호 무작위 생성, 콘솔에 한 번 출력
./bin/server -cookie-secure=false    # localhost가 아닌 주소로 HTTP 접속할 때만
```
