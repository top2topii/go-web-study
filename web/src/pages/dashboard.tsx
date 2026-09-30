import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { BanknoteIcon, CalendarIcon, ReceiptIcon, ShoppingCartIcon, type LucideIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { FullPageError } from '@/components/status'
import { Card, CardAction, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { api, type RevenueRow } from '@/lib/api'
import { formatCompact, formatCurrency, formatNumber } from '@/lib/format'

const DAYS = 30

export function DashboardPage() {
  const { t, i18n } = useTranslation()
  const lang = i18n.language
  // 렌더 중 Date 호출은 "불순(impure)"하다: 렌더마다 값이 달라질 수 있다.
  // useState 초기화 함수는 첫 렌더에 한 번만 실행된다. (서버와 같은 UTC 날짜)
  const [today] = useState(() => new Date().toISOString().slice(0, 10))
  const q = useQuery({
    queryKey: ['stats', 'revenue', DAYS],
    queryFn: () => api<RevenueRow[]>(`/stats/revenue?days=${DAYS}`),
  })

  // 간단한 집계는 클라이언트에서. (6단계에서 서버 요약 API로 옮길 수 있다)
  const rows = q.data ?? []
  const revenue = rows.reduce((s, r) => s + r.revenue, 0)
  const orders = rows.reduce((s, r) => s + r.orders, 0)
  const todayRevenue = rows.find((r) => r.day === today)?.revenue ?? 0

  const kpis: { label: string; value: string; hint?: string; icon: LucideIcon }[] = [
    {
      label: t('dashboard.revenue'),
      value: formatCurrency(revenue, lang),
      hint: t('dashboard.perDay', { value: formatCompact(revenue / DAYS, lang) }),
      icon: BanknoteIcon,
    },
    { label: t('dashboard.orders'), value: formatNumber(orders, lang), icon: ShoppingCartIcon },
    { label: t('dashboard.aov'), value: formatCurrency(orders ? Math.round(revenue / orders) : 0, lang), icon: ReceiptIcon },
    { label: t('dashboard.today'), value: formatCurrency(todayRevenue, lang), icon: CalendarIcon },
  ]

  return (
    <>
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">{t('dashboard.title')}</h1>
        <p className="text-sm text-muted-foreground">{t('dashboard.subtitle', { days: DAYS })}</p>
      </div>

      {q.isError ? (
        <FullPageError onRetry={() => void q.refetch()} />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
          {kpis.map((k) => (
            <Card key={k.label}>
              <CardHeader>
                <CardDescription>{k.label}</CardDescription>
                <CardAction>
                  <k.icon className="size-4 text-muted-foreground" />
                </CardAction>
                {q.isPending ? (
                  <Skeleton className="h-8 w-32" />
                ) : (
                  <CardTitle className="text-2xl font-semibold tabular-nums">{k.value}</CardTitle>
                )}
                {k.hint && !q.isPending && <p className="text-xs text-muted-foreground">{k.hint}</p>}
              </CardHeader>
            </Card>
          ))}
        </div>
      )}
    </>
  )
}
