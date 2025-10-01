'use client'

import { useMemo, useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { DataTable } from '@/components/ui/data-table'
import { ChevronRight, ChevronDown } from 'lucide-react'
import type { EntityCatalogRow } from '@/types/entity-catalog'
import { useEntityCatalog } from '@/hooks'
import { EntityAttributesExpansion } from './EntityAttributesTable'

// Entity type icon mapping
const getEntityTypeIcon = (entityType: string) => {
  const iconMap: Record<string, string> = {
    host: '🖥️',
    container: '📦',
    service: '⚙️',
    process: '🔄',
    os: '💻',
    telemetry: '📊',
  }
  return iconMap[entityType] || '📋'
}

// Entity type color mapping
const getEntityTypeBadgeColor = (entityType: string) => {
  const colorMap: Record<string, string> = {
    host: 'bg-blue-100 text-blue-800 border-blue-200',
    container: 'bg-purple-100 text-purple-800 border-purple-200',
    service: 'bg-green-100 text-green-800 border-green-200',
    process: 'bg-orange-100 text-orange-800 border-orange-200',
    os: 'bg-gray-100 text-gray-800 border-gray-200',
    telemetry: 'bg-indigo-100 text-indigo-800 border-indigo-200',
  }
  return colorMap[entityType] || 'bg-slate-100 text-slate-800 border-slate-200'
}

// Column definitions for Entity Catalog table
const createEntityCatalogColumns = (
  expandedRows: Set<string>,
  onRowClick: (entityType: string) => void
): ColumnDef<EntityCatalogRow>[] => [
  {
    accessorKey: 'entityType',
    header: 'Entity',
    cell: ({ row }) => {
      const entityType = row.getValue('entityType') as string
      const icon = getEntityTypeIcon(entityType)
      const badgeColor = getEntityTypeBadgeColor(entityType)
      const isExpanded = expandedRows.has(entityType)
      
      return (
        <div 
          className="flex items-center gap-3 cursor-pointer hover:bg-muted/50 rounded-md p-2 -m-2 transition-colors"
          onClick={() => onRowClick(entityType)}
          role="button"
          tabIndex={0}
          onKeyDown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault()
              onRowClick(entityType)
            }
          }}
          aria-expanded={isExpanded}
          aria-label={`${isExpanded ? 'Collapse' : 'Expand'} ${entityType} attributes`}
        >
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
              className={`capitalize font-medium ${badgeColor}`}
            >
              {entityType}
            </Badge>
          </div>
        </div>
      )
    },
  },
  {
    accessorKey: 'metrics',
    header: 'Metrics',
    cell: ({ row }) => {
      const count = row.getValue('metrics') as number
      return (
        <div className="text-center">
          <span className="text-sm font-medium">{count.toLocaleString()}</span>
        </div>
      )
    },
  },
  {
    accessorKey: 'logs',
    header: 'Logs',
    cell: ({ row }) => {
      const count = row.getValue('logs') as number
      return (
        <div className="text-center">
          <span className="text-sm font-medium">{count.toLocaleString()}</span>
        </div>
      )
    },
  },
  {
    accessorKey: 'spans',
    header: 'Spans',
    cell: ({ row }) => {
      const count = row.getValue('spans') as number
      return (
        <div className="text-center">
          <span className="text-sm font-medium">{count.toLocaleString()}</span>
        </div>
      )
    },
  },
  {
    accessorKey: 'profiles',
    header: 'Profiles',
    cell: ({ row }) => {
      const count = row.getValue('profiles') as number
      return (
        <div className="text-center">
          <span className="text-sm font-medium">{count.toLocaleString()}</span>
        </div>
      )
    },
  },
  {
    accessorKey: 'total',
    header: 'Total',
    cell: ({ row }) => {
      const count = row.getValue('total') as number
      return (
        <div className="text-center">
          <Badge variant="secondary" className="font-medium">
            {count.toLocaleString()}
          </Badge>
        </div>
      )
    },
  },
]

// Loading state component
const LoadingState = () => (
  <div className="flex items-center justify-center py-12">
    <div className="flex items-center gap-2">
      <div className="h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      <span className="text-sm text-muted-foreground">Loading entity data...</span>
    </div>
  </div>
)

// Error state component
const ErrorState = ({ error }: { error: string }) => (
  <div className="flex flex-col items-center justify-center py-12">
    <div className="text-center">
      <div className="text-2xl mb-2">⚠️</div>
      <h3 className="text-lg font-medium text-foreground mb-2">Error Loading Entity Data</h3>
      <p className="text-sm text-muted-foreground mb-4 max-w-md">{error}</p>
      <button 
        onClick={() => window.location.reload()} 
        className="text-sm text-primary hover:underline"
      >
        Try again
      </button>
    </div>
  </div>
)

// Empty state component
const EmptyState = () => (
  <div className="flex flex-col items-center justify-center py-12">
    <div className="text-center">
      <div className="text-4xl mb-4">📋</div>
      <h3 className="text-lg font-medium text-foreground mb-2">No Entity Data Found</h3>
      <p className="text-sm text-muted-foreground max-w-md">
        No entities are currently available in the system. This could mean no telemetry data has been ingested yet.
      </p>
    </div>
  </div>
)

// Summary component
const EntityCatalogSummary = ({ rows }: { rows: EntityCatalogRow[] }) => {
  const summary = useMemo(() => {
    const totalEntityTypes = rows.length
    const totalTelemetries = rows.reduce((sum, row) => sum + row.total, 0)
    const totalMetrics = rows.reduce((sum, row) => sum + row.metrics, 0)
    const totalLogs = rows.reduce((sum, row) => sum + row.logs, 0)
    const totalSpans = rows.reduce((sum, row) => sum + row.spans, 0)
    const totalProfiles = rows.reduce((sum, row) => sum + row.profiles, 0)

    return {
      totalEntityTypes,
      totalTelemetries,
      totalMetrics,
      totalLogs,
      totalSpans,
      totalProfiles,
    }
  }, [rows])

  if (rows.length === 0) return null

  return (
    <div className="text-sm text-muted-foreground">
      Showing {summary.totalEntityTypes} entity types with {summary.totalTelemetries.toLocaleString()} total telemetry schemas
      ({summary.totalMetrics.toLocaleString()} metrics, {summary.totalLogs.toLocaleString()} logs, {summary.totalSpans.toLocaleString()} spans, {summary.totalProfiles.toLocaleString()} profiles)
    </div>
  )
}

// Main EntityCatalogTable component
interface EntityCatalogTableProps {
  className?: string
}

export function EntityCatalogTable({ className = '' }: EntityCatalogTableProps) {
  const { rows, isLoading, error } = useEntityCatalog()
  
  // Expansion state management
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set())
  
  // Handle row click to toggle expansion
  const handleRowClick = (entityType: string) => {
    setExpandedRows(prev => {
      const newSet = new Set(prev)
      if (newSet.has(entityType)) {
        newSet.delete(entityType)
      } else {
        newSet.add(entityType)
      }
      return newSet
    })
  }
  
  const columns = useMemo(() => 
    createEntityCatalogColumns(expandedRows, handleRowClick), 
    [expandedRows]
  )

  // Handle loading state
  if (isLoading) {
    return (
      <div className={className}>
        <LoadingState />
      </div>
    )
  }

  // Handle error state
  if (error) {
    return (
      <div className={className}>
        <ErrorState error={error} />
      </div>
    )
  }

  // Handle empty state
  if (rows.length === 0) {
    return (
      <div className={className}>
        <EmptyState />
      </div>
    )
  }

  // Render table with data and expansion functionality
  return (
    <div className={`space-y-4 ${className}`}>
      <EntityCatalogSummary rows={rows} />
      
      {/* Custom table structure for proper alignment */}
      <div className="border rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            {/* Table Header */}
            <thead className="bg-muted/50">
              <tr className="border-b">
                <th className="text-left p-4 font-medium">Entity</th>
                <th className="text-center p-4 font-medium">Metrics</th>
                <th className="text-center p-4 font-medium">Logs</th>
                <th className="text-center p-4 font-medium">Spans</th>
                <th className="text-center p-4 font-medium">Profiles</th>
                <th className="text-center p-4 font-medium">Total</th>
              </tr>
            </thead>
            
            {/* Table Body */}
            <tbody>
              {rows.map((row) => (
                <>
                  {/* Main entity row */}
                  <tr 
                    key={row.entityType}
                    className="border-b hover:bg-muted/30 transition-colors"
                  >
                    <td className="p-4">
                      <div 
                        className="flex items-center gap-3 cursor-pointer hover:bg-muted/50 rounded-md p-2 -m-2 transition-colors"
                        onClick={() => handleRowClick(row.entityType)}
                        role="button"
                        tabIndex={0}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter' || e.key === ' ') {
                            e.preventDefault()
                            handleRowClick(row.entityType)
                          }
                        }}
                        aria-expanded={expandedRows.has(row.entityType)}
                        aria-label={`${expandedRows.has(row.entityType) ? 'Collapse' : 'Expand'} ${row.entityType} attributes`}
                      >
                        <div className="flex items-center gap-2">
                          {expandedRows.has(row.entityType) ? (
                            <ChevronDown className="h-4 w-4 text-muted-foreground" />
                          ) : (
                            <ChevronRight className="h-4 w-4 text-muted-foreground" />
                          )}
                          <span className="text-lg">{getEntityTypeIcon(row.entityType)}</span>
                        </div>
                        <div>
                          <Badge 
                            variant="outline" 
                            className={`capitalize font-medium ${getEntityTypeBadgeColor(row.entityType)}`}
                          >
                            {row.entityType}
                          </Badge>
                        </div>
                      </div>
                    </td>
                    <td className="p-4 text-center">
                      <span className="text-sm font-medium">{row.metrics.toLocaleString()}</span>
                    </td>
                    <td className="p-4 text-center">
                      <span className="text-sm font-medium">{row.logs.toLocaleString()}</span>
                    </td>
                    <td className="p-4 text-center">
                      <span className="text-sm font-medium">{row.spans.toLocaleString()}</span>
                    </td>
                    <td className="p-4 text-center">
                      <span className="text-sm font-medium">{row.profiles.toLocaleString()}</span>
                    </td>
                    <td className="p-4 text-center">
                      <Badge variant="secondary" className="font-medium">
                        {row.total.toLocaleString()}
                      </Badge>
                    </td>
                  </tr>
                  
                  {/* Expanded attributes row */}
                  {expandedRows.has(row.entityType) && (
                    <tr key={`${row.entityType}-expanded`}>
                      <td colSpan={6} className="p-0">
                        <div className="px-8 py-4 bg-muted/20">
                          <EntityAttributesExpansion
                            entityType={row.entityType}
                            isExpanded={expandedRows.has(row.entityType)}
                          />
                        </div>
                      </td>
                    </tr>
                  )}
                </>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
