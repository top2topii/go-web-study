# 02. SQLite + goose + sqlc

2단계 코드: `internal/db/`, `internal/store/`(생성됨), `internal/seed/`, `cmd/seed/`, `sqlc.yaml`

## 전체 흐름

```
migrations/*.sql ──(goose)──▶ 실제 DB 스키마
       │
       └──(sqlc가 스키마로 읽음)
queries/*.sql ──(sqlc generate)──▶ internal/store/*.go  (타입 안전한 Go 함수)
                                         │
                         핸들러: s.q.ListProducts(ctx, params)
```

- **SQL이 원본**이고 Go 코드는 생성물이다. `internal/store/`는 직접 수정하지 않는다. 다음 `make generate` 때 덮어써진다.
- ORM과 달리 어떤 SQL이 실행되는지 그대로 보인다. 쿼리가 틀리면 **생성 시점에** 에러가 난다.

## 1. database/sql 기본

| 개념 | 설명 |
|---|---|
| `sql.Open("sqlite", dsn)` | 커넥션 **풀**을 만든다. 실제 연결은 하지 않으므로 `PingContext`로 확인한다 |
| `*sql.DB` | 여러 고루틴이 동시에 써도 안전하다. 앱 전체에서 하나만 공유한다 |
| 드라이버 등록 | `import _ "modernc.org/sqlite"`: 패키지의 `init()`이 "sqlite" 이름으로 등록된다 |
| `sql.ErrNoRows` | `:one` 쿼리 결과가 없을 때 나온다. `errors.Is(err, sql.ErrNoRows)`로 404 처리 |
| `*sql.Tx` | 트랜잭션. sqlc의 `store.New(tx)`는 DB와 Tx를 모두 받는다(`DBTX` 인터페이스) |

## 2. SQLite 운영 설정 — `internal/db/db.go`

| 설정 | 이유 |
|---|---|
| `foreign_keys(1)` | **SQLite는 기본적으로 FK를 검사하지 않는다.** 커넥션마다 켜야 한다 |
| `journal_mode(WAL)` | 읽기와 쓰기가 서로를 막지 않는다. 웹 서버라면 필수다 |
| `busy_timeout(5000)` | 잠금 충돌 시 바로 `SQLITE_BUSY`를 내지 않고 최대 5초 기다린다 |
| `_txlock=immediate` | 트랜잭션을 시작할 때 바로 쓰기 잠금을 잡아, 읽기→쓰기 승격 중에 생기는 교착을 방지한다 |
| `_time_format=sqlite` | `time.Time`을 SQLite 날짜 함수가 이해하는 문자열로 저장한다 |

CGO가 필요 없는 `modernc.org/sqlite`(C 코드를 Go로 변환한 드라이버)를 쓴다. 그래서 크로스 컴파일과 단일 바이너리 배포가 쉽다.

## 3. goose 마이그레이션

- `-- +goose Up` / `-- +goose Down` 두 구역으로 나뉜다. 파일 이름의 숫자 접두사가 순서가 된다.
- `//go:embed migrations/*.sql`로 SQL 파일을 **바이너리에 포함**한다. 배포할 때 파일을 따로 챙기지 않아도 된다.
- 적용 이력은 `goose_db_version` 테이블에 남는다. 서버가 시작할 때마다 `Migrate`를 호출해도 안전하다(멱등성, `TestMigrateIsIdempotent`).
- **이미 적용된 마이그레이션 파일은 수정하지 않는다.** 변경이 필요하면 `00002_xxx.sql`처럼 새 파일을 추가한다.

## 4. 스키마 설계 포인트 — `migrations/00001_init.sql`

- **금액은 정수(원)로 저장한다.** 부동소수점은 `0.1 + 0.2 != 0.3` 문제가 있다.
- `order_items.unit_price`: 주문 당시 가격을 복사해 둔다. 상품 가격이 나중에 바뀌어도 주문 금액은 그대로다.
- `CHECK (status IN (...))`: DB 차원에서 잘못된 값을 막는다. Go 코드에 버그가 있어도 데이터는 안전하다.
- `categories.parent_id`로 자기 참조 트리를 만든다(Tree 컴포넌트용).
- `COLLATE NOCASE`: 이메일 중복 검사에서 대소문자를 구분하지 않는다.

## 5. sqlc 설정 — `sqlc.yaml`

| 옵션 | 효과 |
|---|---|
| `emit_json_tags` | 생성된 struct에 `json:"snake_case"` 태그를 붙인다 |
| `emit_empty_slices` | 결과가 없을 때 nil 대신 `[]`를 반환한다. nil 슬라이스는 JSON에서 `null`이 되어 프론트가 깨진다 |
| `emit_pointers_for_null_types` | NULL 가능 컬럼을 `sql.NullString` 대신 `*string`으로 만든다. `sql.Null*`은 JSON에서 `{"String":"..","Valid":true}`가 된다 |
| `overrides: users.password_hash → json:"-"` | 해시가 실수로라도 응답에 나가지 않게 막는다 |

쿼리 주석 `-- name: ListProducts :many`에서 `:one`은 단일 행(없으면 `ErrNoRows`), `:many`는 슬라이스, `:exec`는 결과 없음을 뜻한다.

### 🐛 실제로 겪은 sqlc 버그

`sqlc.arg(since)`를 쓰는 쿼리 앞에 **한글 주석**이 있으면 `edited query syntax is invalid` 에러가 난다. sqlc가 `sqlc.arg()`를 `?`로 바꿀 때 바이트 오프셋과 문자 오프셋이 어긋나기 때문이다. 에러 위치가 주석 길이에 따라 움직이는 것으로 원인을 확인했다. → **쿼리 파일에는 ASCII만 쓴다.**

## 6. Go 문법 포인트

- **포인터로 NULL 표현**: `ParentID: &parent.ID`는 값이 있을 때, `nil`은 NULL일 때.
- **제네릭**: `type Page[T any] struct{ Items []T ... }`는 `Page[store.ListProductsRow]`처럼 여러 목록 응답에 재사용한다. 테스트의 `decode[T any]`도 제네릭 함수다.
- **이름 붙은 반환값 + defer로 트랜잭션 처리** (`seed.Run`):
  ```go
  func Run(...) (res Result, err error) {
      tx, err := db.BeginTx(ctx, nil)
      defer func() {
          if err != nil { tx.Rollback(); return }
          err = tx.Commit()   // defer 안에서 반환값을 바꿀 수 있다
      }()
  ```
- `for range n`: Go 1.22+ 정수 range.
- `math/rand/v2` + `rand.NewPCG(42, 2026)`: 고정 시드라서 매번 같은 데이터가 나온다.
- `sort.Search`: 정렬된 슬라이스에서 이진 탐색(가입일이 주문일 이전인 고객 고르기).
- `defer`는 LIFO다. `main.run`에서 `conn.Close()`가 `srv.Shutdown` 이후에 실행된다.

## 7. 시간 저장 형식 관찰

| 누가 넣었나 | 저장된 문자열 |
|---|---|
| DB 기본값 `CURRENT_TIMESTAMP` | `2026-09-29 05:54:57` |
| Go `time.Time` (`_time_format=sqlite`) | `2026-09-29 05:46:43.765955184+00:00` |

둘 다 UTC이고 앞부분 형식이 같아서 문자열 비교와 `date()` 집계는 정상 동작한다. 규칙: **Go에서 넣는 시간은 항상 `.UTC()`로 넣는다.** API 응답에서는 둘 다 `2026-09-29T05:57:16Z`(RFC3339)로 나간다.

## 8. 테스트 — 임시 파일 DB

`t.TempDir()`에 실제 SQLite 파일을 만들고, 마이그레이션과 소량 시드를 넣는다. `:memory:` DB는 **커넥션마다 별개의 DB**라서 커넥션 풀과 함께 쓰면 테이블이 안 보이는 함정이 있다. 임시 파일이 더 단순하고 운영 환경과도 같다.

## API (2단계)

| 엔드포인트 | 설명 |
|---|---|
| `GET /api/health` | DB ping 포함. 실패하면 503 `degraded` |
| `GET /api/products?page=&page_size=` | 페이지네이션(최대 100), 카테고리명 JOIN |
| `GET /api/products/{id}` | 없으면 404, 숫자가 아니면 400 |
| `GET /api/stats/revenue?days=30` | 일별 주문 수와 매출(취소·환불 제외), 1~365일 |

## 명령어

```sh
make seed       # data/shopadmin.db 재생성 (카테고리 14, 상품 120, 고객 200, 주문 ~1500)
make generate   # 쿼리 수정 후 Go 코드 재생성
make run
sqlite3 data/shopadmin.db   # .tables / .schema orders / SELECT ...
```
