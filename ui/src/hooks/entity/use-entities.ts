import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api-client'

interface ListTelemetriesParams {
  page: number
  pageSize: number
  search?: string
  sortField?: string
  sortDirection?: 'asc' | 'desc'
  filters?: Record<string, string[]>
  type?: string
}

/**
 * Hook to fetch all telemetries with their associated entities
 * This is the foundation for the Entity Catalog aggregation
 */
export const useEntities = () => {
  return useQuery({
    queryKey: ['entities', 'all-telemetries'],
    queryFn: async () => {
      // Fetch all telemetries with a large page size to get complete data
      // For Entity Catalog, we need all telemetries to aggregate properly
      const params: ListTelemetriesParams = {
        page: 1,
        pageSize: 10000000, // Large enough to get all telemetries in one request
      }
      
      const response = await api.telemetries.listWithParams(params)
      
      // Validate that we got all telemetries (not paginated)
      if (response.items.length < response.total) {
        console.warn(
          `Entity Catalog: Only received ${response.items.length} of ${response.total} telemetries. ` +
          'This may cause inconsistent results. Consider increasing pageSize or implementing proper pagination.'
        )
      }
      
      // Validate that entities are populated (our backend fix should ensure this)
      const telemetriesWithoutEntities = response.items.filter(
        telemetry => !telemetry.entities || Object.keys(telemetry.entities).length === 0
      )
      
      if (telemetriesWithoutEntities.length > 0) {
        console.warn(
          `Found ${telemetriesWithoutEntities.length} telemetries without entities. ` +
          'This may indicate a backend issue or missing entity associations.'
        )
      }
      
      return response
    },
    // Cache for 5 minutes since entity relationships don't change frequently
    staleTime: 5 * 60 * 1000,
    // Keep data in cache for 10 minutes
    gcTime: 10 * 60 * 1000,
  })
}
