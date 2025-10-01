import { Telemetry, TelemetryType } from '@/types/telemetry'
import { ScopeCatalogRow } from '@/types/scope-catalog'

/**
 * Aggregates telemetries by scope name to create scope catalog rows
 */
export const aggregateScopesByName = (telemetries: Telemetry[]): ScopeCatalogRow[] => {
  // Filter out telemetries without scopes
  const telemetriesWithScopes = telemetries.filter(t => t.scope !== null)
  
  if (telemetriesWithScopes.length === 0) {
    return []
  }

  // Group telemetries by scope name
  const scopeGroups = telemetriesWithScopes.reduce((groups, telemetry) => {
    const scopeName = telemetry.scope!.name
    
    if (!groups[scopeName]) {
      groups[scopeName] = []
    }
    groups[scopeName].push(telemetry)
    
    return groups
  }, {} as Record<string, Telemetry[]>)

  // Transform each group into a ScopeCatalogRow
  const rows: ScopeCatalogRow[] = Object.entries(scopeGroups).map(([scopeName, telemetries]) => {
    // Collect unique versions for this scope
    const versions = new Set<string>()
    
    // Count telemetries by type
    let metrics = 0
    let logs = 0
    let spans = 0
    let profiles = 0

    telemetries.forEach(telemetry => {
      // Add version to set (automatically handles uniqueness)
      if (telemetry.scope!.version) {
        versions.add(telemetry.scope!.version)
      }

      // Count by telemetry type
      switch (telemetry.telemetryType) {
        case TelemetryType.Metric:
          metrics++
          break
        case TelemetryType.Log:
          logs++
          break
        case TelemetryType.Span:
          spans++
          break
        case TelemetryType.Profile:
          profiles++
          break
      }
    })

    // Create comma-separated version string, sorted for consistency
    const versionsArray = Array.from(versions).sort()
    const versionsString = versionsArray.length > 0 ? versionsArray.join(', ') : 'UNKNOWN'

    return {
      scopeName,
      versions: versionsString,
      metrics,
      logs,
      spans,
      profiles,
      total: metrics + logs + spans + profiles,
    }
  })

  // Sort by scope name for consistent display
  return rows.sort((a, b) => a.scopeName.localeCompare(b.scopeName))
}

/**
 * Validates that telemetries have the expected structure for scope aggregation
 */
export const validateTelemetriesForScopes = (telemetries: Telemetry[]): {
  valid: boolean
  totalTelemetries: number
  telemetriesWithScopes: number
  telemetriesWithoutScopes: number
  uniqueScopeNames: number
} => {
  const totalTelemetries = telemetries.length
  const telemetriesWithScopes = telemetries.filter(t => t.scope !== null).length
  const telemetriesWithoutScopes = totalTelemetries - telemetriesWithScopes
  
  const uniqueScopeNames = new Set(
    telemetries
      .filter(t => t.scope !== null)
      .map(t => t.scope!.name)
  ).size

  return {
    valid: telemetriesWithScopes > 0,
    totalTelemetries,
    telemetriesWithScopes,
    telemetriesWithoutScopes,
    uniqueScopeNames,
  }
}

/**
 * Gets statistics about scope catalog data
 */
export const getScopeCatalogStats = (rows: ScopeCatalogRow[]) => {
  const totalScopes = rows.length
  const totalTelemetries = rows.reduce((sum, row) => sum + row.total, 0)
  const totalMetrics = rows.reduce((sum, row) => sum + row.metrics, 0)
  const totalLogs = rows.reduce((sum, row) => sum + row.logs, 0)
  const totalSpans = rows.reduce((sum, row) => sum + row.spans, 0)
  const totalProfiles = rows.reduce((sum, row) => sum + row.profiles, 0)

  // Count unique versions across all scopes
  const allVersions = new Set<string>()
  rows.forEach(row => {
    row.versions.split(', ').forEach(version => {
      if (version.trim() && version !== 'UNKNOWN') {
        allVersions.add(version.trim())
      }
    })
  })

  return {
    totalScopes,
    totalTelemetries,
    totalMetrics,
    totalLogs,
    totalSpans,
    totalProfiles,
    uniqueVersions: allVersions.size,
  }
}
