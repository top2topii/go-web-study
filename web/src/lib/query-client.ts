import { QueryCache, QueryClient } from '@tanstack/react-query'
import { isApiError } from './api'

export const meKey = ['auth', 'me'] as const

export const queryClient: QueryClient = new QueryClient({
  // 어떤 쿼리든 401이 나면 세션이 끝난 것 → me를 null로 바꾸면 RequireAuth가 로그인으로 보낸다.
  queryCache: new QueryCache({
    onError: (err) => {
      if (isApiError(err, 401)) queryClient.setQueryData(meKey, null)
    },
  }),
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      // 기본값은 실패 시 3회 재시도(백오프 포함 ~7초). 4xx는 재시도해도 결과가 같으므로 즉시 실패.
      retry: (count, err) => !(isApiError(err) && err.status >= 400 && err.status < 500) && count < 2,
    },
  },
})
