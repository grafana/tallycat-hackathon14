export type TelemetryType = 'metric' | 'log' | 'trace'
export type DataType = 'gauge' | 'counter' | 'histogram' | 'structured' | 'unstructured' | 'span'
export type Status = 'active' | 'draft' | 'deprecated'
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
  fields: number
  source: string
  instrumentationLibrary: string
  format: string
  unit?: string
  aggregation?: string
  cardinality: string
  tags: string[]
  severity?: string
  spanKind?: string
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