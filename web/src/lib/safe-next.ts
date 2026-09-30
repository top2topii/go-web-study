/**
 * 로그인 후 돌아갈 ?next= 값은 같은 사이트 경로여야 한다(open redirect 방지).
 * "/orders"는 허용, "//evil.com"·"/\evil.com"·"https://..."은 거부.
 * (브라우저는 "//host"와 "/\host"를 다른 사이트 주소로 해석한다)
 */
export function safeNext(v: string | null): string {
  if (!v || !v.startsWith('/') || v.startsWith('//') || v.startsWith('/\\')) return '/'
  return v
}
