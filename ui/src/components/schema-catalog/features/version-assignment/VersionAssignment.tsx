"use client"

import { useState, useCallback, useMemo } from "react"
import {
  Edit,
  Tag,
  Database,
  Eye,
  MoreHorizontal,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { DataTable } from "@/components/ui/data-table"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import type { ColumnDef } from "@tanstack/react-table"
import { SchemaDetailsModal } from "@/components/schema-catalog/modals/SchemaDetailsModal"
import type { Schema, VersionAssignmentViewProps } from "@/types/schema-catalog"
import { Status } from "@/types/telemetry"
import { formatDate, DateFormat } from "@/lib/utils"
import { VersionAssignmentDialog } from "./components/VersionAssignmentDialog"
import { StatusBadge } from "@/components/schema-catalog/components/badges/StatusBadge"
import { SearchAndFilterBar } from "./components/SearchAndFilterBar"
import { LoadingState } from "@/components/schema-catalog/components/states/LoadingState"
import { ErrorState } from "@/components/schema-catalog/components/states/ErrorState"
import { useSchemaAssignmentData } from "@/hooks/use-schema-assignment-data"

const createTableColumns = (
  onViewSchema: (schema: Schema) => void,
  onAssignVersion: (schema: Schema) => void
): ColumnDef<Schema>[] => [
  {
    accessorKey: "id",
    header: "Schema ID",
    cell: ({ row }) => (
      <div className="font-mono text-sm text-foreground">{row.original.id}</div>
    ),
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => <StatusBadge status={row.original.status} />,
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
    cell: ({ row }) => (
      <span className="text-sm text-muted-foreground">
        {formatDate(row.original.lastSeen, DateFormat.short)}
      </span>
    ),
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
              <DropdownMenuItem onClick={() => onViewSchema(schema)} className="cursor-pointer">
                <Eye className="mr-2 h-4 w-4" />
                View Details
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => onAssignVersion(schema)} className="cursor-pointer">
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

const getSummaryText = (
  count: number,
  total: number,
  searchQuery: string,
  activeStatus: string[]
) => {
  return `Showing ${count} schema${count !== 1 ? "s" : ""}${
    (searchQuery || activeStatus.length > 0) ? ` (filtered from ${total} total)` : ""
  }`
}

export function VersionAssignmentView({
  schemaData,
  onVersionChange,
}: VersionAssignmentViewProps) {
  const [selectedSchema, setSelectedSchema] = useState<string | null>(null)
  const [isAssigning, setIsAssigning] = useState(false)
  const [viewingSchema, setViewingSchema] = useState<Schema | null>(null)

  const {
    searchQuery,
    setSearchQuery,
    activeStatus,
    setActiveStatus,
    currentPage,
    setCurrentPage,
    pageSize,
    setPageSize,
    tableData,
    isLoading,
    error,
    totalCount,
  } = useSchemaAssignmentData(schemaData.schemaKey)

  const handleAssignVersion = useCallback((schema: Schema) => {
    setSelectedSchema(schema.id)
    setIsAssigning(true)
  }, [])

  const handleViewSchema = useCallback((schema: Schema) => {
    setViewingSchema(schema)
  }, [])

  const columns = useMemo(() => 
    createTableColumns(handleViewSchema, handleAssignVersion),
    [handleViewSchema, handleAssignVersion]
  )

  if (isLoading) {
    return <LoadingState />
  }

  if (error) {
    return <ErrorState />
  }

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

      <SearchAndFilterBar
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        activeStatus={activeStatus}
        onStatusChange={setActiveStatus}
      />

      <DataTable
        columns={columns}
        data={tableData}
        currentPage={currentPage}
        pageSize={pageSize}
        onPageChange={setCurrentPage}
        onPageSizeChange={setPageSize}
        totalCount={totalCount}
        showColumnVisibility={false}
        summaryText={getSummaryText(tableData.length, totalCount, searchQuery, activeStatus)}
      />

      <SchemaDetailsModal
        viewingSchema={viewingSchema}
        onClose={() => setViewingSchema(null)}
        schemaData={schemaData}
      />

      <VersionAssignmentDialog
        isOpen={isAssigning}
        onClose={() => setIsAssigning(false)}
        selectedSchema={selectedSchema}
        currentSchema={tableData.find((s) => s.id === selectedSchema) ?? null}
        onVersionChange={onVersionChange}
      />
    </div>
  )
}
