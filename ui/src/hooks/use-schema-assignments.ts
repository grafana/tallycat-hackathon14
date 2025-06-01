import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api-client'
import type { ListSchemaAssignmentsResponse, Schema } from '@/types/schema-catalog'

interface UseSchemaAssignmentsOptions {
  schemaKey: string
  search?: string
  status?: string[]
  page?: number
  pageSize?: number
}

export const useSchemaAssignments = ({ schemaKey, search, status, page, pageSize }: UseSchemaAssignmentsOptions) => {
  return useQuery<ListSchemaAssignmentsResponse>({
    queryKey: ['schemaAssignments', schemaKey, search, status, page, pageSize],
    queryFn: async () => {
      return api.schemas.listAssignments(schemaKey, { search, status, page, pageSize })
    },
  })
} 