import { useQuery } from '@tanstack/react-query'
import type { TelemetryItem } from '@/types/telemetry'
import { API_ENDPOINTS } from '@/config/api'

interface ListSchemasResponse {
  schemas: TelemetryItem[]
  total: number
  page: number
  pageSize: number
}

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

const fetchSchemas = async (params: ListSchemasParams): Promise<ListSchemasResponse> => {
  const searchParams = new URLSearchParams({
    page: params.page.toString(),
    pageSize: params.pageSize.toString(),
  })

  if (params.search) {
    searchParams.append('search', params.search)
  }
  if (params.type && params.type !== 'all') {
    searchParams.append('type', params.type)
  }
  if (params.sortField) {
    searchParams.append('sort_field', params.sortField)
  }
  if (params.sortDirection) {
    searchParams.append('sort_direction', params.sortDirection)
  }

  // Add filters to query params
  if (params.filters) {
    Object.entries(params.filters).forEach(([key, values]) => {
      if (values.length > 0) {
        values.forEach(value => {
          searchParams.append(`filter_${key}`, value)
        })
      }
    })
  }

  const response = await fetch(`${API_ENDPOINTS.schemas}?${searchParams.toString()}`)
  if (!response.ok) {
    throw new Error('Failed to fetch schemas')
  }
  return response.json()
}

export const useSchemas = (params: ListSchemasParams) => {
  return useQuery({
    queryKey: ['schemas', params],
    queryFn: () => fetchSchemas(params),
  })
} 