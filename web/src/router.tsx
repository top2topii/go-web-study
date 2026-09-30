import { createBrowserRouter } from 'react-router'
import { AppShell } from '@/components/layout/app-shell'
import { RequireAuth, RequireRole } from '@/features/auth/guards'
import { ComingSoonPage } from '@/pages/coming-soon'
import { DashboardPage } from '@/pages/dashboard'
import { DesignHomePage } from '@/pages/design'
import { NotFoundPage } from '@/pages/errors'
import { LoginPage } from '@/pages/login'

// Data 모드 라우터: 라우트 트리를 객체로 선언한다.
// 레이아웃 라우트(path 없음)는 자식에게 <Outlet />으로 화면을 넘긴다.
export const router = createBrowserRouter([
  { path: '/login', element: <LoginPage /> },
  {
    element: <RequireAuth />, // 로그인 검사
    children: [
      {
        element: <AppShell />, // 사이드바 + 상단바
        children: [
          { index: true, element: <DashboardPage /> },
          { path: 'orders', element: <ComingSoonPage phase={6} /> },
          { path: 'products', element: <ComingSoonPage phase={6} /> },
          { path: 'categories', element: <ComingSoonPage phase={6} /> },
          { path: 'customers', element: <ComingSoonPage phase={6} /> },
          {
            path: 'users',
            element: (
              <RequireRole role="admin">
                <ComingSoonPage phase={6} />
              </RequireRole>
            ),
          },
          { path: 'design', element: <DesignHomePage /> },
          { path: '*', element: <NotFoundPage /> },
        ],
      },
    ],
  },
])
