import type { ReactNode } from 'react'
import { Navigate, Outlet, useLocation } from 'react-router'
import type { Role } from '@/lib/api'
import { FullPageError, FullPageSpinner } from '@/components/status'
import { ForbiddenPage } from '@/pages/errors'
import { useMe } from './queries'

/** 로그인이 필요한 라우트의 부모. 비로그인이면 /login?next=현재경로 로 보낸다. */
export function RequireAuth() {
  const { data: me, isPending, isError, refetch } = useMe()
  const loc = useLocation()

  if (isPending) return <FullPageSpinner />
  if (isError) return <FullPageError onRetry={() => void refetch()} />
  if (!me) {
    const next = loc.pathname + loc.search
    return <Navigate to={`/login?next=${encodeURIComponent(next)}`} replace />
  }
  return <Outlet />
}

/** 역할 검사. 화면 차원의 편의 기능일 뿐, 실제 권한은 서버(requireRole)가 강제한다. */
export function RequireRole({ role, children }: { role: Role; children: ReactNode }) {
  const { data: me } = useMe()
  if (me?.role !== role) return <ForbiddenPage />
  return children
}
