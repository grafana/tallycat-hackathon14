'use client'

import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { DataTable } from '@/components/ui/data-table'
import type { EntityAttribute, EntityAttributeData } from '@/types/entity-attributes'
import { useEntityAttributes } from '@/hooks'

// Loading state component for attribute expansion
const AttributesLoadingState = () => (
  <div className="flex items-center justify-center py-8">
    <div className="flex items-center gap-2">
      <div className="h-3 w-3 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      <span className="text-xs text-muted-foreground">Loading attributes...</span>
    </div>
  </div>
)

// Error state component for attribute expansion
const AttributesErrorState = ({ error }: { error: string }) => (
  <div className="flex flex-col items-center justify-center py-8">
    <div className="text-center">
      <div className="text-lg mb-1">⚠️</div>
      <p className="text-xs text-muted-foreground mb-2">Failed to load attributes</p>
      <p className="text-xs text-muted-foreground">{error}</p>
    </div>
  </div>
)

// Empty state component for when no attributes are found
const AttributesEmptyState = ({ entityType }: { entityType: string }) => (
  <div className="flex flex-col items-center justify-center py-8">
    <div className="text-center">
      <div className="text-lg mb-1">📋</div>
      <p className="text-xs text-muted-foreground mb-1">No attributes found</p>
      <p className="text-xs text-muted-foreground">
        No attributes are available for {entityType} entities
      </p>
    </div>
  </div>
)

// Column definitions for the attributes table
const createAttributesColumns = (): ColumnDef<EntityAttribute>[] => [
  {
    accessorKey: 'name',
    header: 'Attribute Name',
    cell: ({ row }) => {
      const name = row.getValue('name') as string
      return (
        <div className="font-mono text-sm">
          <Badge variant="outline" className="font-mono text-xs">
            {name}
          </Badge>
        </div>
      )
    },
  },
  {
    accessorKey: 'values',
    header: 'Values',
    cell: ({ row }) => {
      const values = row.getValue('values') as string
      const valueArray = values.split(', ')
      const isMultipleValues = valueArray.length > 1
      
      return (
        <div className="max-w-md">
          {isMultipleValues ? (
            <div className="flex flex-wrap gap-1">
              {valueArray.map((value, index) => (
                <Badge 
                  key={index} 
                  variant="secondary" 
                  className="text-xs font-mono"
                >
                  {value}
                </Badge>
              ))}
            </div>
          ) : (
            <span className="text-sm font-mono text-muted-foreground">
              {values}
            </span>
          )}
        </div>
      )
    },
  },
]

// Summary component for attribute statistics
const AttributesSummary = ({ 
  entityType, 
  attributes 
}: { 
  entityType: string
  attributes: EntityAttribute[] 
}) => {
  const stats = useMemo(() => {
    const totalAttributes = attributes.length
    const attributesWithMultipleValues = attributes.filter(
      attr => attr.values.includes(',')
    ).length
    const totalUniqueValues = attributes.reduce((sum, attr) => {
      return sum + attr.values.split(', ').length
    }, 0)

    return {
      totalAttributes,
      attributesWithMultipleValues,
      totalUniqueValues
    }
  }, [attributes])

  if (attributes.length === 0) return null

  return (
    <div className="text-xs text-muted-foreground mb-3 px-1">
      <span className="font-medium">{entityType}</span> entities have{' '}
      <span className="font-medium">{stats.totalAttributes}</span> unique attributes with{' '}
      <span className="font-medium">{stats.totalUniqueValues}</span> total values
      {stats.attributesWithMultipleValues > 0 && (
        <span>
          {' '}({stats.attributesWithMultipleValues} attributes have multiple values)
        </span>
      )}
    </div>
  )
}

// Main EntityAttributesTable component
interface EntityAttributesTableProps {
  entityType: string
  className?: string
}

export const EntityAttributesTable = ({ 
  entityType, 
  className = '' 
}: EntityAttributesTableProps) => {
  const { data, isLoading, error } = useEntityAttributes(entityType)
  const columns = useMemo(() => createAttributesColumns(), [])

  // Handle loading state
  if (isLoading) {
    return (
      <div className={`bg-muted/30 rounded-md border ${className}`}>
        <AttributesLoadingState />
      </div>
    )
  }

  // Handle error state
  if (error) {
    return (
      <div className={`bg-muted/30 rounded-md border ${className}`}>
        <AttributesErrorState error={error} />
      </div>
    )
  }

  // Handle empty state
  if (!data || data.attributes.length === 0) {
    return (
      <div className={`bg-muted/30 rounded-md border ${className}`}>
        <AttributesEmptyState entityType={entityType} />
      </div>
    )
  }

  // Render table with data
  return (
    <div className={`bg-muted/30 rounded-md border p-4 ${className}`}>
      <AttributesSummary 
        entityType={entityType} 
        attributes={data.attributes} 
      />
      <DataTable
        columns={columns}
        data={data.attributes}
        currentPage={1}
        pageSize={data.attributes.length}
        onPageChange={() => {}} // No pagination for attributes table
        onPageSizeChange={() => {}} // No pagination for attributes table
        totalCount={data.attributes.length}
        showColumnVisibility={false}
        showPagination={false}
        showSearch={false}
      />
    </div>
  )
}

// Wrapper component for inline expansion within EntityCatalogTable
interface EntityAttributesExpansionProps {
  entityType: string
  isExpanded: boolean
  className?: string
}

export const EntityAttributesExpansion = ({
  entityType,
  isExpanded,
  className = ''
}: EntityAttributesExpansionProps) => {
  if (!isExpanded) return null

  return (
    <div className={`animate-in slide-in-from-top-2 duration-200 ${className}`}>
      <EntityAttributesTable entityType={entityType} />
    </div>
  )
}
