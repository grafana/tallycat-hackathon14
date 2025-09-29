import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api-client'

interface UseSchemaDetailsOptions {
  schemaKey: string
  schemaId: string
  enabled?: boolean
}

export const useSchemaDetails = ({
  schemaKey,
  schemaId,
  enabled = true,
}: UseSchemaDetailsOptions) => {
  return useQuery({
    queryKey: ['schema', schemaKey, schemaId],
    queryFn: async () => {
      try {
        return await api.telemetries.getSchemaById(schemaKey, schemaId)
      } catch (error) {
        if (
          error instanceof Error &&
          'status' in error &&
          (error as { status?: number }).status === 404
        ) {
          return null
        }
        throw error
      }
    },
    enabled,
  })
} 