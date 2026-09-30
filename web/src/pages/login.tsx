import { useEffect, useState, type FormEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { Navigate, useNavigate, useSearchParams } from 'react-router'
import { Loader2Icon, StoreIcon, TriangleAlertIcon } from 'lucide-react'
import { LanguageMenu, ThemeMenu } from '@/components/layout/theme-menu'
import { FullPageSpinner } from '@/components/status'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { useLogin, useMe } from '@/features/auth/queries'
import { isApiError } from '@/lib/api'
import { safeNext } from '@/lib/safe-next'

export function LoginPage() {
  const { t } = useTranslation()
  const [params] = useSearchParams()
  const next = safeNext(params.get('next'))
  const navigate = useNavigate()
  const me = useMe()
  const login = useLogin()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  useEffect(() => {
    document.title = `${t('login.title')} · ${t('app.name')}`
  }, [t])

  if (me.isPending) return <FullPageSpinner />
  // 이미 로그인 상태면 로그인 화면을 보여줄 이유가 없다.
  if (me.data) return <Navigate to={next} replace />

  const onSubmit = (e: FormEvent) => {
    e.preventDefault()
    login.mutate({ email, password }, { onSuccess: () => void navigate(next, { replace: true }) })
  }

  // 서버의 영어 에러 문구 대신 상태 코드별 번역 문구를 보여준다.
  const errorMessage = !login.error
    ? null
    : isApiError(login.error, 401)
      ? t('login.invalid')
      : isApiError(login.error, 400)
        ? t('errors.badRequest')
        : isApiError(login.error, 0)
          ? t('errors.network')
          : t('errors.generic')

  return (
    <div className="grid min-h-svh lg:grid-cols-2">
      {/* 왼쪽: 브랜드 패널(넓은 화면에서만) */}
      <div className="relative hidden flex-col justify-between overflow-hidden bg-primary p-10 text-primary-foreground lg:flex">
        <div className="flex items-center gap-2 text-lg font-semibold">
          <StoreIcon className="size-6" />
          {t('app.name')}
        </div>
        <p className="max-w-sm text-3xl leading-snug font-semibold text-balance">{t('app.tagline')}</p>
        {/* 장식용 원: 브랜드 색 위에 반투명 흰색 */}
        <div aria-hidden className="absolute -right-24 -bottom-24 size-96 rounded-full bg-primary-foreground/10" />
        <div aria-hidden className="absolute right-24 bottom-40 size-40 rounded-full bg-primary-foreground/10" />
      </div>

      {/* 오른쪽: 폼 */}
      <div className="flex flex-col p-4 md:p-10">
        <div className="flex justify-end gap-1">
          <LanguageMenu />
          <ThemeMenu />
        </div>
        <div className="flex flex-1 items-center justify-center">
          <Card className="w-full max-w-sm">
            <CardHeader>
              <CardTitle className="text-xl">{t('login.title')}</CardTitle>
              <CardDescription>{t('login.subtitle')}</CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={onSubmit} noValidate={false}>
                <FieldGroup>
                  {errorMessage && (
                    <Alert variant="destructive" role="alert">
                      <TriangleAlertIcon />
                      <AlertDescription>{errorMessage}</AlertDescription>
                    </Alert>
                  )}
                  <Field>
                    <FieldLabel htmlFor="email">{t('login.email')}</FieldLabel>
                    <Input
                      id="email"
                      type="email"
                      autoComplete="username"
                      required
                      autoFocus
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      aria-invalid={isApiError(login.error, 401) || undefined}
                    />
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="password">{t('login.password')}</FieldLabel>
                    <Input
                      id="password"
                      type="password"
                      autoComplete="current-password"
                      required
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      aria-invalid={isApiError(login.error, 401) || undefined}
                    />
                  </Field>
                  <Button type="submit" className="w-full" disabled={login.isPending}>
                    {login.isPending && <Loader2Icon className="animate-spin" />}
                    {login.isPending ? t('login.submitting') : t('login.submit')}
                  </Button>
                </FieldGroup>
              </form>
              {import.meta.env.DEV && <p className="mt-6 text-xs text-muted-foreground">{t('login.demoHint')}</p>}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
