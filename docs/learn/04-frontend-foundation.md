# 04. 프론트엔드 기반: Vite + React + Tailwind v4 + shadcn/ui

4단계 코드: `web/`

## 스택과 버전 (2026-09 기준)

| 영역 | 선택 | 메모 |
|---|---|---|
| 빌드 | Vite 8 (Rolldown) | `pnpm create vite web --template react-ts` |
| UI | React 19 + TypeScript 6 | TS 7(Go로 재작성된 컴파일러)이 나와 있지만 템플릿 기본값인 6 유지 |
| 스타일 | Tailwind CSS 4 | 설정 파일 없이 CSS 안에서 `@theme`로 설정 |
| 컴포넌트 | shadcn/ui 4 (`radix-nova` 스타일, Radix 기반) | 코드를 `src/components/ui/`에 복사해 소유 |
| 라우팅 | React Router 8 (Data 모드) | `RouterProvider`만 `react-router/dom`에서 import |
| 서버 상태 | TanStack Query 5 | 캐시, 재시도, 로딩/에러 상태 |
| i18n | i18next 26 + react-i18next 17 | 번역을 번들에 포함, 키 타입 검사 |
| 폰트 | Pretendard Variable (dynamic subset) | 필요한 한글 조각만 내려받음 |
| 린트 | oxlint | Rust로 만든 고속 린터(템플릿 기본) |

## 1. 개발 구조

```
브라우저 → http://localhost:5173 (Vite)
             ├─ /src/...  → Vite가 즉시 변환(HMR)
             └─ /api/...  → 프록시 → http://127.0.0.1:8080 (Go)
```

- 브라우저 입장에서는 페이지와 API가 **같은 출처(localhost:5173)**다. 그래서 CORS 설정이 필요 없고, 쿠키와 `CrossOriginProtection`도 운영 환경과 똑같이 동작한다(실측: 프록시를 통한 로그인 POST → 200).
- `make dev`: Go 서버와 Vite를 함께 띄운다. Ctrl+C 한 번이면 둘 다 종료된다(`trap 'kill 0'`).
- 7단계에서 `vite build` 결과물을 Go `embed`로 넣으면 단일 바이너리가 된다.

## 2. 디자인 토큰 — `web/src/index.css`

**의미(semantic) 토큰**만 쓰는 것이 핵심이다. 컴포넌트는 `bg-blue-600`이 아니라 `bg-primary`를 쓴다.

```
:root / .dark          ← 토큰 값 (라이트/다크)
[data-brand=violet]    ← 브랜드 색만 덮어씀 (--brand)
@theme inline          ← 토큰 → Tailwind 클래스 (bg-primary, text-success ...)
```

| 토큰 | 용도 |
|---|---|
| `background` / `foreground` | 페이지 바탕, 기본 글자 |
| `card`, `popover` | 떠 있는 표면 |
| `primary` | 브랜드. 주요 버튼, 활성 상태 |
| `secondary`, `muted`, `accent` | 보조 표면. 호버, 비활성 텍스트 |
| `destructive` | 삭제, 에러 |
| `success`, `warning`, `info` | **직접 추가**: 주문 상태 배지 등 admin 필수 색 |
| `border`, `input`, `ring` | 테두리, 입력창, 포커스 링 |
| `sidebar-*` | 사이드바 전용 |

### 색 공간: OKLCH

`oklch(명도 채도 색상각)`. RGB/HSL과 달리 **명도가 사람 눈에 균일하다.** 같은 L 값이면 어떤 색이든 체감 밝기가 비슷해서, 다크 모드용 색을 만들 때 L만 올리면 된다.

### 함정 1: 클래스는 `<html>`에

드롭다운, 시트, 툴팁은 **포털**로 `body` 바로 아래에 렌더링된다(`#root` 밖). 그래서 `.dark`와 `data-brand`는 반드시 `document.documentElement`에 건다. **실측: 다크 모드에서 테마 메뉴(포털)에 다크 스타일이 정상 적용됨.**

### 함정 2: CSS 명시도(specificity)

`:root[data-brand]` (0,2,0)는 `.dark` (0,1,0)를 이긴다. 그래서 다크 모드에서도 라이트용 브랜드 색이 적용될 수 있다.
→ 브랜드마다 라이트 블록과 `.dark[data-brand=x]` 블록(0,2,0, 뒤에 위치)을 둔다.

**실측** (`getComputedStyle(html).getPropertyValue('--primary')`):
```
light/blue    → oklch(0.546 0.245 262.881)   dark/blue    → oklch(0.623 0.214 259.815)
light/neutral → oklch(0.205 0 0)             dark/neutral → oklch(0.922 0 0)
light/emerald → oklch(0.508 0.118 165.612)   dark/emerald → oklch(0.696 0.17 162.48)
```

### 함정 3: shadcn 새 버전의 variant 클래스

Separator 기본 클래스가 `data-vertical:self-stretch`다. 여기에 예전 예제대로 `h-4`를 덧붙이면 높이는 고정되는데 늘이기(stretch)는 적용되지 않아, 요소가 위쪽 끝에 붙는다. → **같은 변형으로** `data-vertical:self-center`를 줘야 이긴다(명시도가 같고 나중에 온 규칙이 이김).

## 3. 깜빡임(FOUC) 방지 — `web/index.html`

React는 JS가 로드된 **뒤에** 실행된다. 그 사이 브라우저는 기본(라이트) 화면을 먼저 그린다. `<head>`의 인라인 스크립트가 localStorage를 읽어 `class="dark"`와 `data-brand`를 **첫 페인트 전에** 적용한다. 잘못된 값이 저장돼 있어도 기본값으로 돌아간다(실측: 테스트 스크립트 버그로 `"dark emerald"`가 저장됐을 때 system과 blue로 복구됨).

## 4. 서버 상태 — TanStack Query

```ts
useQuery({ queryKey: ['stats', 'revenue', 30], queryFn: () => api('/stats/revenue?days=30') })
```

- **queryKey**: 캐시 주소. 같은 키를 쓰는 컴포넌트는 요청 1번을 공유한다.
- **retry 설정**: 기본값은 3회 재시도(약 7초)다. 4xx는 다시 시도해도 결과가 같으므로 즉시 실패시킨다. 이게 없으면 비로그인 상태에서 로그인 화면으로 넘어가는 데 7초가 걸린다.
- **전역 401 처리**: `QueryCache.onError`에서 401을 받으면 `me` 쿼리를 `null`로 바꾸고, `RequireAuth`가 로그인 화면으로 보낸다.
- `useMe()`: `undefined`는 로딩 중, `null`은 비로그인, `User`는 로그인 상태. 401을 에러가 아닌 **정상 상태(null)**로 모델링한다.
- 로그인/로그아웃 시 `queryClient.clear()`: 이전 사용자의 캐시가 섞이는 것을 막는다.

## 5. 라우팅과 권한 — `web/src/router.tsx`

```
/login                     공개. 이미 로그인 상태면 next로 이동
<RequireAuth>              비로그인 → /login?next=/orders
  <AppShell>               Sidebar + 상단바 + <Outlet/>
    /                      대시보드
    /users                 <RequireRole role="admin">
    *                      404
```

- **레이아웃 라우트**: `path` 없이 `element`만 있는 라우트다. 자식을 `<Outlet />`에 렌더링한다.
- **open redirect 방지** (`lib/safe-next.ts`): `?next=//evil.com`을 그대로 따르면 로그인 직후 피싱 사이트로 이동한다. `/`로 시작하되 `//`와 `/\`로 시작하지 않는 값만 허용한다.
- 프론트의 역할 검사는 **편의 기능**이다. 실제 보안은 서버의 `requireRole`이 담당한다(실측: viewer로 `/users` → 메뉴 숨김 + 403 화면).

## 6. 단일 진실 공급원 — `web/src/config/nav.ts`

사이드바 메뉴, 브레드크럼, 문서 제목이 모두 `NAV` 설정 하나에서 나온다. 메뉴를 추가하면 세 곳에 자동으로 반영된다.

## 7. i18n 타입 안전성

```ts
const en: typeof ko = { ... }             // 키가 빠지면 컴파일 에러
declare module 'i18next' { interface CustomTypeOptions { resources: { translation: typeof ko } } }
t('nav.dashbord')                          // 오타 → 컴파일 에러
```

- 서버의 영어 에러 메시지를 그대로 보여주지 않는다. **HTTP 상태 코드별로 번역 문구**를 고른다(`login.tsx`).
- 금액은 `Intl.NumberFormat`으로 포맷하고 언어별로 캐시한다(`lib/format.ts`). `formatCompact`는 한국어 `586.4만`, 영어 `5.9M`처럼 나온다.

## 8. TypeScript 설정이 주는 제약

| 옵션 | 영향 |
|---|---|
| `erasableSyntaxOnly` | `enum`과 `constructor(public x)` 금지. 타입만 지우면 JS가 되는 문법만 허용한다(Node의 타입 제거 실행과 호환) |
| `verbatimModuleSyntax` | 타입은 `import type`으로 가져와야 한다 |
| `noUnusedLocals/Parameters` | 안 쓰는 변수는 에러 |

## 9. React 규칙 (oxlint가 잡아준 것)

- **렌더링 중 `new Date()`는 불순(impure)하다**: 렌더링할 때마다 값이 달라진다. → `useState(() => new Date()...)` 초기화 함수로 한 번만 계산한다.
- **effect 안에서 setState 호출** (shadcn의 `use-mobile` 원본): 외부 상태(미디어 쿼리) 구독은 `useSyncExternalStore`가 정석이다. 첫 렌더링부터 정확한 값이 나오고 불필요한 재렌더링도 없다. 테마 Provider의 OS 다크 모드 감지도 같은 방식이다.

## 10. shadcn CLI 사용 기록

```sh
pnpm dlx shadcn@latest init -b radix -t vite -p nova   # 프리셋: nova, vega, maia, lyra, mira, luma, sera, rhea
pnpm dlx shadcn@latest add sidebar card input ...      # 컴포넌트 추가 = 코드 복사
```

- `cn` 패키지: shadcn이 만든 `clsx`와 `tailwind-merge` 대체 패키지다. 설치 전에 설치 스크립트(install hook)가 없는 것, 저장소가 `shadcn-ui/cn`인 것을 확인했다.
- `sonner`가 Next.js 전용 `next-themes`를 끌고 와서, 자체 `useTheme`로 교체하고 의존성을 제거했다.

## 11. BrowserSkill로 검증하며 알게 된 것

- **비밀번호 입력란이 있는 페이지**에서는 다른 확장(비밀번호 관리자)이 프레임을 삽입해 Chrome이 BrowserSkill의 접근(CDP)을 차단한다. `/login`만 막히고 나머지 페이지는 정상이다.
- 우회 방법: 비밀번호 입력란이 없는 페이지에서 `fetch('/api/auth/login')`으로 로그인한 뒤 보호된 화면을 검증한다.
- `pkill -f 'cmd/server'`는 그 문자열이 들어 있는 **자기 자신의 셸 명령**까지 종료시킨다.

## 명령어

```sh
make dev        # Go + Vite → http://localhost:5173
make check      # Go(gofmt, vet, race test) + web(oxlint, tsc, vite build)
pnpm -C web dlx shadcn@latest add <component>
```
