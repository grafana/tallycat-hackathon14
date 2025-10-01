// Entity Catalog Types
// Types for the Entity Catalog feature that aggregates telemetry data by entity type

export interface EntityCatalogRow {
  /** The entity type (e.g., "host", "container", "service") */
  entityType: string
  /** Count of metric schemas associated with this entity type */
  metrics: number
  /** Count of log schemas associated with this entity type */
  logs: number
  /** Count of span schemas associated with this entity type */
  spans: number
  /** Count of profile schemas associated with this entity type */
  profiles: number
  /** Total count of all telemetry schemas for this entity type */
  total: number
}

export interface EntityCatalogData {
  /** Array of entity catalog rows, one per unique entity type */
  rows: EntityCatalogRow[]
  /** Total number of unique entity types */
  totalEntityTypes: number
  /** Loading state */
  isLoading: boolean
  /** Error state */
  error: string | null
}

// Re-export existing types that we'll use
export type { Telemetry, TelemetryEntity } from '@/types/telemetry'
