'use client'

import { useState } from 'react'
import { Server, Database, BarChart3, ChevronDown, Search } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import type { Attribute } from '@/types/telemetry'

interface AttributesViewProps {
  attributes: Attribute[]
}

export function AttributesView({ attributes }: AttributesViewProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [expandedSections, setExpandedSections] = useState({
    resource: false,
    scope: false,
    data: false,
  })

  // Filter attributes based on search query
  const filteredAttributes = attributes.filter((attr) => {
    if (!searchQuery) return true
    return (
      attr.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      attr.type.toLowerCase().includes(searchQuery.toLowerCase())
    )
  })

  // Group attributes by backend source values
  const resourceAttributes = filteredAttributes.filter(
    (attr) => attr.source === 'Resource'
  )
  const scopeAttributes = filteredAttributes.filter(
    (attr) => attr.source === 'Scope'
  )
  const dataPointAttributes = filteredAttributes.filter(
    (attr) => attr.source === 'DataPoint'
  )

  const toggleSection = (section: keyof typeof expandedSections) => {
    setExpandedSections((prev) => ({
      ...prev,
      [section]: !prev[section],
    }))
  }

  const renderAttributesTable = (attributes: Attribute[], showRequired = false) => (
    <div className="rounded-md border">
      <ScrollArea className="h-[200px]">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[250px]">Name</TableHead>
              <TableHead className="w-[100px]">Type</TableHead>
              {showRequired && <TableHead className="w-[100px]">Required</TableHead>}
            </TableRow>
          </TableHeader>
          <TableBody>
            {attributes.length === 0 ? (
              <TableRow>
                <TableCell colSpan={showRequired ? 3 : 2} className="text-center text-muted-foreground">
                  No attributes found.
                </TableCell>
              </TableRow>
            ) : (
              attributes.map((attr, index) => (
                <TableRow key={index}>
                  <TableCell className="font-medium font-mono">
                    {attr.name}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className="font-mono text-xs">
                      {attr.type}
                    </Badge>
                  </TableCell>
                  {showRequired && (
                    <TableCell>
                      {/* Only service.name is required for resource attributes, others optional */}
                      {attr.name === 'service.name' ? (
                        <Badge variant="outline" className="bg-blue-500/10 text-blue-500 border-blue-500/20">
                          Required
                        </Badge>
                      ) : (
                        <Badge variant="outline" className="bg-gray-500/10 text-gray-500 border-gray-500/20">
                          Optional
                        </Badge>
                      )}
                    </TableCell>
                  )}
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </ScrollArea>
    </div>
  )

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h3 className="text-lg font-medium">Schema Definition</h3>
        <p className="text-sm text-muted-foreground">
          All attributes defined in this schema variant
        </p>
      </div>

      {/* Search */}
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
        <div className="flex items-center gap-2">
          <h3 className="text-base font-medium">Schema Attributes</h3>
        </div>
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search attributes..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full sm:w-[250px] pl-9"
          />
        </div>
      </div>

      {/* Collapsible Sections */}
      <div className="space-y-4">
        {/* Resource Section */}
        <Collapsible
          open={expandedSections.resource}
          onOpenChange={() => toggleSection('resource')}
          className="border rounded-lg overflow-hidden"
        >
          <CollapsibleTrigger className="flex items-center justify-between w-full p-4 bg-blue-50 hover:bg-blue-100 dark:bg-blue-950/10 dark:hover:bg-blue-950/20 text-left">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-md bg-blue-100 dark:bg-blue-900/30">
                <Server className="h-5 w-5 text-blue-500" />
              </div>
              <div>
                <h4 className="font-medium">Resource</h4>
                <p className="text-sm text-muted-foreground">
                  Describes the entity producing the telemetry
                </p>
              </div>
            </div>
            <ChevronDown
              className={`h-5 w-5 text-muted-foreground transition-transform ${
                expandedSections.resource ? 'rotate-180' : ''
              }`}
            />
          </CollapsibleTrigger>
          <CollapsibleContent>
            <div className="p-4 border-t">
              {renderAttributesTable(resourceAttributes, true)}
            </div>
          </CollapsibleContent>
        </Collapsible>

        {/* Scope Section */}
        <Collapsible
          open={expandedSections.scope}
          onOpenChange={() => toggleSection('scope')}
          className="border rounded-lg overflow-hidden"
        >
          <CollapsibleTrigger className="flex items-center justify-between w-full p-4 bg-green-50 hover:bg-green-100 dark:bg-green-950/10 dark:hover:bg-green-950/20 text-left">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-md bg-green-100 dark:bg-green-900/30">
                <Database className="h-5 w-5 text-green-500" />
              </div>
              <div>
                <h4 className="font-medium">Instrumentation Scope</h4>
                <p className="text-sm text-muted-foreground">
                  Identifies the library that created the telemetry
                </p>
              </div>
            </div>
            <ChevronDown
              className={`h-5 w-5 text-muted-foreground transition-transform ${
                expandedSections.scope ? 'rotate-180' : ''
              }`}
            />
          </CollapsibleTrigger>
          <CollapsibleContent>
            <div className="p-4 border-t">
              {renderAttributesTable(scopeAttributes)}
            </div>
          </CollapsibleContent>
        </Collapsible>

        {/* Data Point Section */}
        <Collapsible
          open={expandedSections.data}
          onOpenChange={() => toggleSection('data')}
          className="border rounded-lg overflow-hidden"
        >
          <CollapsibleTrigger className="flex items-center justify-between w-full p-4 bg-purple-50 hover:bg-purple-100 dark:bg-purple-950/10 dark:hover:bg-purple-950/20 text-left">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-md bg-purple-100 dark:bg-purple-900/30">
                <BarChart3 className="h-5 w-5 text-purple-500" />
              </div>
              <div>
                <h4 className="font-medium">Data Point</h4>
                <p className="text-sm text-muted-foreground">
                  Contains the actual data points for this schema
                </p>
              </div>
            </div>
            <ChevronDown
              className={`h-5 w-5 text-muted-foreground transition-transform ${
                expandedSections.data ? 'rotate-180' : ''
              }`}
            />
          </CollapsibleTrigger>
          <CollapsibleContent>
            <div className="p-4 border-t">
              {renderAttributesTable(dataPointAttributes)}
            </div>
          </CollapsibleContent>
        </Collapsible>
      </div>
    </div>
  )
} 