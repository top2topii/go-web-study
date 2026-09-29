# Admin Design System — 프로젝트 스펙

Go 학습 + React/Tailwind admin 디자인 시스템 구축. 구현 가능한 거의 모든 컴포넌트를 쇼케이스.
결과 확인은 BrowserSkill(`bsk`) CLI로 Claude가 직접 브라우저에서 검증.

## 결정 사항 (인터뷰 2026-09-29)

| 항목 | 결정 |
|---|---|
| Go 버전 | 1.22.2 → 최신 안정 버전으로 업그레이드 |
| HTTP | 표준 `net/http` ServeMux (메서드·경로 패턴), 프레임워크 없음 |
| DB | SQLite, `modernc.org/sqlite`(CGO 불필요) + `database/sql` + sqlc, 마이그레이션 goose |
| 인증 | 세션 쿠키 + bcrypt, 역할 admin / viewer |
| 도메인 | 이커머스: 사용자·상품·카테고리·주문 CRUD |
| 프론트 | Vite + React + TypeScript + Tailwind, pnpm |
| 컴포넌트 | shadcn/ui 스타일 (Radix 헤드리스 + Tailwind, 코드 소유) |
| 테마 | CSS 변수 토큰, 라이트/다크, 브랜드 컬러 교체 가능 |
| i18n | 한국어 / 영어 |
| 배포 구조 | 개발: Vite dev server → `/api` Go로 프록시. 배포: Go `embed`로 단일 바이너리 |
| 카탈로그 | 앱 내 `/design` 라우트 (컴포넌트별 변형·상태·코드 예제) |
| 테스트 | Go: `httptest` + 인메모리 SQLite |
| 버전 관리 | git, 단계별 커밋 |
| 협업 | Claude가 전부 구현하고 핵심 포인트를 문서·주석으로 설명 |
| 브라우저 검증 | Windows Chrome + BrowserSkill 확장, WSL의 `bsk`와 연결 |

## 컴포넌트 범위 (모두 포함)

- **기본/폼**: Button, Input, Textarea, Select, Combobox, Checkbox, Radio, Switch, Slider, DatePicker, FileUpload, Form 검증
- **데이터 표시**: DataTable(정렬·필터·페이지·선택), Card, Badge, Avatar, Stat, Timeline, Tree
- **차트/대시보드**: Line/Bar/Area/Pie, KPI 타일, Sparkline
- **오버레이/피드백/내비**: Modal, Drawer, Popover, Tooltip, Toast, Alert, Tabs, Sidebar, Breadcrumb, Command palette

## 디렉토리 구조 (안)

```
.
├── cmd/server/          # main
├── internal/
│   ├── http/            # 핸들러, 미들웨어, 라우팅
│   ├── auth/            # 세션, bcrypt
│   ├── store/           # sqlc 생성 코드 + 쿼리
│   └── db/migrations/   # goose SQL
├── web/                 # Vite React 앱
│   └── src/{components/ui,pages,design,i18n,lib}
├── docs/                # 스펙, 학습 노트
└── Makefile
```

## 단계 (안)

0. 환경: Go 업그레이드, git init, bsk 설치 및 Windows Chrome 연결 확인
1. Go 기본 서버: `/api/health`, 구조화 로깅, graceful shutdown, 테스트
2. SQLite + goose + sqlc: 스키마·시드 데이터
3. 인증: 로그인/로그아웃/세션 미들웨어/역할
4. Vite + Tailwind + 토큰/다크모드 + i18n 골격, 앱 셸(Sidebar/Topbar)
5. 컴포넌트 그룹별 구현 + `/design` 카탈로그 (그룹마다 bsk 스크린샷 검증)
6. 이커머스 화면: 대시보드, 사용자·상품·주문 CRUD
7. `embed` 단일 바이너리 빌드
