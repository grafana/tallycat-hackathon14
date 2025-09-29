import type { ListResponse } from '@/lib/api-client'
import type {
  Attribute,
  Status,
  Telemetry,
  TelemetryEntity,
} from './telemetry'

// Extend Telemetry to include schema-specific fields
export interface TelemetrySchema {
  id: string
  version: string | null
  status: Status
  lastSeen: string
  entities: Record<string, TelemetryEntity>
  attributes: Attribute[]
}

export interface VersionAssignmentViewProps {
  telemetry: Telemetry
}

export interface TelemetrySchemaGrid {
  schemaId: string
  status: Status
  version: string | null
  entityCount: number
  lastSeen: string
}

export interface ListTelemetrySchemasResponse
  extends ListResponse<TelemetrySchemaGrid> {}
