import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api-client'

interface UseScopesParams {
  page?: number
  pageSize?: number
  search?: string
}

export const useScopes = (params: UseScopesParams = {}) => {
  const {
    page = 1,
    pageSize = 100, // Get more scopes by default for the modal
    search = '',
  } = params

  return useQuery({
    queryKey: ['scopes', { page, pageSize, search }],
    queryFn: () =>
      api.scopes.list({
        page,
        pageSize,
        search,
      }),
    staleTime: 30000, // 30 seconds
    refetchOnWindowFocus: false,
  })
}
