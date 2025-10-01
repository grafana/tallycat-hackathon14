import { useMemo } from 'react'
import { useScopeTelemetries } from './use-scopes'
import { aggregateScopesByName, getScopeCatalogStats } from '@/utils/scope-aggregation'
import { ScopeCatalogData } from '@/types/scope-catalog'

/**
 * Hook that combines scope data fetching with aggregation logic
 * Returns processed scope catalog data ready for display
 */
export const useScopeCatalog = (): ScopeCatalogData => {
  const { data: scopesResponse, isLoading, error } = useScopeTelemetries()

  const processedData = useMemo(() => {
    if (!scopesResponse?.items) {
      return {
        rows: [],
        isLoading,
        error: error?.message || null,
      }
    }

    try {
      // Aggregate telemetries by scope name
      const rows = aggregateScopesByName(scopesResponse.items)
      
      // Log statistics for debugging
      if (rows.length > 0) {
        const stats = getScopeCatalogStats(rows)
        console.info('Scope Catalog Stats:', stats)
      }

      return {
        rows,
        isLoading,
        error: null,
      }
    } catch (err) {
      console.error('Error processing scope catalog data:', err)
      return {
        rows: [],
        isLoading,
        error: err instanceof Error ? err.message : 'Failed to process scope data',
      }
    }
  }, [scopesResponse, isLoading, error])

  return processedData
}
