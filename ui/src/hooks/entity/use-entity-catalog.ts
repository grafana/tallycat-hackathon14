import { useMemo } from 'react'
import { useEntities } from './use-entities'
import { aggregateEntitiesByType, validateTelemetryEntities } from '@/utils/entity-aggregation'
import type { EntityCatalogData } from '@/types/entity-catalog'

/**
 * Hook that provides Entity Catalog data by fetching telemetries and aggregating by entity type
 * 
 * @returns EntityCatalogData with aggregated rows, loading state, and error handling
 */
export const useEntityCatalog = (): EntityCatalogData => {
  const { 
    data: telemetriesResponse, 
    isLoading, 
    error,
    isError 
  } = useEntities()

  const entityCatalogData = useMemo((): EntityCatalogData => {
    // Handle loading state
    if (isLoading) {
      return {
        rows: [],
        totalEntityTypes: 0,
        isLoading: true,
        error: null,
      }
    }

    // Handle error state
    if (isError || !telemetriesResponse) {
      return {
        rows: [],
        totalEntityTypes: 0,
        isLoading: false,
        error: error?.message || 'Failed to fetch telemetry data',
      }
    }

    const telemetries = telemetriesResponse.items

    // Validate that telemetries have entities populated
    if (!validateTelemetryEntities(telemetries)) {
      return {
        rows: [],
        totalEntityTypes: 0,
        isLoading: false,
        error: 'Telemetry data is missing entity associations. Please check the backend configuration.',
      }
    }

    // Aggregate telemetries by entity type
    try {
      const rows = aggregateEntitiesByType(telemetries)
      
      return {
        rows,
        totalEntityTypes: rows.length,
        isLoading: false,
        error: null,
      }
    } catch (aggregationError) {
      console.error('Error aggregating entity data:', aggregationError)
      return {
        rows: [],
        totalEntityTypes: 0,
        isLoading: false,
        error: 'Failed to process entity data. Please try again.',
      }
    }
  }, [telemetriesResponse, isLoading, isError, error])

  return entityCatalogData
}

/**
 * Hook that provides additional Entity Catalog statistics and utilities
 * 
 * @returns Extended data including summary statistics
 */
export const useEntityCatalogWithStats = () => {
  const entityCatalog = useEntityCatalog()
  
  const stats = useMemo(() => {
    if (entityCatalog.isLoading || entityCatalog.error || entityCatalog.rows.length === 0) {
      return {
        totalTelemetries: 0,
        totalMetrics: 0,
        totalLogs: 0,
        totalSpans: 0,
        totalProfiles: 0,
        averageTelemetriesPerEntity: 0,
      }
    }

    const totalTelemetries = entityCatalog.rows.reduce((sum, row) => sum + row.total, 0)
    const totalMetrics = entityCatalog.rows.reduce((sum, row) => sum + row.metrics, 0)
    const totalLogs = entityCatalog.rows.reduce((sum, row) => sum + row.logs, 0)
    const totalSpans = entityCatalog.rows.reduce((sum, row) => sum + row.spans, 0)
    const totalProfiles = entityCatalog.rows.reduce((sum, row) => sum + row.profiles, 0)
    const averageTelemetriesPerEntity = entityCatalog.totalEntityTypes > 0 
      ? Math.round(totalTelemetries / entityCatalog.totalEntityTypes * 10) / 10
      : 0

    return {
      totalTelemetries,
      totalMetrics,
      totalLogs,
      totalSpans,
      totalProfiles,
      averageTelemetriesPerEntity,
    }
  }, [entityCatalog])

  return {
    ...entityCatalog,
    stats,
  }
}
