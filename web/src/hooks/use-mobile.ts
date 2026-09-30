import { useSyncExternalStore } from "react"

const MOBILE_BREAKPOINT = 768
const query = `(max-width: ${MOBILE_BREAKPOINT - 1}px)`

// shadcn 원본은 useEffect + setState였다. 브라우저 상태(미디어 쿼리) 구독은
// useSyncExternalStore가 정석: 첫 렌더부터 정확한 값, 불필요한 재렌더 없음.
const subscribe = (cb: () => void) => {
  const mql = window.matchMedia(query)
  mql.addEventListener("change", cb)
  return () => mql.removeEventListener("change", cb)
}

export function useIsMobile() {
  return useSyncExternalStore(subscribe, () => window.matchMedia(query).matches)
}
