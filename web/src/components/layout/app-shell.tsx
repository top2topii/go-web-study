import { Fragment, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Link, Outlet, useLocation } from 'react-router'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { Separator } from '@/components/ui/separator'
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
import { findNav } from '@/config/nav'
import { AppSidebar } from './app-sidebar'
import { LanguageMenu, ThemeMenu } from './theme-menu'

// shadcn Sidebar는 펼침 상태를 이 쿠키에 쓰지만, SPA에서는 읽어오지 않는다(Next.js 서버 렌더링 전제).
// 그래서 첫 렌더링 때 직접 읽어 defaultOpen으로 넘긴다.
function readSidebarOpen() {
  return !document.cookie.split('; ').includes('sidebar_state=false')
}

export function AppShell() {
  const { t } = useTranslation()
  const { pathname } = useLocation()
  const current = findNav(pathname)

  // 문서 제목: "주문 · ShopAdmin"
  useEffect(() => {
    const page = current ? t(`nav.${current.item.key}`) : undefined
    document.title = page ? `${page} · ${t('app.name')}` : t('app.name')
  }, [current, t])

  const crumbs = current
    ? [{ label: t(`nav.groups.${current.group.key}`) }, { label: t(`nav.${current.item.key}`) }]
    : []

  return (
    <SidebarProvider defaultOpen={readSidebarOpen()}>
      <AppSidebar />
      <SidebarInset>
        <header className="sticky top-0 z-10 flex h-14 shrink-0 items-center gap-2 border-b bg-background/80 px-4 backdrop-blur">
          <SidebarTrigger className="-ml-1" aria-label={t('nav.toggleSidebar')} />
          {/* 기본 스타일 data-vertical:self-stretch 를 같은 변형으로 덮어써야 이긴다(명시도 동일 → 가운데 정렬) */}
          <Separator orientation="vertical" className="mr-2 data-vertical:h-4 data-vertical:self-center" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem className="hidden md:block">
                <BreadcrumbLink asChild>
                  <Link to="/">{t('nav.home')}</Link>
                </BreadcrumbLink>
              </BreadcrumbItem>
              {crumbs.map((c, i) => (
                <Fragment key={i}>
                  <BreadcrumbSeparator className="hidden md:block" />
                  <BreadcrumbItem className={i < crumbs.length - 1 ? 'hidden md:block' : undefined}>
                    {i === crumbs.length - 1 ? <BreadcrumbPage>{c.label}</BreadcrumbPage> : c.label}
                  </BreadcrumbItem>
                </Fragment>
              ))}
            </BreadcrumbList>
          </Breadcrumb>
          <div className="ml-auto flex items-center gap-1">
            <LanguageMenu />
            <ThemeMenu />
          </div>
        </header>
        <main className="flex flex-1 flex-col gap-6 p-4 md:p-6">
          <Outlet />
        </main>
      </SidebarInset>
    </SidebarProvider>
  )
}
