// Entity Aggregation Utilities
// Functions to aggregate telemetry data by entity type for the Entity Catalog

import type { Telemetry } from '@/types/telemetry'
import type { EntityCatalogRow } from '@/types/entity-catalog'

/**
 * Aggregates telemetry data by entity type to create Entity Catalog rows
 * 
 * @param telemetries Array of telemetry objects with populated entities
 * @returns Array of EntityCatalogRow objects, one per unique entity type
 */
export function aggregateEntitiesByType(telemetries: Telemetry[]): EntityCatalogRow[] {
  // Create a map to track counts by entity type
  const entityTypeMap = new Map<string, {
    metrics: number
    logs: number
    spans: number
    profiles: number
  }>()

  // Process each telemetry and its associated entities
  for (const telemetry of telemetries) {
    if (!telemetry.entities) continue

    // Get unique entity types for this telemetry
    const entityTypes = new Set<string>()
    for (const entity of Object.values(telemetry.entities)) {
      entityTypes.add(entity.type)
    }

    // Count this telemetry for each associated entity type
    for (const entityType of entityTypes) {
      if (!entityTypeMap.has(entityType)) {
        entityTypeMap.set(entityType, {
          metrics: 0,
          logs: 0,
          spans: 0,
          profiles: 0,
        })
      }

      const counts = entityTypeMap.get(entityType)!
      
      // Increment the appropriate counter based on telemetry type
      switch (telemetry.telemetryType) {
        case 'Metric':
          counts.metrics++
          break
        case 'Log':
          counts.logs++
          break
        case 'Span':
          counts.spans++
          break
        case 'Profile':
          counts.profiles++
          break
      }
    }
  }

  // Convert map to array of EntityCatalogRow objects
  const rows: EntityCatalogRow[] = []
  for (const [entityType, counts] of entityTypeMap.entries()) {
    const total = counts.metrics + counts.logs + counts.spans + counts.profiles
    rows.push({
      entityType,
      metrics: counts.metrics,
      logs: counts.logs,
      spans: counts.spans,
      profiles: counts.profiles,
      total,
    })
  }

  // Sort by entity type for consistent ordering
  return rows.sort((a, b) => a.entityType.localeCompare(b.entityType))
}

/**
 * Validates that telemetry objects have the required entities field populated
 * 
 * @param telemetries Array of telemetry objects
 * @returns true if all telemetries have entities populated, false otherwise
 */
export function validateTelemetryEntities(telemetries: Telemetry[]): boolean {
  return telemetries.every(telemetry => 
    telemetry.entities !== null && 
    telemetry.entities !== undefined
  )
}

/**
 * Gets summary statistics for the Entity Catalog
 * 
 * @param rows Array of EntityCatalogRow objects
 * @returns Summary statistics object
 */
export function getEntityCatalogSummary(rows: EntityCatalogRow[]) {
  const totalEntityTypes = rows.length
  const totalTelemetries = rows.reduce((sum, row) => sum + row.total, 0)
  const totalMetrics = rows.reduce((sum, row) => sum + row.metrics, 0)
  const totalLogs = rows.reduce((sum, row) => sum + row.logs, 0)
  const totalSpans = rows.reduce((sum, row) => sum + row.spans, 0)
  const totalProfiles = rows.reduce((sum, row) => sum + row.profiles, 0)

  return {
    totalEntityTypes,
    totalTelemetries,
    totalMetrics,
    totalLogs,
    totalSpans,
    totalProfiles,
  }
}
