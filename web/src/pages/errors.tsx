import { CompassIcon, ShieldXIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router'
import { Button } from '@/components/ui/button'

function Message({ icon, title, desc }: { icon: React.ReactNode; title: string; desc: string }) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-3 py-24 text-center">
      <div className="text-muted-foreground">{icon}</div>
      <h1 className="text-xl font-semibold">{title}</h1>
      <p className="text-muted-foreground">{desc}</p>
      <Button asChild variant="outline" className="mt-2">
        <Link to="/">{t('notFound.back')}</Link>
      </Button>
    </div>
  )
}

export function NotFoundPage() {
  const { t } = useTranslation()
  return <Message icon={<CompassIcon className="size-10" />} title={t('notFound.title')} desc={t('notFound.desc')} />
}

export function ForbiddenPage() {
  const { t } = useTranslation()
  return <Message icon={<ShieldXIcon className="size-10" />} title="403" desc={t('errors.forbidden')} />
}
