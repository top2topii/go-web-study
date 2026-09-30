// t('nav.dashboard') 같은 키를 타입 검사한다. 없는 키를 쓰면 컴파일 에러.
import 'i18next'
import type ko from './locales/ko'

declare module 'i18next' {
  interface CustomTypeOptions {
    defaultNS: 'translation'
    resources: { translation: typeof ko }
  }
}
