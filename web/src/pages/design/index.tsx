import { useTranslation } from 'react-i18next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

// 5단계에서 컴포넌트 카탈로그로 확장한다. 지금은 토큰이 제대로 동작하는지 보는 최소 견본.
const SWATCHES = [
  ['primary', 'bg-primary text-primary-foreground'],
  ['secondary', 'bg-secondary text-secondary-foreground'],
  ['muted', 'bg-muted text-muted-foreground'],
  ['accent', 'bg-accent text-accent-foreground'],
  ['destructive', 'bg-destructive text-white'],
  ['success', 'bg-success text-success-foreground'],
  ['warning', 'bg-warning text-warning-foreground'],
  ['info', 'bg-info text-info-foreground'],
] as const

export function DesignHomePage() {
  const { t } = useTranslation()
  return (
    <>
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">{t('design.title')}</h1>
        <p className="text-sm text-muted-foreground">{t('design.subtitle')}</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Color tokens</CardTitle>
          <CardDescription>테마·브랜드를 바꾸면 즉시 반영됩니다.</CardDescription>
        </CardHeader>
        <CardContent className="grid grid-cols-2 gap-3 sm:grid-cols-4">
          {SWATCHES.map(([name, cls]) => (
            <div key={name} className={`flex h-16 items-end rounded-lg border p-2 text-xs font-medium ${cls}`}>
              {name}
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Button · Badge</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-wrap items-center gap-2">
          <Button>Default</Button>
          <Button variant="secondary">Secondary</Button>
          <Button variant="outline">Outline</Button>
          <Button variant="ghost">Ghost</Button>
          <Button variant="destructive">Destructive</Button>
          <Button variant="link">Link</Button>
          <Badge>Badge</Badge>
          <Badge variant="secondary">Secondary</Badge>
          <Badge variant="outline">Outline</Badge>
          <Badge variant="destructive">Destructive</Badge>
        </CardContent>
      </Card>
    </>
  )
}
