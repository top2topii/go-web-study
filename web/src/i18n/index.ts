import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import en from './locales/en'
import ko from './locales/ko'

export const LANGS = ['ko', 'en'] as const
export type Lang = (typeof LANGS)[number]
export const LANG_LABELS: Record<Lang, string> = { ko: '한국어', en: 'English' }

const STORAGE_KEY = 'lang' // index.html 인라인 스크립트와 같은 키

function initialLang(): Lang {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved === 'ko' || saved === 'en') return saved
  } catch {
    /* localStorage 접근 불가 */
  }
  return navigator.language.toLowerCase().startsWith('ko') ? 'ko' : 'en'
}

// 번역 파일을 번들에 포함한다(HTTP 로딩 없음). Go embed 단일 바이너리 배포가 단순해진다.
void i18n.use(initReactI18next).init({
  resources: { ko: { translation: ko }, en: { translation: en } },
  lng: initialLang(),
  fallbackLng: 'ko',
  interpolation: { escapeValue: false }, // React가 이미 XSS 이스케이프를 한다
})

document.documentElement.lang = i18n.language
i18n.on('languageChanged', (lng) => {
  document.documentElement.lang = lng
  try {
    localStorage.setItem(STORAGE_KEY, lng)
  } catch {
    /* 무시 */
  }
})

export default i18n
