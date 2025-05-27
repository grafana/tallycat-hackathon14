export enum TelemetryType {
  Metric = 'Metric',
  Log = 'Log',
  Trace = 'Trace',
  Event = 'Event'
}

export enum Status {
  Active = 'Active',
  Deprecated = 'Deprecated',
  Experimental = 'Experimental',
  Stable = 'Stable'
}

export type ViewMode = 'grid' | 'list'

export interface TelemetryProducer {
  name: string
  namespace: string
  firstSeen: string
  lastSeen: string
}

export interface Telemetry {
  schemaId: string
  schemaKey: string
  schemaVersion: string
  schemaUrl?: string
  telemetryType: TelemetryType
  metricUnit: string
  metricType: string
  metricTemporality: string
  brief?: string
  note?: string
  protocol: string
  seenCount: number
  createdAt: string
  updatedAt: string
  attributes: Attribute[]
  producers: Record<string, TelemetryProducer>
}

export interface Attribute {
  name: string
  type: string
  source: string
}

