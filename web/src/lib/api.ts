// API 클라이언트: fetch를 얇게 감싸 에러를 ApiError로 통일한다.

export class ApiError extends Error {
  // tsconfig의 erasableSyntaxOnly 때문에 `constructor(public status)` 축약형 대신 필드를 명시한다.
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export const isApiError = (e: unknown, status?: number): e is ApiError =>
  e instanceof ApiError && (status === undefined || e.status === status)

type RequestOptions = Omit<RequestInit, 'body'> & {
  /** 객체를 넘기면 JSON으로 직렬화하고 Content-Type을 붙인다. */
  json?: unknown
}

export async function api<T>(path: string, { json, headers, ...init }: RequestOptions = {}): Promise<T> {
  const h = new Headers(headers)
  let body: BodyInit | undefined
  if (json !== undefined) {
    h.set('Content-Type', 'application/json')
    body = JSON.stringify(json)
  }

  let res: Response
  try {
    // 같은 출처이므로 쿠키는 자동 전송된다(credentials 기본값 same-origin).
    res = await fetch(`/api${path}`, { ...init, headers: h, body })
  } catch {
    throw new ApiError(0, 'network error') // 서버 꺼짐, 오프라인 등
  }

  if (!res.ok) {
    let message = res.statusText
    try {
      message = ((await res.json()) as { error?: string }).error ?? message
    } catch {
      /* JSON이 아닌 응답(예: CSRF 403 text/plain) */
    }
    throw new ApiError(res.status, message)
  }
  if (res.status === 204) return undefined as T
  return (await res.json()) as T
}

// ---- 서버 응답 타입 (internal/store 의 Go struct와 1:1) ----

export type Role = 'admin' | 'viewer'

export interface User {
  id: number
  email: string
  name: string
  role: Role
  created_at: string
  updated_at: string
}

export interface Page<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export interface RevenueRow {
  day: string // YYYY-MM-DD (UTC)
  orders: number
  revenue: number
}
