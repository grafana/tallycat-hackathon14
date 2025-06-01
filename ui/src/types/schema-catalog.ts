import type { Attribute, Status, Telemetry, TelemetryProducer } from './telemetry'

// Extend TelemetryProducer to include additional fields needed for schema catalog
export interface Producer extends TelemetryProducer {
  id: string
  team: string
  environment: string
}

// Extend Telemetry to include schema-specific fields
export interface Schema {
  id: string
  name: string
  version: string | null
  status: Status
  discoveredAt: string
  lastSeen: string
  producers: Producer[]
  resourceAttributes: Attribute[]
  instrumentationAttributes: Attribute[]
  telemetryAttributes: Attribute[]
}

export interface VersionAssignmentViewProps {
  schemaId: string
  schemaData: Telemetry
  onVersionChange: (version: string) => void
}

export interface AssignmentForm {
  version: string
  description: string
}

export interface VersionValidation {
  isValid: boolean
  message: string
}

export interface Filters {
  status: Status[]
  minProducers: string
  maxProducers: string
}

export interface SchemaAssignment {
  schemaId: string
  status: string
  version: string
  producerCount: number
  lastSeen: string
}

export interface ListSchemaAssignmentsResponse {
  items: SchemaAssignment[]
  total: number
  page: number
  pageSize: number
} 