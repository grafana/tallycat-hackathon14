import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api-client'
import { validateTelemetriesForScopes } from '@/utils/scope-aggregation'

interface ListTelemetriesParams {
  page: number
  pageSize: number
}

/**
 * Hook to fetch all telemetries with scope data for scope catalog
 * Uses a large page size to get all telemetries in one request to avoid pagination issues
 */
export const useScopeTelemetries = () => {
  return useQuery({
    queryKey: ['scopes', 'all-telemetries'],
    queryFn: async () => {
      const params: ListTelemetriesParams = {
        page: 1,
        pageSize: 10000000, // Large enough to get all telemetries in one request
      }
      const response = await api.telemetries.listWithParams(params)

      // Validate that we got all telemetries (not paginated)
      if (response.items.length < response.total) {
        console.warn(
          `Scope Catalog: Only received ${response.items.length} of ${response.total} telemetries. ` +
          'This may cause inconsistent results. Consider increasing pageSize or implementing proper pagination.'
        )
      }

      // Validate telemetries for scope processing
      const validation = validateTelemetriesForScopes(response.items)
      
      if (!validation.valid) {
        console.warn(
          `Scope Catalog: No telemetries with scopes found. ` +
          `Total: ${validation.totalTelemetries}, With scopes: ${validation.telemetriesWithScopes}`
        )
      } else {
        console.info(
          `Scope Catalog: Successfully loaded ${validation.telemetriesWithScopes} telemetries with scopes ` +
          `(${validation.uniqueScopeNames} unique scope names) out of ${validation.totalTelemetries} total telemetries.`
        )
      }

      return response
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
    refetchOnWindowFocus: false,
    retry: 3,
  })
}
