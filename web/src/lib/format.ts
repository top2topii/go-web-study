// 숫자·통화·날짜 포맷. 언어에 따라 Intl 로케일을 바꾼다.
// Intl.NumberFormat 생성은 비싸므로 로케일별로 캐시한다.

const localeOf = (lang: string) => (lang.startsWith('ko') ? 'ko-KR' : 'en-US')

const cache = new Map<string, Intl.NumberFormat>()
function nf(key: string, lang: string, opts: Intl.NumberFormatOptions) {
  const k = `${key}:${lang}`
  let f = cache.get(k)
  if (!f) {
    f = new Intl.NumberFormat(localeOf(lang), opts)
    cache.set(k, f)
  }
  return f
}

/** 원화: ko → "₩3,307,000", en → "₩3,307,000" */
export const formatCurrency = (v: number, lang: string) =>
  nf('krw', lang, { style: 'currency', currency: 'KRW' }).format(v)

/** 큰 금액 축약: ko → "330만", en → "3.3M" */
export const formatCompact = (v: number, lang: string) =>
  nf('compact', lang, { notation: 'compact', maximumFractionDigits: 1 }).format(v)

export const formatNumber = (v: number, lang: string) => nf('num', lang, {}).format(v)
