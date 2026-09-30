import { StoreIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Link, useLocation } from 'react-router'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
  useSidebar,
} from '@/components/ui/sidebar'
import { isActive, NAV } from '@/config/nav'
import { useCurrentUser } from '@/features/auth/queries'
import { NavUser } from './nav-user'

export function AppSidebar() {
  const { t } = useTranslation()
  const { pathname } = useLocation()
  const user = useCurrentUser()
  const { isMobile, setOpenMobile } = useSidebar()

  // 모바일에서는 메뉴를 누르면 시트(서랍)를 닫는다.
  const onNavigate = () => {
    if (isMobile) setOpenMobile(false)
  }

  return (
    // collapsible="icon": 접으면 아이콘만 남는 좁은 사이드바(툴팁으로 이름 표시)
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild tooltip={t('app.name')}>
              <Link to="/" onClick={onNavigate}>
                <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                  <StoreIcon className="size-4" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold">{t('app.name')}</span>
                  <span className="truncate text-xs text-muted-foreground">{t('app.tagline')}</span>
                </div>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent>
        {NAV.map((group) => {
          // 역할이 맞지 않는 항목은 숨긴다(서버도 403으로 막는다).
          const items = group.items.filter((i) => !i.role || i.role === user.role)
          if (items.length === 0) return null
          return (
            <SidebarGroup key={group.key}>
              <SidebarGroupLabel>{t(`nav.groups.${group.key}`)}</SidebarGroupLabel>
              <SidebarMenu>
                {items.map((item) => (
                  <SidebarMenuItem key={item.key}>
                    <SidebarMenuButton asChild isActive={isActive(item.path, pathname)} tooltip={t(`nav.${item.key}`)}>
                      <Link to={item.path} onClick={onNavigate}>
                        <item.icon />
                        <span>{t(`nav.${item.key}`)}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroup>
          )
        })}
      </SidebarContent>

      <SidebarFooter>
        <NavUser />
      </SidebarFooter>
      {/* 사이드바 가장자리를 클릭해 접기/펴기 */}
      <SidebarRail />
    </Sidebar>
  )
}
