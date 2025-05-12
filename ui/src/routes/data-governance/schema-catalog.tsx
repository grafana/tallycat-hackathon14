import { createFileRoute } from '@tanstack/react-router'
import { Database, FileText, ArrowUpDown, Filter, Search, X, ChevronDown } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { DropdownMenu, DropdownMenuCheckboxItem, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useTechnicalFacets } from '@/hooks/use-technical-facets'
import type { TechnicalFacet, FacetOption } from '@/data/technical-facets'
import { Tabs, TabsTrigger, TabsList } from '@/components/ui/tabs'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useTelemetry } from '@/hooks/use-telemetry'
import { TelemetryCard } from '@/components/telemetry/telemetry-card'
import { DataTable } from '@/components/ui/data-table'
import type { SortField, ViewMode, TelemetryItem } from '@/types/telemetry'
import type { ColumnDef } from '@tanstack/react-table'
import { DataTypeIcon } from '@/components/telemetry/telemetry-icons'
import { getTelemetryTypeBgColor, formatDate, getStatusBadge } from '@/utils/telemetry'
import { Link } from '@tanstack/react-router'

const columns: ColumnDef<TelemetryItem>[] = [
  {
    accessorKey: "name",
    header: "Name",
    cell: ({ row }) => {
      const item = row.original
      return (
        <div className="flex items-center gap-3">
          <div
            className={`flex h-8 w-8 items-center justify-center rounded-md ${getTelemetryTypeBgColor(
              item.type,
            )}`}
          >
            <DataTypeIcon dataType={item.dataType} />
          </div>
          <div>
            <Link
              to={`/data-governance/schema-catalog`}
              className="font-medium hover:text-primary hover:underline"
            >
              {item.name}
            </Link>
            <p className="text-xs text-muted-foreground line-clamp-1">{item.description}</p>
          </div>
        </div>
      )
    },
  },
  {
    accessorKey: "type",
    header: "Type",
    cell: ({ row }) => {
      const item = row.original
      return (
        <Badge variant="outline" className="capitalize">
          {item.type}
        </Badge>
      )
    },
  },
  {
    accessorKey: "dataType",
    header: "Data Type",
    cell: ({ row }) => {
      const item = row.original
      return (
        <div className="flex items-center gap-1.5">
          <DataTypeIcon dataType={item.dataType} />
          <span className="text-sm">{item.dataType}</span>
        </div>
      )
    },
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => {
      const item = row.original
      const statusBadge = getStatusBadge(item.status)
      return statusBadge ? (
        <Badge variant="outline" className={statusBadge.className}>
          {statusBadge.label}
        </Badge>
      ) : null
    },
  },
  {
    accessorKey: "format",
    header: "Format",
    cell: ({ row }) => {
      const item = row.original
      return <span className="font-mono text-xs">{item.format}</span>
    },
  },
  {
    accessorKey: "lastUpdated",
    header: "Last Updated",
    cell: ({ row }) => {
      const item = row.original
      return formatDate(item.lastUpdated)
    },
  },
  {
    id: "actions",
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
const SearchBar = ({ value, onChange }: { value: string; onChange: (value: string) => void }) => (
  <div className="relative w-full sm:max-w-md">
    <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
    <Input
      type="search"
      placeholder="Search telemetry signals..."
      className="w-full pl-9 pr-4"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  </div>
)

const FilterDropdown = ({
  activeFilters,
  activeFilterCount,
  onToggleFilter,
  isLoading,
  error
}: {
  activeFilters: Record<string, string[]>
  activeFilterCount: number
  onToggleFilter: (facetId: string, value: string) => void
  isLoading: boolean
  error: Error | null
}) => {
  const { data: technicalFacets } = useTechnicalFacets()

  if (isLoading) {
    return (
      <Button variant="outline" size="sm" className="h-9" disabled>
        <Filter className="mr-2 h-4 w-4" />
        Loading...
      </Button>
    )
  }

  if (error) {
    return (
      <Button variant="outline" size="sm" className="h-9 text-destructive" disabled>
        <Filter className="mr-2 h-4 w-4" />
        Error loading filters
      </Button>
    )
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="h-9">
          <Filter className="mr-2 h-4 w-4" />
          Filter
          {activeFilterCount > 0 && (
            <Badge variant="secondary" className="ml-1 h-5 min-w-5 px-1">
              {activeFilterCount}
            </Badge>
          )}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-[220px]">
        <ScrollArea className="h-[500px]">
          {technicalFacets?.map((facet: TechnicalFacet) => (
            <div key={facet.id} className="px-2 py-1.5">
              <DropdownMenuLabel className="px-0">{facet.name}</DropdownMenuLabel>
              <DropdownMenuSeparator className="mb-1" />
              {facet.options.map((option: FacetOption) => (
                <DropdownMenuCheckboxItem
                  key={option.id}
                  checked={(activeFilters[facet.id] || []).includes(option.id)}
                  onCheckedChange={() => onToggleFilter(facet.id, option.id)}
                >
                  {option.name}
                </DropdownMenuCheckboxItem>
              ))}
              {facet !== technicalFacets[technicalFacets.length - 1] && (
                <DropdownMenuSeparator className="mt-1" />
              )}
            </div>
          ))}
        </ScrollArea>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

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
          const facet = technicalFacets.find((f: TechnicalFacet) => f.id === facetId)
          const option = facet?.options.find((o: FacetOption) => o.id === value)
          return (
            <Badge key={`${facetId}-${value}`} variant="secondary" className="flex items-center gap-1 px-2 py-1">
              <span className="text-xs text-muted-foreground">{facet?.name}:</span>
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
      <Button variant="ghost" size="sm" className="h-7 text-xs" onClick={clearAllFilters}>
        Clear all
      </Button>
    </div>
  )
}

const SortDropdown = ({
  sortField,
  sortDirection,
  onSort
}: {
  sortField: SortField
  sortDirection: string
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
      <DropdownMenuItem onClick={() => onSort("name")}>
        Name {sortField === "name" && (sortDirection === "asc" ? "↑" : "↓")}
      </DropdownMenuItem>
      <DropdownMenuItem onClick={() => onSort("lastUpdated")}>
        Last Updated {sortField === "lastUpdated" && (sortDirection === "asc" ? "↑" : "↓")}
      </DropdownMenuItem>
      <DropdownMenuItem onClick={() => onSort("type")}>
        Telemetry Type {sortField === "type" && (sortDirection === "asc" ? "↑" : "↓")}
      </DropdownMenuItem>
      <DropdownMenuItem onClick={() => onSort("dataType")}>
        Data Type {sortField === "dataType" && (sortDirection === "asc" ? "↑" : "↓")}
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
)

const ViewToggle = ({
  viewMode,
  onViewModeChange
}: {
  viewMode: ViewMode
  onViewModeChange: (mode: ViewMode) => void
}) => (
  <div className="flex items-center gap-1 rounded-md border p-1">
    <Button
      variant={viewMode === "grid" ? "secondary" : "ghost"}
      size="sm"
      className="h-7 w-7 p-0"
      onClick={() => onViewModeChange("grid")}
      aria-label="Grid view"
    >
      <Database className="h-4 w-4" />
    </Button>
    <Button
      variant={viewMode === "list" ? "secondary" : "ghost"}
      size="sm"
      className="h-7 w-7 p-0"
      onClick={() => onViewModeChange("list")}
      aria-label="List view"
    >
      <FileText className="h-4 w-4" />
    </Button>
  </div>
)

export const SchemaCatalog = () => {
  const {
    searchQuery,
    setSearchQuery,
    viewMode,
    setViewMode,
    activeFilters,
    activeFilterCount,
    toggleFilter,
    removeFilter,
    clearAllFilters,
    sortField,
    sortDirection,
    handleSort,
    activeTab,
    setActiveTab,
    filteredItems,
  } = useTelemetry()

  const { isLoading, error } = useTechnicalFacets()

  return (
    <div className="mx-auto">
      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-2">
          <h1 className="text-3xl font-medium">Schema Catalog</h1>
          <p className="text-muted-foreground">Browse and manage your observability telemetry signals</p>
        </div>

        <div className="flex flex-col gap-6">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <SearchBar value={searchQuery} onChange={setSearchQuery} />

            <div className="flex flex-wrap items-center gap-2">
              <FilterDropdown
                activeFilters={activeFilters}
                activeFilterCount={activeFilterCount}
                onToggleFilter={toggleFilter}
                isLoading={isLoading}
                error={error}
              />
              <SortDropdown
                sortField={sortField}
                sortDirection={sortDirection}
                onSort={handleSort}
              />
              <ViewToggle
                viewMode={viewMode}
                onViewModeChange={setViewMode}
              />
            </div>
          </div>

          {activeFilterCount > 0 && (
            <ActiveFilters
              activeFilters={activeFilters}
              removeFilter={removeFilter}
              clearAllFilters={clearAllFilters}
            />
          )}

          <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
            <TabsList>
              <TabsTrigger value="all">All Telemetry</TabsTrigger>
              <TabsTrigger value="metric">Metrics</TabsTrigger>
              <TabsTrigger value="log">Logs</TabsTrigger>
              <TabsTrigger value="trace">Traces</TabsTrigger>
            </TabsList>
          </Tabs>

          <div className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground">
              Showing <span className="font-medium text-foreground">{filteredItems.length}</span> telemetry signals
              {activeFilterCount > 0 && " with applied filters"}
            </p>
            <Select defaultValue="10">
              <SelectTrigger className="w-[80px] h-8 text-xs">
                <SelectValue placeholder="10 per page" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="10">10</SelectItem>
                <SelectItem value="20">20</SelectItem>
                <SelectItem value="50">50</SelectItem>
                <SelectItem value="100">100</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Results */}

          {viewMode === "grid" ? (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {filteredItems.map((item) => (
                <TelemetryCard key={item.id} item={item} />
              ))}
            </div>
          ) : (
            <DataTable
              columns={columns}
              data={filteredItems}
              searchKey="name"
              searchPlaceholder="Search telemetry signals..."
              pageSize={10}
              showColumnVisibility={false}
              showPagination={true}
              showSearch={false}
            />
          )}

          <div className="flex items-center justify-between">
            <Button variant="outline" size="sm" disabled>
              Previous
            </Button>
            <div className="flex items-center gap-1">
              <Button variant="outline" size="sm" className="h-8 w-8 p-0 font-medium">
                1
              </Button>
              <Button variant="ghost" size="sm" className="h-8 w-8 p-0" disabled>
                2
              </Button>
              <Button variant="ghost" size="sm" className="h-8 w-8 p-0" disabled>
                3
              </Button>
            </div>
            <Button variant="outline" size="sm" disabled>
              Next
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}

export const Route = createFileRoute('/data-governance/schema-catalog')({
  component: SchemaCatalog,
}) 