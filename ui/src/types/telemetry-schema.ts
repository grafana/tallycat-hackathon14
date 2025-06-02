import type { ListResponse } from '@/lib/api-client'
import type {
  Attribute,
  Status,
  Telemetry,
  TelemetryProducer,
} from './telemetry'

// Extend Telemetry to include schema-specific fields
export interface TelemetrySchema {
  id: string
  version: string | null
  status: Status
  lastSeen: string
  producers: TelemetryProducer[]
  attributes: Attribute[]
}

export interface VersionAssignmentViewProps {
  telemetry: Telemetry
}

export interface TelemetrySchemaGrid {
  schemaId: string
  status: Status
  version: string | null
  producerCount: number
  lastSeen: string
}

export interface ListTelemetrySchemasResponse
  extends ListResponse<TelemetrySchemaGrid> {}
