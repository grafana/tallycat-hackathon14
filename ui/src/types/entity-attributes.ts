// Entity Attributes Types
// Types for the Interactive Entity Catalog feature that shows detailed attribute information

export interface EntityAttribute {
  /** The attribute name (e.g., "host.name", "container.id") */
  name: string
  /** Comma-separated string of unique values for this attribute */
  values: string
}

export interface EntityAttributeData {
  /** The entity type (e.g., "host", "container", "service") */
  entityType: string
  /** Array of aggregated attributes for this entity type */
  attributes: EntityAttribute[]
}

export interface EntityAttributesState {
  /** Map of entity type to its attribute data */
  data: Map<string, EntityAttributeData>
  /** Loading states per entity type */
  loading: Map<string, boolean>
  /** Error states per entity type */
  errors: Map<string, string | null>
}

// Re-export existing types that we'll use
export type { Telemetry, TelemetryEntity } from '@/types/telemetry'
