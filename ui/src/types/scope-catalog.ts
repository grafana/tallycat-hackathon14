export interface ScopeCatalogRow {
  scopeName: string
  versions: string // Comma-separated versions if multiple
  metrics: number
  logs: number
  spans: number
  profiles: number
  total: number
}

export interface ScopeCatalogData {
  rows: ScopeCatalogRow[]
  isLoading: boolean
  error: string | null
}

// Re-export types we need from telemetry
export type { Telemetry, TelemetryScope } from '@/types/telemetry'
