import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api-client'

export type SortDirection = 'asc' | 'desc'
export type SortField = 'name' | 'type' | 'dataType' | 'lastUpdated'

interface ListSchemasParams {
  page: number
  pageSize: number
  search?: string
  type?: string
  sortField?: SortField
  sortDirection?: SortDirection
  filters?: Record<string, string[]>
}

export const useSchemas = (params: ListSchemasParams) => {
  return useQuery({
    queryKey: ['schemas', params],
    queryFn: () => api.telemetries.listWithParams(params),
  })
}
