import { LanguagesIcon, MonitorIcon, MoonIcon, PaletteIcon, SunIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { LANG_LABELS, LANGS } from '@/i18n'
import { BRANDS, MODES, useTheme, type Brand, type Mode } from '@/lib/theme-context'

const MODE_ICON = { light: SunIcon, dark: MoonIcon, system: MonitorIcon } as const

/** 모드(라이트/다크/시스템) + 브랜드 색 선택 */
export function ThemeMenu() {
  const { t } = useTranslation()
  const { mode, resolved, brand, setMode, setBrand } = useTheme()
  const Icon = resolved === 'dark' ? MoonIcon : SunIcon

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" aria-label={t('theme.label')}>
          <Icon />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48">
        <DropdownMenuLabel>{t('theme.mode')}</DropdownMenuLabel>
        <DropdownMenuRadioGroup value={mode} onValueChange={(v) => setMode(v as Mode)}>
          {MODES.map((m) => {
            const MIcon = MODE_ICON[m]
            return (
              <DropdownMenuRadioItem key={m} value={m}>
                <MIcon />
                {t(`theme.${m}`)}
              </DropdownMenuRadioItem>
            )
          })}
        </DropdownMenuRadioGroup>
        <DropdownMenuSeparator />
        <DropdownMenuLabel className="flex items-center gap-1.5">
          <PaletteIcon className="size-3.5" />
          {t('theme.brand')}
        </DropdownMenuLabel>
        <DropdownMenuRadioGroup value={brand} onValueChange={(v) => setBrand(v as Brand)}>
          {BRANDS.map((b) => (
            <DropdownMenuRadioItem key={b} value={b}>
              {/* 스와치: data-brand를 이 span에 걸어 해당 브랜드의 --brand를 미리 보여준다 */}
              <span
                data-brand={b}
                className="size-3.5 rounded-full border bg-(--brand) data-[brand=neutral]:bg-foreground"
                aria-hidden
              />
              {t(`theme.brands.${b}`)}
            </DropdownMenuRadioItem>
          ))}
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

export function LanguageMenu() {
  const { t, i18n } = useTranslation()
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" aria-label={t('lang.label')}>
          <LanguagesIcon />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuRadioGroup value={i18n.language} onValueChange={(v) => void i18n.changeLanguage(v)}>
          {LANGS.map((l) => (
            <DropdownMenuRadioItem key={l} value={l} lang={l}>
              {LANG_LABELS[l]}
            </DropdownMenuRadioItem>
          ))}
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
