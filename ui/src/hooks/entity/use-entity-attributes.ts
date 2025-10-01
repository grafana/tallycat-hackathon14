import { useMemo } from 'react'
import { useEntities } from './use-entities'
import { 
  aggregateEntityAttributes, 
  createEntityAttributeData,
  validateTelemetriesForAttributes 
} from '@/utils/entity-attribute-aggregation'
import type { EntityAttributeData } from '@/types/entity-attributes'

/**
 * Hook that provides entity attributes for a specific entity type
 * 
 * @param entityType The entity type to get attributes for (e.g., "host", "container")
 * @returns EntityAttributeData with loading and error states
 */
export const useEntityAttributes = (entityType: string) => {
  const { 
    data: telemetriesResponse, 
    isLoading, 
    error,
    isError 
  } = useEntities()

  const entityAttributeData = useMemo(() => {
    // Handle loading state
    if (isLoading) {
      return {
        data: null,
        isLoading: true,
        error: null,
        stats: null
      }
    }

    // Handle error state
    if (isError || !telemetriesResponse) {
      return {
        data: null,
        isLoading: false,
        error: error?.message || 'Failed to fetch telemetry data',
        stats: null
      }
    }

    const telemetries = telemetriesResponse.items

    // Validate telemetry data structure
    const validation = validateTelemetriesForAttributes(telemetries)
    if (!validation.isValid) {
      console.warn('Entity attributes validation issues:', validation.issues)
    }

    try {
      // Create attribute data for the specific entity type
      const attributeData = createEntityAttributeData(telemetries, entityType)
      
      // Calculate stats
      const stats = {
        totalAttributes: attributeData.attributes.length,
        attributesWithMultipleValues: attributeData.attributes.filter(
          attr => attr.values.includes(',')
        ).length,
        validationStats: validation.stats
      }

      return {
        data: attributeData,
        isLoading: false,
        error: null,
        stats
      }
    } catch (aggregationError) {
      console.error('Error aggregating entity attributes:', aggregationError)
      return {
        data: null,
        isLoading: false,
        error: 'Failed to process entity attribute data. Please try again.',
        stats: null
      }
    }
  }, [telemetriesResponse, isLoading, isError, error, entityType])

  return entityAttributeData
}

/**
 * Hook that provides a map of all entity types and their attributes
 * Useful for pre-loading all attribute data
 * 
 * @returns Map of entity type to EntityAttributeData with loading and error states
 */
export const useAllEntityAttributes = () => {
  const { 
    data: telemetriesResponse, 
    isLoading, 
    error,
    isError 
  } = useEntities()

  const allEntityAttributeData = useMemo(() => {
    // Handle loading state
    if (isLoading) {
      return {
        data: new Map<string, EntityAttributeData>(),
        isLoading: true,
        error: null,
        entityTypes: [],
        stats: null
      }
    }

    // Handle error state
    if (isError || !telemetriesResponse) {
      return {
        data: new Map<string, EntityAttributeData>(),
        isLoading: false,
        error: error?.message || 'Failed to fetch telemetry data',
        entityTypes: [],
        stats: null
      }
    }

    const telemetries = telemetriesResponse.items

    // Validate telemetry data structure
    const validation = validateTelemetriesForAttributes(telemetries)
    if (!validation.isValid) {
      console.warn('Entity attributes validation issues:', validation.issues)
    }

    try {
      // Get all unique entity types
      const entityTypes = new Set<string>()
      for (const telemetry of telemetries) {
        if (!telemetry.entities) continue
        for (const entity of Object.values(telemetry.entities)) {
          entityTypes.add(entity.type)
        }
      }

      // Create attribute data for each entity type
      const attributeDataMap = new Map<string, EntityAttributeData>()
      const entityTypeArray = Array.from(entityTypes).sort()

      for (const entityType of entityTypeArray) {
        const attributeData = createEntityAttributeData(telemetries, entityType)
        attributeDataMap.set(entityType, attributeData)
      }

      // Calculate overall stats
      const totalAttributes = Array.from(attributeDataMap.values())
        .reduce((sum, data) => sum + data.attributes.length, 0)

      const stats = {
        totalEntityTypes: entityTypeArray.length,
        totalAttributes,
        validationStats: validation.stats
      }

      return {
        data: attributeDataMap,
        isLoading: false,
        error: null,
        entityTypes: entityTypeArray,
        stats
      }
    } catch (aggregationError) {
      console.error('Error aggregating all entity attributes:', aggregationError)
      return {
        data: new Map<string, EntityAttributeData>(),
        isLoading: false,
        error: 'Failed to process entity attribute data. Please try again.',
        entityTypes: [],
        stats: null
      }
    }
  }, [telemetriesResponse, isLoading, isError, error])

  return allEntityAttributeData
}

/**
 * Hook that provides entity attributes with caching per entity type
 * More efficient for interactive use where users expand different rows
 * 
 * @returns Object with methods to get attributes for specific entity types
 */
export const useEntityAttributesCache = () => {
  const allData = useAllEntityAttributes()
  
  const getEntityAttributes = (entityType: string): EntityAttributeData | null => {
    return allData.data.get(entityType) || null
  }

  const hasEntityType = (entityType: string): boolean => {
    return allData.data.has(entityType)
  }

  const getEntityTypes = (): string[] => {
    return allData.entityTypes
  }

  return {
    ...allData,
    getEntityAttributes,
    hasEntityType,
    getEntityTypes
  }
}
