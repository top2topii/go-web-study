import { ConstructionIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useLocation } from 'react-router'
import { findNav } from '@/config/nav'

/** 아직 구현하지 않은 화면의 자리표시자 */
export function ComingSoonPage({ phase }: { phase: number }) {
  const { t } = useTranslation()
  const current = findNav(useLocation().pathname)

  return (
    <>
      {current && <h1 className="text-2xl font-semibold tracking-tight">{t(`nav.${current.item.key}`)}</h1>}
      <div className="flex flex-1 flex-col items-center justify-center gap-2 rounded-xl border border-dashed py-24 text-center">
        <ConstructionIcon className="size-8 text-muted-foreground" />
        <p className="font-medium">{t('common.comingSoon')}</p>
        <p className="text-sm text-muted-foreground">{t('common.comingSoonDesc', { phase })}</p>
      </div>
    </>
  )
}
