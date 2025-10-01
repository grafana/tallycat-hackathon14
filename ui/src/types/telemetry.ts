export enum TelemetryType {
  Metric = 'Metric',
  Log = 'Log',
  Span = 'Span',
  Profile = 'Profile',
}

export enum Status {
  Active = 'Active',
  Deprecated = 'Deprecated',
  Experimental = 'Experimental',
  Stable = 'Stable',
}

export type ViewMode = 'grid' | 'list'

export interface TelemetryEntity {
  id: string
  type: string
  attributes: Record<string, any>
  firstSeen: string
  lastSeen: string
}

export interface TelemetryScope {
  id: string
  name: string
  version: string
  schemaURL: string
  attributes: Record<string, any>
  firstSeen: string
  lastSeen: string
}

export interface Telemetry {
  schemaId: string
  schemaKey: string
  schemaUrl?: string
  telemetryType: TelemetryType
  // Metric fields
  metricUnit: string
  metricType: string
  metricTemporality: string
  // Log fields
  logSeverityNumber: number
  logSeverityText: string
  logBody: string
  logFlags: number
  logTraceID: string
  logSpanID: string
  logEventName: string
  logDroppedAttributesCount: number
  // Trace fields
  spanKind: string
  spanName: string
  spanID: string
  traceID: string
  // Profile fields
  profileSampleAggregationTemporality: string
  profileSampleUnit: string
  // Common fields
  brief?: string
  note?: string
  protocol: string
  seenCount: number
  createdAt: string
  updatedAt: string
  attributes: Attribute[]
  entities: Record<string, TelemetryEntity>
  scope: TelemetryScope | null
}

export interface Attribute {
  name: string
  type: string
  source: string
  brief?: string
}

export interface TelemetryHistory {
  id: number
  schemaKey: string
  version: string
  timestamp: string
  summary: string
  status: string
  snapshot: string
  createdAt: string
  updatedAt: string
}
