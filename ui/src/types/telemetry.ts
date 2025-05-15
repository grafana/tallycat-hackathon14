export type TelemetryType = 'metric' | 'log' | 'trace' | 'event'
export type DataType = 'counter' | 'gauge' | 'histogram' | 'summary' | 'string' | 'number' | 'boolean' | 'object'
export type Status = 'active' | 'deprecated' | 'experimental' | 'stable'
export type SortField = 'name' | 'lastUpdated' | 'type' | 'dataType'
export type SortDirection = 'asc' | 'desc'
export type ViewMode = 'grid' | 'list'

export interface TelemetryItem {
  id: string
  name: string
  type: TelemetryType
  dataType: DataType
  status: Status
  description: string
  lastUpdated: string
  schemaVersionCount: number
  created: string
  fields: number
  source: string
  instrumentationLibrary: string
  format: string
  unit: string
  aggregation: string
  cardinality: string
  tags: string[]
  sourceTeams: string[]
}

export interface FilterState {
  [key: string]: string[]
}

export interface TelemetryFilters {
  searchQuery: string
  activeFilters: FilterState
  activeFilterCount: number
  sortField: SortField
  sortDirection: SortDirection
  viewMode: ViewMode
  activeTab: string
}

export interface GetSchemaResponse {
  id: string
  name: string
  type: TelemetryType
  dataType: DataType
  status: Status
  description: string
  lastUpdated: string
  schemaVersionCount: number
  created: string
  fields: number
  source: string
  instrumentationLibrary: string
  format: string
  unit: string
  aggregation: string
  cardinality: string
  tags: string[]
  sources: SchemaSource[]
  sourceTeams: string[]
  schema: SchemaField[]
  metricDetails: MetricDetails
  usedBy: SchemaUsage[]
  history: SchemaVersion[]
  examples: SchemaExample[]
  validationRules: ValidationRule[]
}

export interface SchemaSource {
  id: string
  name: string
  team: string
  environment: string
  health: string
  version: string
  volume: number
  dailyAverage: number
  peak: number
  contribution: number
  compliance: string
  requiredFieldsPresent: number
  requiredFieldsTotal: number
  optionalFieldsPresent: number
  optionalFieldsTotal: number
  lastValidated: string
}

export interface SchemaField {
  name: string
  type: string
  description: string
  required: boolean
  source: string
}

export interface MetricDetails {
  type: string
  unit: string
  aggregation: string
  metricName: string
  otelCompatible: boolean
  buckets: number[]
  monotonic: boolean
  instrumentationScope: string
  semanticConventions: string
}

export interface SchemaUsage {
  name: string
  type: string
  id: string
}

export interface SchemaVersion {
  version: string
  date: string
  author: string
  changes: string
  validationStatus: string
}

export interface SchemaExample {
  description: string
  value: Record<string, unknown>
}

export interface ValidationRule {
  field: string
  rule: string
  description: string
} 