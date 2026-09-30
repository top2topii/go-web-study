import { createContext, useContext } from 'react'

export const MODES = ['light', 'dark', 'system'] as const
export type Mode = (typeof MODES)[number]

export const BRANDS = ['blue', 'violet', 'emerald', 'rose', 'orange', 'neutral'] as const
export type Brand = (typeof BRANDS)[number]

// index.html의 인라인 스크립트와 같은 키를 쓴다.
export const STORAGE = { mode: 'theme', brand: 'brand' } as const
export const DEFAULT_BRAND: Brand = 'blue'

export interface ThemeState {
  mode: Mode
  /** system일 때 실제 적용된 값 */
  resolved: 'light' | 'dark'
  brand: Brand
  setMode: (m: Mode) => void
  setBrand: (b: Brand) => void
}

export const ThemeContext = createContext<ThemeState | null>(null)

export function useTheme() {
  const ctx = useContext(ThemeContext)
  if (!ctx) throw new Error('useTheme must be used within <ThemeProvider>')
  return ctx
}
