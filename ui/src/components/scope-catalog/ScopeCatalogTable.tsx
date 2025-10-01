import { useMemo, useState } from 'react'
import { ChevronDown, ChevronRight, Search } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { useScopeCatalog } from '@/hooks/scope'
import { ScopeCatalogRow } from '@/types/scope-catalog'
import { ScopeTelemetriesExpansion } from './ScopeTelemetriesTable'

interface ScopeCatalogTableProps {
  className?: string
}

// Get icon for scope (using magnifying glass for instrumentation scopes)
const getScopeIcon = (scopeName: string): string => {
  // Use different icons based on scope type/name patterns
  if (scopeName.includes('receiver')) return '📡'
  if (scopeName.includes('instrumentation')) return '🔧'
  if (scopeName.includes('otelhttp')) return '🌐'
  if (scopeName.includes('requests')) return '📨'
  return '🔍' // Default magnifying glass for scopes
}

// Get badge color for scope
const getScopeBadgeColor = (scopeName: string): string => {
  // Use blue theme for all scopes as specified in the project
  return 'bg-blue-50 text-blue-700 border-blue-200 hover:bg-blue-100'
}

// Format scope name for display (shorten long names)
const formatScopeName = (scopeName: string): string => {
  if (scopeName.length <= 50) return scopeName
  
  // Try to shorten common long scope names
  const parts = scopeName.split('/')
  if (parts.length > 3) {
    return `${parts[0]}/.../${parts[parts.length - 1]}`
  }
  
  return scopeName.length > 60 ? `${scopeName.substring(0, 57)}...` : scopeName
}

// Summary component for scope catalog stats
const ScopeCatalogSummary = ({ rows }: { rows: ScopeCatalogRow[] }) => {
  const totalScopes = rows.length
  const totalTelemetries = rows.reduce((sum, row) => sum + row.total, 0)
  const totalMetrics = rows.reduce((sum, row) => sum + row.metrics, 0)
  const totalLogs = rows.reduce((sum, row) => sum + row.logs, 0)
  const totalSpans = rows.reduce((sum, row) => sum + row.spans, 0)
  const totalProfiles = rows.reduce((sum, row) => sum + row.profiles, 0)

  return (
    <div className="grid grid-cols-2 md:grid-cols-6 gap-4 p-4 bg-muted/30 rounded-lg">
      <div className="text-center">
        <div className="text-2xl font-bold text-blue-600">{totalScopes}</div>
        <div className="text-sm text-muted-foreground">Scopes</div>
      </div>
      <div className="text-center">
        <div className="text-2xl font-bold">{totalTelemetries}</div>
        <div className="text-sm text-muted-foreground">Total</div>
      </div>
      <div className="text-center">
        <div className="text-2xl font-bold text-green-600">{totalMetrics}</div>
        <div className="text-sm text-muted-foreground">Metrics</div>
      </div>
      <div className="text-center">
        <div className="text-2xl font-bold text-orange-600">{totalLogs}</div>
        <div className="text-sm text-muted-foreground">Logs</div>
      </div>
      <div className="text-center">
        <div className="text-2xl font-bold text-purple-600">{totalSpans}</div>
        <div className="text-sm text-muted-foreground">Spans</div>
      </div>
      <div className="text-center">
        <div className="text-2xl font-bold text-red-600">{totalProfiles}</div>
        <div className="text-sm text-muted-foreground">Profiles</div>
      </div>
    </div>
  )
}

export function ScopeCatalogTable({ className = '' }: ScopeCatalogTableProps) {
  const { rows, isLoading, error } = useScopeCatalog()

  // Expansion state management
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set())

  // Handle row click to toggle expansion
  const handleRowClick = (scopeName: string) => {
    setExpandedRows(prev => {
      const newSet = new Set(prev)
      if (newSet.has(scopeName)) {
        newSet.delete(scopeName)
      } else {
        newSet.add(scopeName)
      }
      return newSet
    })
  }

  // Loading state
  if (isLoading) {
    return (
      <div className={`space-y-4 ${className}`}>
        <div className="h-24 bg-muted/30 rounded-lg animate-pulse" />
        <div className="space-y-2">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="h-16 bg-muted/20 rounded-lg animate-pulse" />
          ))}
        </div>
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className={`space-y-4 ${className}`}>
        <div className="p-8 text-center border-2 border-dashed border-red-200 rounded-lg">
          <div className="text-red-600 text-lg font-medium mb-2">Error Loading Scope Catalog</div>
          <div className="text-red-500 text-sm">{error}</div>
        </div>
      </div>
    )
  }

  // Empty state
  if (rows.length === 0) {
    return (
      <div className={`space-y-4 ${className}`}>
        <div className="p-8 text-center border-2 border-dashed border-muted-foreground/25 rounded-lg">
          <Search className="mx-auto h-12 w-12 text-muted-foreground/50 mb-4" />
          <div className="text-lg font-medium text-muted-foreground mb-2">No Scopes Found</div>
          <div className="text-sm text-muted-foreground">
            No instrumentation scopes were found in your telemetry data.
          </div>
        </div>
      </div>
    )
  }

  // Render table with data and expansion functionality
  return (
    <div className={`space-y-4 ${className}`}>
      <ScopeCatalogSummary rows={rows} />
      
      <div className="border rounded-lg overflow-hidden">
        <table className="w-full">
          <thead className="bg-muted/50">
            <tr>
              <th className="text-left p-4 font-medium text-muted-foreground">Scope</th>
              <th className="text-left p-4 font-medium text-muted-foreground">Version</th>
              <th className="text-center p-4 font-medium text-muted-foreground">Metrics</th>
              <th className="text-center p-4 font-medium text-muted-foreground">Logs</th>
              <th className="text-center p-4 font-medium text-muted-foreground">Spans</th>
              <th className="text-center p-4 font-medium text-muted-foreground">Profiles</th>
              <th className="text-center p-4 font-medium text-muted-foreground">Total</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => {
              const isExpanded = expandedRows.has(row.scopeName)
              const icon = getScopeIcon(row.scopeName)
              const badgeColor = getScopeBadgeColor(row.scopeName)
              const displayName = formatScopeName(row.scopeName)

              return (
                <>
                  {/* Main scope row */}
                  <tr
                    key={row.scopeName}
                    className="border-t hover:bg-muted/30 cursor-pointer transition-colors"
                    onClick={() => handleRowClick(row.scopeName)}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault()
                        handleRowClick(row.scopeName)
                      }
                    }}
                    aria-expanded={isExpanded}
                    aria-label={`${isExpanded ? 'Collapse' : 'Expand'} ${row.scopeName} attributes`}
                  >
                    <td className="p-4">
                      <div className="flex items-center gap-3">
                        <div className="flex items-center gap-2">
                          {isExpanded ? (
                            <ChevronDown className="h-4 w-4 text-muted-foreground" />
                          ) : (
                            <ChevronRight className="h-4 w-4 text-muted-foreground" />
                          )}
                          <span className="text-lg">{icon}</span>
                        </div>
                        <div>
                          <Badge
                            variant="outline"
                            className={`font-medium ${badgeColor}`}
                            title={row.scopeName} // Show full name on hover
                          >
                            {displayName}
                          </Badge>
                        </div>
                      </div>
                    </td>
                    <td className="p-4">
                      <span className="text-sm font-mono text-muted-foreground">
                        {row.versions}
                      </span>
                    </td>
                    <td className="p-4 text-center">
                      <span className="inline-flex items-center justify-center w-8 h-8 rounded-full bg-green-100 text-green-700 text-sm font-medium">
                        {row.metrics}
                      </span>
                    </td>
                    <td className="p-4 text-center">
                      <span className="inline-flex items-center justify-center w-8 h-8 rounded-full bg-orange-100 text-orange-700 text-sm font-medium">
                        {row.logs}
                      </span>
                    </td>
                    <td className="p-4 text-center">
                      <span className="inline-flex items-center justify-center w-8 h-8 rounded-full bg-purple-100 text-purple-700 text-sm font-medium">
                        {row.spans}
                      </span>
                    </td>
                    <td className="p-4 text-center">
                      <span className="inline-flex items-center justify-center w-8 h-8 rounded-full bg-red-100 text-red-700 text-sm font-medium">
                        {row.profiles}
                      </span>
                    </td>
                    <td className="p-4 text-center">
                      <span className="inline-flex items-center justify-center w-8 h-8 rounded-full bg-blue-100 text-blue-700 text-sm font-medium">
                        {row.total}
                      </span>
                    </td>
                  </tr>

                  {/* Expanded telemetries row */}
                  {isExpanded && (
                    <tr key={`${row.scopeName}-expanded`}>
                      <td colSpan={7} className="px-8 py-4 bg-muted/20 border-t">
                        <ScopeTelemetriesExpansion
                          scopeName={row.scopeName}
                          isExpanded={isExpanded}
                        />
                      </td>
                    </tr>
                  )}
                </>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}
