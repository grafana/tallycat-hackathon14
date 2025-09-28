import { createFileRoute } from '@tanstack/react-router'
import { Database, FileText, ArrowUpDown, X, ChevronDown } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useTechnicalFacets } from '@/hooks'
import type { TechnicalFacet, FacetOption } from '@/data/technical-facets'
import { Tabs, TabsTrigger, TabsList } from '@/components/ui/tabs'
import { useSchemas } from '@/hooks'
import { TelemetryCard } from '@/components/telemetry/telemetry-card'
import { DataTable } from '@/components/ui/data-table'
import type { ViewMode, Telemetry } from '@/types/telemetry'
import type { ColumnDef } from '@tanstack/react-table'
import { DataTypeIcon, getDataType } from '@/components/telemetry/telemetry-icons'
import { getTelemetryTypeBgColor } from '@/utils/telemetry'
import { formatDate, DateFormat } from '@/lib/utils'
import { Link } from '@tanstack/react-router'
import { useState } from 'react'
import type { SortField, SortDirection } from '@/hooks'
import { useDebounce } from '@/hooks'
import { SearchBar } from '@/components/telemetry-catalog/components/search/SearchBar'
import { FilterDropdown } from '@/components/telemetry-catalog/components/filters/FilterDropdown'

const columns: ColumnDef<Telemetry>[] = [
  {
    accessorKey: 'schemaKey',
    header: 'Name',
    cell: ({ row }) => {
      const item = row.original
      return (
        <div className="flex items-center gap-3">
          <div
            className={`flex h-8 w-8 items-center justify-center rounded-md ${getTelemetryTypeBgColor(
              item.telemetryType,
            )}`}
          >
            <DataTypeIcon dataType={getDataType(item)} />
          </div>
          <div>
            <Link
              to="/data-governance/$telemetryName"
              params={{ telemetryName: item.schemaKey }}
              className="font-medium hover:text-primary hover:underline"
            >
              {item.schemaKey}
            </Link>
            {/* <p className="text-xs text-muted-foreground line-clamp-1">{item.note}</p> */}
          </div>
        </div>
      )
    },
  },
  {
    accessorKey: 'telemetryType',
    header: 'Type',
    cell: ({ row }) => {
      const item = row.original
      return (
        <Badge variant="outline" className="capitalize">
          {item.telemetryType}
        </Badge>
      )
    },
  },
  {
    accessorKey: 'metricType',
    header: 'Data Type',
    cell: ({ row }) => {
      const item = row.original
      const dataType = getDataType(item)
      return (
        <div className="flex items-center gap-1.5">
          <DataTypeIcon dataType={dataType} />
          <span className="text-sm">{dataType}</span>
        </div>
      )
    },
  },
  {
    accessorKey: 'seenCount',
    header: 'Schema Versions',
    cell: ({ row }) => {
      const item = row.original
      return <span className="text-sm">{item.seenCount}</span>
    },
  },
  {
    accessorKey: 'format',
    header: 'Format',
    cell: ({ row }) => {
      const item = row.original
      return <span className="font-mono text-xs">{item.protocol}</span>
    },
  },
  {
    accessorKey: 'lastUpdated',
    header: 'Last Updated',
    cell: ({ row }) => {
      const item = row.original
      return formatDate(item.updatedAt, DateFormat.short)
    },
  },
  {
    id: 'actions',
    cell: () => {
      return (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="h-8 w-8">
              <ChevronDown className="h-4 w-4" />
              <span className="sr-only">Open menu</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem>View Details</DropdownMenuItem>
            <DropdownMenuItem>Edit</DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem>Export</DropdownMenuItem>
            <DropdownMenuItem className="text-red-500">Delete</DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      )
    },
  },
]

// Components
const ActiveFilters = ({
  activeFilters,
  removeFilter,
  clearAllFilters,
}: {
  activeFilters: Record<string, string[]>
  removeFilter: (facetId: string, value: string) => void
  clearAllFilters: () => void
}) => {
  const { data: technicalFacets } = useTechnicalFacets()

  if (!technicalFacets) return null

  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="text-sm text-muted-foreground">Active filters:</span>
      {Object.entries(activeFilters).map(([facetId, values]) =>
        values.map((value) => {
          const facet = technicalFacets.find(
            (f: TechnicalFacet) => f.id === facetId,
          )
          const option = facet?.options.find((o: FacetOption) => o.id === value)
          return (
            <Badge
              key={`${facetId}-${value}`}
              variant="secondary"
              className="flex items-center gap-1 px-2 py-1"
            >
              <span className="text-xs text-muted-foreground">
                {facet?.name}:
              </span>
              <span>{option?.name || value}</span>
              <Button
                variant="ghost"
                size="icon"
                className="h-4 w-4 p-0 ml-1"
                onClick={() => removeFilter(facetId, value)}
              >
                <X className="h-3 w-3" />
                <span className="sr-only">Remove filter</span>
              </Button>
            </Badge>
          )
        }),
      )}
      <Button
        variant="ghost"
        size="sm"
        className="h-7 text-xs"
        onClick={clearAllFilters}
      >
        Clear all
      </Button>
    </div>
  )
}

const SortDropdown = ({
  sortField,
  sortDirection,
  onSort,
}: {
  sortField: SortField
  sortDirection: SortDirection
  onSort: (field: SortField) => void
}) => (
  <DropdownMenu>
    <DropdownMenuTrigger asChild>
      <Button variant="outline" size="sm" className="h-9">
        <ArrowUpDown className="mr-2 h-4 w-4" />
        Sort
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="end">
      <DropdownMenuItem onClick={() => onSort('name')}>
        Name {sortField === 'name' && (sortDirection === 'asc' ? '↑' : '↓')}
      </DropdownMenuItem>
      <DropdownMenuItem onClick={() => onSort('lastUpdated')}>
        Last Updated{' '}
        {sortField === 'lastUpdated' && (sortDirection === 'asc' ? '↑' : '↓')}
      </DropdownMenuItem>
      <DropdownMenuItem onClick={() => onSort('type')}>
        Telemetry Type{' '}
        {sortField === 'type' && (sortDirection === 'asc' ? '↑' : '↓')}
      </DropdownMenuItem>
      <DropdownMenuItem onClick={() => onSort('dataType')}>
        Data Type{' '}
        {sortField === 'dataType' && (sortDirection === 'asc' ? '↑' : '↓')}
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
)

const ViewToggle = ({
  viewMode,
  onViewModeChange,
}: {
  viewMode: ViewMode
  onViewModeChange: (mode: ViewMode) => void
}) => (
  <div className="flex items-center gap-1 rounded-md border p-1">
    <Button
      variant={viewMode === 'grid' ? 'secondary' : 'ghost'}
      size="sm"
      className="h-7 w-7 p-0"
      onClick={() => onViewModeChange('grid')}
      aria-label="Grid view"
    >
      <Database className="h-4 w-4" />
    </Button>
    <Button
      variant={viewMode === 'list' ? 'secondary' : 'ghost'}
      size="sm"
      className="h-7 w-7 p-0"
      onClick={() => onViewModeChange('list')}
      aria-label="List view"
    >
      <FileText className="h-4 w-4" />
    </Button>
  </div>
)

export const SchemaCatalog = () => {
  const [searchQuery, setSearchQuery] = useState('')
  const debouncedSearchQuery = useDebounce(
    searchQuery.length >= 2 ? searchQuery : '',
    600,
  )
  const [viewMode, setViewMode] = useState<ViewMode>('list')
  const [activeFilters, setActiveFilters] = useState<Record<string, string[]>>(
    {},
  )
  const [activeFilterCount, setActiveFilterCount] = useState(0)
  const [sortField, setSortField] = useState<SortField>('name')
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc')
  const [activeTab, setActiveTab] = useState('all')
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)

  const {
    data: schemasData,
    isLoading: isLoadingSchemas,
    error: schemasError,
  } = useSchemas({
    page: currentPage,
    pageSize,
    search: debouncedSearchQuery,
    type: activeTab !== 'all' ? activeTab : undefined,
    sortField,
    sortDirection,
    filters: activeFilters,
  })

  const { isLoading: isLoadingFacets, error: facetsError } =
    useTechnicalFacets()

  const toggleFilter = (facetId: string, value: string) => {
    setActiveFilters((prev) => {
      const currentValues = prev[facetId] || []
      const newValues = currentValues.includes(value)
        ? currentValues.filter((v) => v !== value)
        : [...currentValues, value]

      // Update filter count
      const newCount = Object.values({
        ...prev,
        [facetId]: newValues,
      }).reduce((acc, curr) => acc + curr.length, 0)
      setActiveFilterCount(newCount)

      // Reset to first page when filters change
      setCurrentPage(1)

      return {
        ...prev,
        [facetId]: newValues,
      }
    })
  }

  const removeFilter = (facetId: string, value: string) => {
    setActiveFilters((prev) => {
      const currentValues = prev[facetId] || []
      const newValues = currentValues.filter((v) => v !== value)

      // Update filter count
      const newCount = Object.values({
        ...prev,
        [facetId]: newValues,
      }).reduce((acc, curr) => acc + curr.length, 0)
      setActiveFilterCount(newCount)

      // Reset to first page when filters change
      setCurrentPage(1)

      return {
        ...prev,
        [facetId]: newValues,
      }
    })
  }

  const clearAllFilters = () => {
    setActiveFilters({})
    setActiveFilterCount(0)
    // Reset to first page when filters change
    setCurrentPage(1)
  }

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortField(field)
      setSortDirection('asc')
    }
    // Reset to first page when sorting changes
    setCurrentPage(1)
  }

  const handlePageChange = (page: number) => {
    setCurrentPage(page)
  }

  const handlePageSizeChange = (size: number) => {
    setPageSize(size)
    setCurrentPage(1) // Reset to first page when changing page size
  }

  if (isLoadingSchemas) {
    return (
      <div className="flex h-[50vh] items-center justify-center">
        <p className="text-muted-foreground">Loading schemas...</p>
      </div>
    )
  }

  if (schemasError) {
    return (
      <div className="flex h-[50vh] items-center justify-center">
        <p className="text-destructive">
          Error loading schemas. Please try again later.
        </p>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-medium">Telemetry Catalog</h1>
        <p className="text-muted-foreground">
          Browse and manage your observability telemetry signals
        </p>
      </div>

      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <SearchBar value={searchQuery} onChange={setSearchQuery} />

          <div className="flex flex-wrap items-center gap-2">
            <FilterDropdown
              activeFilters={activeFilters}
              activeFilterCount={activeFilterCount}
              onToggleFilter={toggleFilter}
              isLoading={isLoadingFacets}
              error={facetsError}
            />
            <SortDropdown
              sortField={sortField}
              sortDirection={sortDirection}
              onSort={handleSort}
            />
            <ViewToggle viewMode={viewMode} onViewModeChange={setViewMode} />
          </div>
        </div>

        {activeFilterCount > 0 && (
          <ActiveFilters
            activeFilters={activeFilters}
            removeFilter={removeFilter}
            clearAllFilters={clearAllFilters}
          />
        )}

        <Tabs
          value={activeTab}
          onValueChange={setActiveTab}
          className="w-full"
        >
          <TabsList>
            <TabsTrigger value="all">All Telemetry</TabsTrigger>
            <TabsTrigger value="metric">Metrics</TabsTrigger>
            <TabsTrigger value="log">Logs</TabsTrigger>
            <TabsTrigger value="span">Spans</TabsTrigger>
            <TabsTrigger value="profile">Profiles</TabsTrigger>
          </TabsList>
        </Tabs>

        {/* Table and bottom controls */}
        <div className="flex flex-col gap-2">
          {viewMode === 'grid' ? (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {schemasData?.items.map((item) => (
                <TelemetryCard
                  key={`${item.schemaId}-${item.schemaKey}`}
                  item={item}
                />
              ))}
            </div>
          ) : (
            <DataTable
              columns={columns}
              data={schemasData?.items || []}
              currentPage={currentPage}
              pageSize={pageSize}
              onPageChange={handlePageChange}
              onPageSizeChange={handlePageSizeChange}
              totalCount={schemasData?.total || 0}
              showColumnVisibility={false}
            />
          )}
        </div>
      </div>
    </div>
  )
}

export const Route = createFileRoute('/data-governance/telemetry-catalog')({
  component: SchemaCatalog,
})
