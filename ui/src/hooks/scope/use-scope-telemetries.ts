import { useMemo } from 'react'
import { useScopeTelemetries } from './use-scopes'
import { Telemetry, TelemetryType } from '@/types/telemetry'

export interface ScopeTelemetryItem {
  schemaKey: string
  telemetryType: TelemetryType
}

export interface ScopeTelemetriesData {
  scopeName: string
  telemetries: ScopeTelemetryItem[]
  totalCount: number
}

/**
 * Hook to get telemetries for a specific scope name
 */
export const useScopeTelemetriesData = (scopeName: string): ScopeTelemetriesData => {
  const { data: scopesResponse } = useScopeTelemetries()

  const scopeTelemetries = useMemo(() => {
    if (!scopesResponse?.items) {
      return {
        scopeName,
        telemetries: [],
        totalCount: 0,
      }
    }

    // Filter telemetries that belong to this scope
    const relevantTelemetries = scopesResponse.items.filter(
      telemetry => telemetry.scope !== null && telemetry.scope.name === scopeName
    )

    // Transform to simple objects with just schema key and type
    const telemetryItems: ScopeTelemetryItem[] = relevantTelemetries.map(telemetry => ({
      schemaKey: telemetry.schemaKey,
      telemetryType: telemetry.telemetryType,
    }))

    // Sort by schema key for consistent display
    telemetryItems.sort((a, b) => a.schemaKey.localeCompare(b.schemaKey))

    return {
      scopeName,
      telemetries: telemetryItems,
      totalCount: telemetryItems.length,
    }
  }, [scopesResponse, scopeName])

  return scopeTelemetries
}
