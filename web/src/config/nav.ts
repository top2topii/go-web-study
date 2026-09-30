import {
  FolderTreeIcon,
  LayoutDashboardIcon,
  PackageIcon,
  PaletteIcon,
  ShieldIcon,
  ShoppingCartIcon,
  UsersIcon,
  type LucideIcon,
} from 'lucide-react'
import type { Role } from '@/lib/api'

// 사이드바·브레드크럼·문서 제목이 모두 이 설정 하나를 본다(단일 진실 공급원).

export type NavKey =
  | 'dashboard'
  | 'orders'
  | 'products'
  | 'categories'
  | 'customers'
  | 'users'
  | 'designOverview'

export interface NavItem {
  key: NavKey // i18n: nav.<key>
  path: string
  icon: LucideIcon
  role?: Role // 지정 시 해당 역할만 메뉴에 표시
}

export interface NavGroup {
  key: 'operations' | 'management' | 'design' // i18n: nav.groups.<key>
  items: NavItem[]
}

export const NAV: NavGroup[] = [
  {
    key: 'operations',
    items: [
      { key: 'dashboard', path: '/', icon: LayoutDashboardIcon },
      { key: 'orders', path: '/orders', icon: ShoppingCartIcon },
      { key: 'products', path: '/products', icon: PackageIcon },
      { key: 'categories', path: '/categories', icon: FolderTreeIcon },
      { key: 'customers', path: '/customers', icon: UsersIcon },
    ],
  },
  {
    key: 'management',
    items: [{ key: 'users', path: '/users', icon: ShieldIcon, role: 'admin' }],
  },
  {
    key: 'design',
    items: [{ key: 'designOverview', path: '/design', icon: PaletteIcon }],
  },
]

export const isActive = (itemPath: string, pathname: string) =>
  itemPath === '/' ? pathname === '/' : pathname === itemPath || pathname.startsWith(itemPath + '/')

/** 현재 경로에 해당하는 그룹·항목 */
export function findNav(pathname: string): { group: NavGroup; item: NavItem } | undefined {
  for (const group of NAV) {
    const item = group.items.find((i) => isActive(i.path, pathname))
    if (item) return { group, item }
  }
}
