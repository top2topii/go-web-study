import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api, isApiError, type User } from '@/lib/api'
import { meKey } from '@/lib/query-client'

/**
 * 현재 로그인 사용자. 401이면 에러가 아니라 null(로그아웃 상태)을 돌려준다.
 * data: undefined = 아직 모름(로딩), null = 비로그인, User = 로그인
 */
export function useMe() {
  return useQuery({
    queryKey: meKey,
    queryFn: async () => {
      try {
        return await api<User>('/auth/me')
      } catch (e) {
        if (isApiError(e, 401)) return null
        throw e
      }
    },
    staleTime: 5 * 60_000,
  })
}

/** 보호된 화면 안에서만 쓴다. RequireAuth가 로그인을 보장한다. */
export function useCurrentUser(): User {
  const { data } = useMe()
  if (!data) throw new Error('useCurrentUser outside of RequireAuth')
  return data
}

export function useLogin() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: { email: string; password: string }) =>
      api<User>('/auth/login', { method: 'POST', json: body }),
    onSuccess: (user) => {
      qc.clear() // 이전 사용자의 캐시가 섞이지 않도록
      qc.setQueryData(meKey, user)
    },
  })
}

export function useLogout() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => api<void>('/auth/logout', { method: 'POST' }),
    onSettled: () => {
      // 성공·실패와 무관하게 클라이언트 상태는 비운다.
      qc.clear()
      qc.setQueryData(meKey, null)
    },
  })
}
