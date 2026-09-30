import { useEffect, useMemo, useState, useSyncExternalStore, type ReactNode } from 'react'
import { BRANDS, DEFAULT_BRAND, MODES, STORAGE, ThemeContext, type Brand, type Mode } from '@/lib/theme-context'

function read<T extends string>(key: string, allowed: readonly T[], fallback: T): T {
  try {
    const v = localStorage.getItem(key)
    return v && (allowed as readonly string[]).includes(v) ? (v as T) : fallback
  } catch {
    return fallback // 사생활 보호 모드 등에서 localStorage 접근이 막힐 수 있다
  }
}

function write(key: string, value: string) {
  try {
    localStorage.setItem(key, value)
  } catch {
    /* 무시: 저장 실패해도 현재 화면에는 적용된다 */
  }
}

// OS 다크 모드 설정을 외부 저장소처럼 구독한다.
const media = () => window.matchMedia('(prefers-color-scheme: dark)')
const subscribe = (cb: () => void) => {
  const m = media()
  m.addEventListener('change', cb)
  return () => m.removeEventListener('change', cb)
}
const getSystemDark = () => media().matches

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [mode, setModeState] = useState<Mode>(() => read(STORAGE.mode, MODES, 'system'))
  const [brand, setBrandState] = useState<Brand>(() => read(STORAGE.brand, BRANDS, DEFAULT_BRAND))
  const systemDark = useSyncExternalStore(subscribe, getSystemDark)

  const resolved = mode === 'system' ? (systemDark ? 'dark' : 'light') : mode

  // 클래스는 <html>에 둔다: Radix 포털(body 직속)까지 테마가 적용되도록.
  useEffect(() => {
    document.documentElement.classList.toggle('dark', resolved === 'dark')
  }, [resolved])

  useEffect(() => {
    document.documentElement.dataset.brand = brand
  }, [brand])

  const value = useMemo(
    () => ({
      mode,
      resolved,
      brand,
      setMode: (m: Mode) => {
        write(STORAGE.mode, m)
        setModeState(m)
      },
      setBrand: (b: Brand) => {
        write(STORAGE.brand, b)
        setBrandState(b)
      },
    }),
    [mode, resolved, brand],
  )

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}
