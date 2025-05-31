"use client"

import { useState, useMemo, useCallback } from "react"
import {
  CheckCircle2,
  AlertTriangle,
  XCircle,
  Edit,
  Save,
  X,
  Tag,
  Database,
  Eye,
  MoreHorizontal,
  Search,
  SlidersHorizontal,
  Info,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { DataTable } from "@/components/ui/data-table"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import type { ColumnDef } from "@tanstack/react-table"
import { SchemaDetailsModal } from "@/components/schema-catalog/SchemaDetailsModal"
import type { Schema, VersionAssignmentViewProps, AssignmentForm, VersionValidation } from "@/types/schema-catalog"
import { Status } from "@/types/telemetry"
import { useSchemaAssignments } from '@/hooks/use-schema-assignments'

// Utility functions
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  })
}

const getStatusBadge = (status: string) => {
  switch (status) {
    case 'Assigned':
      return (
        <Badge variant="outline" className="bg-green-500/20 text-green-400 border-green-500/30 font-medium">
          <CheckCircle2 className="h-3 w-3 mr-1" />
          Assigned
        </Badge>
      )
    case 'Unassigned':
      return (
        <Badge variant="outline" className="bg-yellow-500/20 text-yellow-400 border-yellow-500/30 font-medium">
          <AlertTriangle className="h-3 w-3 mr-1" />
          Unassigned
        </Badge>
      )
    default:
      return null
  }
}

// Semantic versioning validation
const validateSemanticVersion = (version: string): VersionValidation => {
  if (!version.trim()) {
    return { isValid: false, message: "Version is required" }
  }

  const semverRegex =
    /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$/

  if (!semverRegex.test(version)) {
    return {
      isValid: false,
      message: "Invalid semantic version format. Use MAJOR.MINOR.PATCH (e.g., 1.2.3, 2.0.0-beta.1)",
    }
  }

  return { isValid: true, message: "Valid semantic version" }
}

// Add ActiveFilters component before VersionAssignmentView
const ActiveFilters = ({
  activeFilters,
  removeFilter,
  clearAllFilters,
}: {
  activeFilters: Record<string, string[]>
  removeFilter: (facetId: string, value: string) => void
  clearAllFilters: () => void
}) => {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="text-sm text-muted-foreground">Active filters:</span>
      {Object.entries(activeFilters).map(([facetId, values]) =>
        values.map((value) => (
          <Badge key={`${facetId}-${value}`} variant="secondary" className="flex items-center gap-1 px-2 py-1">
            <span className="text-xs text-muted-foreground">{facetId}:</span>
            <span>{value}</span>
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
        )),
      )}
      <Button variant="ghost" size="sm" className="h-7 text-xs" onClick={clearAllFilters}>
        Clear all
      </Button>
    </div>
  )
}

export function VersionAssignmentView({
  schemaData,
  onVersionChange,
}: VersionAssignmentViewProps) {
  // State management
  const [selectedSchema, setSelectedSchema] = useState<string | null>(null)
  const [isAssigning, setIsAssigning] = useState(false)
  const [viewingSchema, setViewingSchema] = useState<Schema | null>(null)
  const [assignmentForm, setAssignmentForm] = useState<AssignmentForm>({
    version: "",
    description: "",
  })
  const [versionValidation, setVersionValidation] = useState<VersionValidation>({ isValid: true, message: "" })

  // Filter states
  const [searchQuery, setSearchQuery] = useState("")
  const [activeStatus, setActiveStatus] = useState<string[]>([])
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)

  // Fetch schema assignments from backend with filtering
  const { data, isLoading, error } = useSchemaAssignments({
    schemaKey: schemaData.schemaKey,
    search: searchQuery,
    status: activeStatus,
    page: currentPage,
    pageSize,
  })

  // Event handlers
  const handleSearchChange = useCallback((value: string) => {
    setSearchQuery(value)
  }, [])

  const handleVersionChange = useCallback((value: string) => {
    setAssignmentForm((prev) => ({ ...prev, version: value }))
    const validation = validateSemanticVersion(value)
    setVersionValidation(validation)
  }, [])

  const handleAssignVersion = useCallback((schema: Schema) => {
    setSelectedSchema(schema.id)
    setIsAssigning(true)
    setAssignmentForm({
      version: schema.version || "",
      description: "",
    })
    setVersionValidation({ isValid: true, message: "" })
  }, [])

  const handleSaveAssignment = useCallback(() => {
    const validation = validateSemanticVersion(assignmentForm.version)
    if (!validation.isValid) {
      setVersionValidation(validation)
      return
    }

    onVersionChange(assignmentForm.version)
    setIsAssigning(false)
    setSelectedSchema(null)
    setAssignmentForm({ version: "", description: "" })
    setVersionValidation({ isValid: true, message: "" })
  }, [assignmentForm, onVersionChange])

  const handleCancelAssignment = useCallback(() => {
    setIsAssigning(false)
    setSelectedSchema(null)
    setAssignmentForm({ version: "", description: "" })
    setVersionValidation({ isValid: true, message: "" })
  }, [])

  const handleViewSchema = useCallback((schema: Schema) => {
    setViewingSchema(schema)
  }, [])

  // Map backend assignments to Schema[] for DataTable
  const tableData = (data?.items ?? []).map(item => ({
    id: item.schemaId,
    name: item.schemaId,
    status: item.version && item.version !== 'Unassigned' ? 'Assigned' : 'Unassigned',
    version: item.version === 'Unassigned' ? null : item.version,
    producers: Array(item.producerCount).fill({}),
    lastSeen: item.lastSeen,
    discoveredAt: '',
    resourceAttributes: [],
    instrumentationAttributes: [],
    telemetryAttributes: [],
  }))

  // Get the current schema being edited
  const currentSchema = selectedSchema ? tableData.find((s) => s.id === selectedSchema) : null

  if (isLoading) {
    return (
      <div className="flex h-[50vh] items-center justify-center">
        <p className="text-muted-foreground">Loading schema assignments...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex h-[50vh] items-center justify-center">
        <p className="text-destructive">Error loading schema assignments. Please try again later.</p>
      </div>
    )
  }

  // Add columns definition here, after the handlers
  const columns: ColumnDef<Schema>[] = [
    {
      accessorKey: "id",
      header: "Schema ID",
      cell: ({ row }) => {
        const schema = row.original
        return (
          <div className="font-mono text-sm text-foreground">{schema.id}</div>
        )
      },
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => {
        const schema = row.original
        return getStatusBadge(schema.status)
      },
    },
    {
      accessorKey: "version",
      header: "Version",
      cell: ({ row }) => {
        const schema = row.original
        return schema.version ? (
          <Badge
            variant="outline"
            className="font-mono text-xs bg-blue-500/10 text-blue-400 border-blue-500/30"
          >
            <Tag className="h-3 w-3 mr-1" />v{schema.version}
          </Badge>
        ) : (
          <span className="text-muted-foreground text-sm">Unassigned</span>
        )
      },
    },
    {
      accessorKey: "producers",
      header: "Producers",
      cell: ({ row }) => {
        const schema = row.original
        return (
          <div className="flex items-center gap-2">
            <Database className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm text-foreground">{schema.producers.length}</span>
          </div>
        )
      },
    },
    {
      accessorKey: "lastSeen",
      header: "Last Seen",
      cell: ({ row }) => {
        const schema = row.original
        return <span className="text-sm text-muted-foreground">{formatDate(schema.lastSeen)}</span>
      },
    },
    {
      id: "actions",
      cell: ({ row }) => {
        const schema = row.original
        return (
          <div className="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="h-8 w-8 p-0 hover:bg-muted">
                  <span className="sr-only">Open menu</span>
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-48">
                <DropdownMenuLabel>Actions</DropdownMenuLabel>
                <DropdownMenuItem onClick={() => handleViewSchema(schema)} className="cursor-pointer">
                  <Eye className="mr-2 h-4 w-4" />
                  View Details
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={() => handleAssignVersion(schema)} className="cursor-pointer">
                  {schema.status === Status.Experimental ? (
                    <>
                      <Tag className="mr-2 h-4 w-4" />
                      Assign Version
                    </>
                  ) : (
                    <>
                      <Edit className="mr-2 h-4 w-4" />
                      Update Version
                    </>
                  )}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        )
      },
    },
  ]

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold text-foreground">Schema Version Assignment</h2>
          <p className="text-sm text-muted-foreground">
            Manage schema versions discovered at runtime for {schemaData.name}
          </p>
        </div>
      </div>

      {/* Search and Filter Bar */}
      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between gap-4">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search schema IDs..."
              value={searchQuery}
              onChange={(e) => handleSearchChange(e.target.value)}
              className="pl-9 pr-4"
            />
          </div>
          <div className="flex items-center gap-2">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="outline"
                  size="sm"
                  className={`h-9 gap-1 ${activeStatus.length > 0 ? "bg-primary/10 border-primary/30 text-primary" : ""}`}
                >
                  <SlidersHorizontal className="h-4 w-4" />
                  Filter
                  {activeStatus.length > 0 && (
                    <Badge variant="secondary" className="ml-1">
                      {activeStatus.length}
                    </Badge>
                  )}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56">
                <DropdownMenuLabel>Filter by Status</DropdownMenuLabel>
                <DropdownMenuSeparator />
                {["assigned", "unassigned", "deprecated"].map((status) => (
                  <DropdownMenuItem
                    key={status}
                    className="cursor-pointer"
                    onClick={() => setActiveStatus([status])}
                  >
                    <div className="flex items-center gap-2">
                      {activeStatus.includes(status) ? (
                        <CheckCircle2 className="h-4 w-4 text-primary" />
                      ) : (
                        <div className="h-4 w-4" />
                      )}
                      <span className="capitalize">{status}</span>
                    </div>
                  </DropdownMenuItem>
                ))}
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  className="cursor-pointer text-muted-foreground"
                  onClick={() => setActiveStatus([])}
                >
                  Clear all filters
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </div>

      {/* Data Table */}
      <DataTable
        columns={columns}
        data={tableData}
        currentPage={currentPage}
        pageSize={pageSize}
        onPageChange={setCurrentPage}
        onPageSizeChange={setPageSize}
        totalCount={data?.total ?? 0}
        showColumnVisibility={false}
        summaryText={`Showing ${tableData.length} schema${tableData.length !== 1 ? "s" : ""}${
          (searchQuery || activeStatus.length > 0) ? ` (filtered from ${data?.total ?? 0} total)` : ""
        }`}
      />

      {/* Schema Details Modal */}
      <SchemaDetailsModal
        viewingSchema={viewingSchema}
        onClose={() => setViewingSchema(null)}
        schemaData={schemaData}
      />

      {/* Version Assignment Dialog */}
      <Dialog open={isAssigning} onOpenChange={() => setIsAssigning(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Tag className="h-5 w-5" />
              {currentSchema?.status === Status.Experimental ? "Assign Schema Version" : "Update Schema Version"}
            </DialogTitle>
            <DialogDescription>
              {currentSchema?.status === Status.Experimental
                ? `Assign a semantic version to schema ${selectedSchema}`
                : `Update the version for schema ${selectedSchema}`}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {/* Current Version Info */}
            {currentSchema?.version && (
              <div className="p-3 bg-muted/50 rounded-lg border">
                <div className="flex items-center gap-2 text-sm">
                  <Info className="h-4 w-4 text-muted-foreground" />
                  <span className="text-muted-foreground">Current version:</span>
                  <Badge variant="outline" className="font-mono text-xs">
                    v{currentSchema.version}
                  </Badge>
                </div>
              </div>
            )}

            {/* Version Input */}
            <div className="space-y-2">
              <Label htmlFor="version" className="text-sm font-medium">
                Version *
              </Label>
              <Input
                id="version"
                placeholder="e.g., 1.2.3, 2.0.0-beta.1"
                value={assignmentForm.version}
                onChange={(e) => handleVersionChange(e.target.value)}
                className={`font-mono ${!versionValidation.isValid ? "border-red-500 focus:border-red-500" : ""}`}
              />
              {!versionValidation.isValid && (
                <p className="text-xs text-red-500 flex items-center gap-1">
                  <XCircle className="h-3 w-3" />
                  {versionValidation.message}
                </p>
              )}
            </div>

            {/* Change Description */}
            <div className="space-y-2">
              <Label htmlFor="description" className="text-sm font-medium">
                Change Description *
              </Label>
              <Textarea
                id="description"
                placeholder="Describe the changes or reason for this version assignment..."
                value={assignmentForm.description}
                onChange={(e) => setAssignmentForm({ ...assignmentForm, description: e.target.value })}
                rows={3}
              />
            </div>

            {/* Action Buttons */}
            <div className="flex items-center gap-2 pt-2">
              <Button
                onClick={handleSaveAssignment}
                disabled={!versionValidation.isValid || !assignmentForm.version || !assignmentForm.description}
                className="flex items-center gap-2"
              >
                <Save className="h-4 w-4" />
                {currentSchema?.status === Status.Experimental ? "Assign Version" : "Update Version"}
              </Button>
              <Button variant="outline" onClick={handleCancelAssignment} className="flex items-center gap-2">
                <X className="h-4 w-4" />
                Cancel
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
