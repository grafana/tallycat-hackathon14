"use client"

import { useState } from "react"
import { Server, Database, BarChart3, ChevronDown, Search } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import type { SchemaField, GetSchemaResponse } from "@/types/telemetry"

interface SchemaDefinitionViewProps {
  schemaData: GetSchemaResponse
}

export function SchemaDefinitionView({ schemaData }: SchemaDefinitionViewProps) {
  const [searchQuery, setSearchQuery] = useState("")
  const [attributeFilter, setAttributeFilter] = useState("all")
  // Changed initial state to false for all sections
  const [expandedSections, setExpandedSections] = useState<Record<string, boolean>>({
    resource: false,
    scope: false,
    data: false,
  })

  // Get resource attributes based on telemetry type
  const getResourceAttributes = () => {
    if (!schemaData.schema) {
      return {}
    }

    return schemaData.schema.reduce((acc: Record<string, string>, field: SchemaField) => {
      if (field.source === "resource") {
        acc[field.name] = field.description || ""
      }
      return acc
    }, {} as Record<string, string>)
  }

  // Get scope attributes
  const getScopeAttributes = () => {
    if (!schemaData.schema) {
      return {}
    }

    return schemaData.schema.reduce((acc: Record<string, string>, field: SchemaField) => {
      if (field.source === "scope") {
        acc[field.name] = field.description || ""
      }
      return acc
    }, {} as Record<string, string>)
  }

  // Get data attributes based on telemetry type
  const getDataAttributes = () => {
    if (!schemaData.schema) {
      return []
    }

    return schemaData.schema
      .filter((field) => field.source === "datapoint" || field.source === "log" || field.source === "span")
      .map((field) => ({
        name: field.name,
        type: field.type,
        description: field.description || "",
        required: field.required,
        value: field.description || "",
      }))
  }

  // Filter attributes based on search query and filter
  const filterAttributes = (attributes: any[]) => {
    return attributes.filter((attr) => {
      const matchesSearch =
        searchQuery === "" ||
        attr.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        attr.description?.toLowerCase().includes(searchQuery.toLowerCase())

      if (attributeFilter === "all") return matchesSearch
      if (attributeFilter === "required") return matchesSearch && attr.required
      if (attributeFilter === "optional") return matchesSearch && !attr.required

      return matchesSearch
    })
  }

  const toggleSection = (section: string) => {
    setExpandedSections((prev) => ({
      ...prev,
      [section]: !prev[section],
    }))
  }

  // Get data section name based on telemetry type
  const getDataSectionName = () => {
    switch (schemaData.type) {
      case "metric":
        return "Metric Data"
      case "log":
        return "Log Record"
      case "trace":
        return "Span"
      default:
        return "Data"
    }
  }

  // Get data section icon based on telemetry type
  const getDataSectionIcon = () => {
    switch (schemaData.type) {
      case "metric":
        return <BarChart3 className="h-5 w-5 text-purple-500" />
      case "log":
        return <BarChart3 className="h-5 w-5 text-purple-500" />
      case "trace":
        return <BarChart3 className="h-5 w-5 text-purple-500" />
      default:
        return null
    }
  }

  return (
    <div className="space-y-6">
      {/* Schema Definition Header */}
      <div>
        <h3 className="text-lg font-medium">Schema Definition</h3>
        <p className="text-sm text-muted-foreground">
          OpenTelemetry {schemaData.type} schema for {schemaData.name}
        </p>
      </div>

      {/* Search and Filter */}
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
        <div className="flex items-center gap-2">
          <h3 className="text-base font-medium">
            OpenTelemetry {schemaData.type.charAt(0).toUpperCase() + schemaData.type.slice(1)} Schema
          </h3>
        </div>
        <div className="flex flex-col sm:flex-row gap-2 w-full sm:w-auto">
          <div className="relative">
            <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search attributes..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full sm:w-[250px] pl-9"
            />
          </div>
          <Select value={attributeFilter} onValueChange={setAttributeFilter}>
            <SelectTrigger className="w-full sm:w-[150px]">
              <SelectValue placeholder="Filter attributes" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Attributes</SelectItem>
              <SelectItem value="required">Required Only</SelectItem>
              <SelectItem value="optional">Optional Only</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Collapsible sections for Resource, Instrumentation Scope, and Data */}
      <div className="space-y-4">
        {/* Resource Section */}
        <Collapsible
          open={expandedSections.resource}
          onOpenChange={() => toggleSection("resource")}
          className="border rounded-lg overflow-hidden"
        >
          {/* Modified the trigger to align content to the left */}
          <CollapsibleTrigger className="flex items-center justify-between w-full p-4 bg-blue-50 hover:bg-blue-100 dark:bg-blue-950/10 dark:hover:bg-blue-950/20 text-left">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-md bg-blue-100 dark:bg-blue-900/30">
                <Server className="h-5 w-5 text-blue-500" />
              </div>
              <div>
                <h4 className="font-medium">Resource</h4>
                <p className="text-sm text-muted-foreground">Describes the entity producing the telemetry</p>
              </div>
            </div>
            <ChevronDown
              className={`h-5 w-5 text-muted-foreground transition-transform ${
                expandedSections.resource ? "rotate-180" : ""
              }`}
            />
          </CollapsibleTrigger>

          <CollapsibleContent>
            <div className="p-4 border-t">
              <ScrollArea className="h-[300px]">
                <div className="rounded-md border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-[250px]">Name</TableHead>
                        <TableHead className="w-[100px]">Type</TableHead>
                        <TableHead className="w-[100px]">Required</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {Object.entries(getResourceAttributes())
                        .filter(([key]) => searchQuery === "" || key.toLowerCase().includes(searchQuery.toLowerCase()))
                        .filter(([key]) => {
                          if (attributeFilter === "all") return true
                          if (attributeFilter === "required") return key === "service.name"
                          if (attributeFilter === "optional") return key !== "service.name"
                          return true
                        })
                        .map(([key], index) => (
                          <TableRow key={index}>
                            <TableCell className="font-medium font-mono">{key}</TableCell>
                            <TableCell>
                              <Badge variant="outline" className="font-mono text-xs">
                                string
                              </Badge>
                            </TableCell>
                            <TableCell>
                              {key === "service.name" ? (
                                <Badge variant="outline" className="bg-blue-500/10 text-blue-500 border-blue-500/20">
                                  Required
                                </Badge>
                              ) : (
                                <Badge variant="outline" className="bg-gray-500/10 text-gray-500 border-gray-500/20">
                                  Optional
                                </Badge>
                              )}
                            </TableCell>
                          </TableRow>
                        ))}
                    </TableBody>
                  </Table>
                </div>
              </ScrollArea>
            </div>
          </CollapsibleContent>
        </Collapsible>

        {/* Instrumentation Scope Section */}
        <Collapsible
          open={expandedSections.scope}
          onOpenChange={() => toggleSection("scope")}
          className="border rounded-lg overflow-hidden"
        >
          {/* Modified the trigger to align content to the left */}
          <CollapsibleTrigger className="flex items-center justify-between w-full p-4 bg-green-50 hover:bg-green-100 dark:bg-green-950/10 dark:hover:bg-green-950/20 text-left">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-md bg-green-100 dark:bg-green-900/30">
                <Database className="h-5 w-5 text-green-500" />
              </div>
              <div>
                <h4 className="font-medium">Instrumentation Scope</h4>
                <p className="text-sm text-muted-foreground">Identifies the library that created the telemetry</p>
              </div>
            </div>
            <ChevronDown
              className={`h-5 w-5 text-muted-foreground transition-transform ${
                expandedSections.scope ? "rotate-180" : ""
              }`}
            />
          </CollapsibleTrigger>

          <CollapsibleContent>
            <div className="p-4 border-t">
              <ScrollArea className="h-[300px]">
                <div className="rounded-md border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="">Name</TableHead>
                        <TableHead className="">Type</TableHead>
                        <TableHead className="">Required</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      <TableRow>
                        <TableCell className="font-medium font-mono">name</TableCell>
                        <TableCell>
                          <Badge variant="outline" className="font-mono text-xs">
                            string
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline" className="bg-blue-500/10 text-blue-500 border-blue-500/20">
                            Required
                          </Badge>
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell className="font-medium font-mono">version</TableCell>
                        <TableCell>
                          <Badge variant="outline" className="font-mono text-xs">
                            string
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline" className="bg-gray-500/10 text-gray-500 border-gray-500/20">
                            Optional
                          </Badge>
                        </TableCell>
                      </TableRow>
                      {Object.entries(getScopeAttributes())
                        .filter(([key]) => searchQuery === "" || key.toLowerCase().includes(searchQuery.toLowerCase()))
                        .filter(() => {
                          if (attributeFilter === "required") return false
                          if (attributeFilter === "optional") return true
                          return true
                        })
                        .map(([key, value], index) => (
                          <TableRow key={index}>
                            <TableCell className="font-medium font-mono">attributes.{key}</TableCell>
                            <TableCell>{String(value)}</TableCell>
                            <TableCell>
                              <Badge variant="outline" className="font-mono text-xs">
                                string
                              </Badge>
                            </TableCell>
                            <TableCell>
                              <Badge variant="outline" className="bg-gray-500/10 text-gray-500 border-gray-500/20">
                                Optional
                              </Badge>
                            </TableCell>
                          </TableRow>
                        ))}
                    </TableBody>
                  </Table>
                </div>
              </ScrollArea>
            </div>
          </CollapsibleContent>
        </Collapsible>

        {/* Data Section (Metric Data, Log Record, or Span) */}
        <Collapsible
          open={expandedSections.data}
          onOpenChange={() => toggleSection("data")}
          className="border rounded-lg overflow-hidden"
        >
          {/* Modified the trigger to align content to the left */}
          <CollapsibleTrigger className="flex items-center justify-between w-full p-4 bg-purple-50 hover:bg-purple-100 dark:bg-purple-950/10 dark:hover:bg-purple-950/20 text-left">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-md bg-purple-100 dark:bg-purple-900/30">
                {getDataSectionIcon()}
              </div>
              <div>
                <h4 className="font-medium">{getDataSectionName()}</h4>
                <p className="text-sm text-muted-foreground">Contains the actual {schemaData.type} data</p>
              </div>
            </div>
            <ChevronDown
              className={`h-5 w-5 text-muted-foreground transition-transform ${
                expandedSections.data ? "rotate-180" : ""
              }`}
            />
          </CollapsibleTrigger>

          <CollapsibleContent>
            <div className="p-4 border-t">
              <ScrollArea className="h-[300px]">
                <div className="rounded-md border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-[250px]">Name</TableHead>
                        <TableHead>Value/Description</TableHead>
                        <TableHead className="w-[100px]">Type</TableHead>
                        <TableHead className="w-[100px]">Required</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {filterAttributes(getDataAttributes()).map((attr, index) => (
                        <TableRow key={index}>
                          <TableCell className="font-medium font-mono">{attr.name}</TableCell>
                          <TableCell>{attr.value || attr.description}</TableCell>
                          <TableCell>
                            <Badge variant="outline" className="font-mono text-xs">
                              {attr.type}
                            </Badge>
                          </TableCell>
                          <TableCell>
                            {attr.required ? (
                              <Badge variant="outline" className="bg-blue-500/10 text-blue-500 border-blue-500/20">
                                Required
                              </Badge>
                            ) : (
                              <Badge variant="outline" className="bg-gray-500/10 text-gray-500 border-gray-500/20">
                                Optional
                              </Badge>
                            )}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              </ScrollArea>
            </div>
          </CollapsibleContent>
        </Collapsible>
      </div>
    </div>
  )
}
