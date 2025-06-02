import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api-client'
import type { ListTelemetrySchemasResponse } from '@/types/telemetry-schema'

interface UseSchemaAssignmentsOptions {
  schemaKey: string
  search?: string
  status?: string[]
  page?: number
  pageSize?: number
}

export const useSchemaAssignments = ({
  schemaKey,
  search,
  status,
  page,
  pageSize,
}: UseSchemaAssignmentsOptions) => {
  return useQuery<ListTelemetrySchemasResponse>({
    queryKey: ['schemaAssignments', schemaKey, search, status, page, pageSize],
    queryFn: async () => {
      return api.telemetries.listSchemas(schemaKey, {
        search,
        status,
        page,
        pageSize,
      })
    },
  })
}
