"use client"

import { Database, Users, Search, SlidersHorizontal, CheckCircle2, Server, MoreHorizontal, Eye, AlertTriangle } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { DataTable } from "@/components/ui/data-table"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { SchemaDefinitionView } from "@/components/schema-catalog/features/schema-definition/SchemaDefinitionView"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import type { ColumnDef } from "@tanstack/react-table"
import type { Schema, Producer } from "@/types/schema-catalog"

interface SchemaDetailsModalProps {
  viewingSchema: Schema | null
  onClose: () => void
  schemaData: any
}

// Producer columns definition
const producerColumns: ColumnDef<Producer>[] = [
  {
    accessorKey: "name",
    header: "Service",
    cell: ({ row }) => {
      const producer = row.original
      return (
        <div className="flex items-center gap-3">
          <div className="p-1.5 rounded-md bg-primary/10">
            <Server className="h-4 w-4 text-primary" />
          </div>
          <div>
            <div className="font-medium text-sm">{producer.name}</div>
            <div className="text-xs text-muted-foreground font-mono">{producer.id}</div>
          </div>
        </div>
      )
    },
  },
  {
    accessorKey: "team",
    header: "Team",
    cell: ({ row }) => {
      const producer = row.original
      return (
        <Badge variant="outline" className="text-xs">
          {producer.team}
        </Badge>
      )
    },
  },
  {
    accessorKey: "environment",
    header: "Environment",
    cell: ({ row }) => {
      const producer = row.original
      return (
        <Badge
          variant="outline"
          className={`text-xs ${
            producer.environment === "production"
              ? "bg-red-500/10 text-red-500 border-red-500/20"
              : producer.environment === "staging"
                ? "bg-yellow-500/10 text-yellow-500 border-yellow-500/20"
                : "bg-blue-500/10 text-blue-500 border-blue-500/20"
          }`}
        >
          {producer.environment}
        </Badge>
      )
    },
  },
  {
    id: "health",
    header: "Health",
    cell: () => (
      <Badge
        variant="outline"
        className="bg-green-500/10 text-green-500 border-green-500/20 text-xs"
      >
        <CheckCircle2 className="h-3 w-3 mr-1" />
        Healthy
      </Badge>
    ),
  },
  {
    id: "actions",
    cell: () => (
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
            <DropdownMenuItem className="cursor-pointer">
              <Eye className="mr-2 h-4 w-4" />
              View Service Details
            </DropdownMenuItem>
            <DropdownMenuItem className="cursor-pointer">
              <Database className="mr-2 h-4 w-4" />
              View Metrics
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem className="cursor-pointer">
              <AlertTriangle className="mr-2 h-4 w-4" />
              View Issues
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    ),
  },
]

export function SchemaDetailsModal({ viewingSchema, onClose, schemaData }: SchemaDetailsModalProps) {
  return (
    <Dialog open={!!viewingSchema} onOpenChange={onClose}>
      <DialogContent className="w-[90vw] max-w-3xl md:w-[60vw] md:max-w-4xl px-8 py-6 max-h-[80vh] overflow-hidden">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Database className="h-5 w-5" />
            Schema Details: {viewingSchema?.id}
          </DialogTitle>
          <DialogDescription>
            Detailed information about this schema variant including all attributes and producers
          </DialogDescription>
        </DialogHeader>
        <div className="overflow-y-auto max-h-[60vh] mt-4">
          {viewingSchema && (
            <Tabs defaultValue="schema" className="w-full h-full">
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="schema">Schema Definition</TabsTrigger>
                <TabsTrigger value="producers" className="flex items-center gap-2">
                  Producers
                  <Badge variant="secondary" className="ml-1">
                    {viewingSchema.producers.length}
                  </Badge>
                </TabsTrigger>
              </TabsList>

              <TabsContent value="schema" className="mt-4">
                <SchemaDefinitionView schemaData={schemaData} />
              </TabsContent>

              <TabsContent value="producers" className="mt-4 space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-lg font-semibold flex items-center gap-2">
                      <Users className="h-5 w-5 text-primary" />
                      Telemetry Producers
                    </h3>
                    <p className="text-sm text-muted-foreground">Services currently producing this schema variant</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="relative">
                      <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                      <Input placeholder="Search producers..." className="pl-9 w-64" />
                    </div>
                    <Popover>
                      <PopoverTrigger asChild>
                        <Button variant="outline" size="sm">
                          <SlidersHorizontal className="h-4 w-4 mr-2" />
                          Filter
                        </Button>
                      </PopoverTrigger>
                      <PopoverContent className="w-80" align="end" sideOffset={5}>
                        <div className="space-y-4">
                          <div className="flex items-center justify-between">
                            <h4 className="font-medium text-sm">Filter Producers</h4>
                          </div>

                          <div className="space-y-4">
                            {/* Team Filter */}
                            <div className="space-y-2">
                              <Label className="text-xs font-medium text-foreground">Team</Label>
                              <Select>
                                <SelectTrigger className="h-8">
                                  <SelectValue placeholder="All Teams" />
                                </SelectTrigger>
                                <SelectContent>
                                  <SelectItem value="all">All Teams</SelectItem>
                                  <SelectItem value="api-team">API Team</SelectItem>
                                  <SelectItem value="user-team">User Team</SelectItem>
                                  <SelectItem value="order-team">Order Team</SelectItem>
                                  <SelectItem value="payment-team">Payment Team</SelectItem>
                                  <SelectItem value="notification-team">Notification Team</SelectItem>
                                </SelectContent>
                              </Select>
                            </div>

                            {/* Environment Filter */}
                            <div className="space-y-2">
                              <Label className="text-xs font-medium text-foreground">Environment</Label>
                              <Select>
                                <SelectTrigger className="h-8">
                                  <SelectValue placeholder="All Environments" />
                                </SelectTrigger>
                                <SelectContent>
                                  <SelectItem value="all">All Environments</SelectItem>
                                  <SelectItem value="production">Production</SelectItem>
                                  <SelectItem value="staging">Staging</SelectItem>
                                  <SelectItem value="development">Development</SelectItem>
                                </SelectContent>
                              </Select>
                            </div>

                            {/* Health Status Filter */}
                            <div className="space-y-2">
                              <Label className="text-xs font-medium text-foreground">Health Status</Label>
                              <Select>
                                <SelectTrigger className="h-8">
                                  <SelectValue placeholder="All Statuses" />
                                </SelectTrigger>
                                <SelectContent>
                                  <SelectItem value="all">All Statuses</SelectItem>
                                  <SelectItem value="healthy">Healthy</SelectItem>
                                  <SelectItem value="warning">Warning</SelectItem>
                                  <SelectItem value="critical">Critical</SelectItem>
                                </SelectContent>
                              </Select>
                            </div>
                          </div>

                          <div className="flex items-center justify-end gap-2 pt-2 border-t">
                            <Button variant="outline" size="sm" className="h-7 px-3 text-xs">
                              Clear All
                            </Button>
                            <Button size="sm" className="h-7 px-3 text-xs">
                              Apply Filters
                            </Button>
                          </div>
                        </div>
                      </PopoverContent>
                    </Popover>
                  </div>
                </div>

                <DataTable
                  columns={producerColumns}
                  data={viewingSchema.producers}
                  currentPage={1}
                  pageSize={10}
                  onPageChange={() => {}}
                  onPageSizeChange={() => {}}
                  totalCount={viewingSchema.producers.length}
                  showColumnVisibility={false}
                  summaryText={`Showing ${viewingSchema.producers.length} producer${viewingSchema.producers.length !== 1 ? "s" : ""}`}
                />

                {/* Summary Stats */}
                <div className="grid grid-cols-4 gap-4 pt-4 border-t">
                  <div className="text-center p-3 rounded-md bg-muted/50">
                    <div className="text-lg font-semibold">{viewingSchema.producers.length}</div>
                    <div className="text-xs text-muted-foreground">Total Services</div>
                  </div>
                  <div className="text-center p-3 rounded-md bg-muted/50">
                    <div className="text-lg font-semibold">
                      {Array.from(new Set(viewingSchema.producers.map((p) => p.team))).length}
                    </div>
                    <div className="text-xs text-muted-foreground">Teams</div>
                  </div>
                  <div className="text-center p-3 rounded-md bg-muted/50">
                    <div className="text-lg font-semibold">
                      {Array.from(new Set(viewingSchema.producers.map((p) => p.environment))).length}
                    </div>
                    <div className="text-xs text-muted-foreground">Environments</div>
                  </div>
                  <div className="text-center p-3 rounded-md bg-muted/50">
                    <div className="text-lg font-semibold text-green-500">{viewingSchema.producers.length}</div>
                    <div className="text-xs text-muted-foreground">Healthy</div>
                  </div>
                </div>
              </TabsContent>
            </Tabs>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
} 