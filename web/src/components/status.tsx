import { Loader2Icon, TriangleAlertIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'

export function FullPageSpinner() {
  return (
    <div className="flex min-h-svh items-center justify-center" role="status" aria-live="polite">
      <Loader2Icon className="size-6 animate-spin text-muted-foreground" />
    </div>
  )
}

export function FullPageError({ onRetry }: { onRetry?: () => void }) {
  const { t } = useTranslation()
  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-4 p-6 text-center">
      <TriangleAlertIcon className="size-8 text-destructive" />
      <p className="text-muted-foreground">{t('errors.network')}</p>
      {onRetry && (
        <Button variant="outline" onClick={onRetry}>
          {t('errors.retry')}
        </Button>
      )}
    </div>
  )
}
