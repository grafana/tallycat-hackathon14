import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useScopeTelemetries } from './use-scopes'
import { 
  createScopeAttributeData, 
  aggregateAllScopeAttributes,
  validateTelemetriesForScopeAttributes,
  getScopeAttributeStats 
} from '@/utils/scope-attribute-aggregation'
import { ScopeAttributesState, ScopeAttributeData } from '@/types/scope-attributes'

/**
 * Hook to fetch and aggregate attributes for a specific scope name
 */
export const useScopeAttributes = (scopeName: string): ScopeAttributesState => {
  const { data: scopesResponse, isLoading: scopesLoading, error: scopesError } = useScopeTelemetries()

  const processedData = useMemo(() => {
    if (scopesLoading) {
      return {
        data: null,
        isLoading: true,
        error: null,
      }
    }

    if (scopesError) {
      return {
        data: null,
        isLoading: false,
        error: scopesError.message,
      }
    }

    if (!scopesResponse?.items) {
      return {
        data: null,
        isLoading: false,
        error: 'No telemetry data available',
      }
    }

    try {
      // Validate data for this specific scope
      const validation = validateTelemetriesForScopeAttributes(scopesResponse.items, scopeName)
      
      if (!validation.valid) {
        console.warn(
          `Scope Attributes: No telemetries found for scope "${scopeName}". ` +
          `Total telemetries: ${validation.totalTelemetries}, Relevant: ${validation.relevantTelemetries}`
        )
        return {
          data: {
            scopeName,
            attributes: [],
            totalUniqueValues: 0,
          },
          isLoading: false,
          error: null,
        }
      }

      // Create attribute data
      const attributeData = createScopeAttributeData(scopesResponse.items, scopeName)
      
      // Log statistics for debugging
      const stats = getScopeAttributeStats(attributeData)
      console.info(`Scope Attributes for "${scopeName}":`, stats)

      return {
        data: attributeData,
        isLoading: false,
        error: null,
      }
    } catch (err) {
      console.error(`Error processing scope attributes for "${scopeName}":`, err)
      return {
        data: null,
        isLoading: false,
        error: err instanceof Error ? err.message : 'Failed to process scope attributes',
      }
    }
  }, [scopesResponse, scopesLoading, scopesError, scopeName])

  return processedData
}

/**
 * Hook to fetch and aggregate attributes for all scopes
 * Useful for preloading or bulk operations
 */
export const useAllScopeAttributes = () => {
  const { data: scopesResponse, isLoading, error } = useScopeTelemetries()

  return useQuery({
    queryKey: ['scope-attributes', 'all'],
    queryFn: async () => {
      if (!scopesResponse?.items) {
        throw new Error('No telemetry data available')
      }

      const allAttributes = aggregateAllScopeAttributes(scopesResponse.items)
      
      // Log statistics
      const scopeNames = Object.keys(allAttributes)
      const totalAttributes = Object.values(allAttributes).reduce(
        (sum, data) => sum + data.attributes.length, 
        0
      )
      
      console.info(
        `All Scope Attributes: ${scopeNames.length} scopes, ${totalAttributes} total attributes`
      )

      return allAttributes
    },
    enabled: !!scopesResponse?.items && !isLoading && !error,
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
    refetchOnWindowFocus: false,
  })
}
