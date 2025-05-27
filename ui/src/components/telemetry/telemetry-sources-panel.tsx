"use client"

import { useState, useEffect } from "react"
import { X, Search, Server, Clock, Tag, Hash } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import type { Telemetry, TelemetryProducer } from "@/types/telemetry"

interface TelemetryProducersPanelProps {
  schemaData: Telemetry | null
  isOpen: boolean
  onClose: () => void
}

export function TelemetryProducersPanel({ schemaData, isOpen, onClose }: TelemetryProducersPanelProps) {
  // State for the sources panel
  const [searchQuery, setSearchQuery] = useState("")
  const [filteredSources, setFilteredSources] = useState<any[]>([])

  // Compact date format function
  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return new Intl.DateTimeFormat("en-US", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      hour12: false,
    }).format(date)
  }

  useEffect(() => {
    if (schemaData?.producers) {
      // Add mock data for namespace, firstSeen, and lastSeen
      const sourcesWithMockData = Object.values(schemaData.producers).map((source: TelemetryProducer) => ({
        ...source,
        namespace: source.namespace,
        firstSeen: source.firstSeen,
        lastSeen: source.lastSeen,
      }))

      setFilteredSources(sourcesWithMockData)
    }
  }, [schemaData])

  // Reset state when panel is opened
  useEffect(() => {
    if (isOpen) {
      setSearchQuery("")
    }
  }, [isOpen])

  useEffect(() => {
    if (schemaData?.producers) {
      const sourcesWithMockData = Object.values(schemaData.producers).map((source: TelemetryProducer) => ({
        ...source,
        namespace: source.namespace || '',
        firstSeen: source.firstSeen,
        lastSeen: source.lastSeen,
      }))

      const newFilteredSources = sourcesWithMockData?.filter((source: any) => {
        if (!searchQuery) return true;
        
        const name = source.name?.toLowerCase() || '';
        const namespace = source.namespace?.toLowerCase() || '';
        const query = searchQuery.toLowerCase();
        
        return name.includes(query) || namespace.includes(query);
      })
      setFilteredSources(newFilteredSources)
    }
  }, [searchQuery, schemaData])

  if (!isOpen) return null

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }

  return (
    <div className="fixed inset-0 z-50 bg-background/10 backdrop-blur-sm" onClick={handleBackdropClick}>
      <div className="fixed inset-y-0 right-0 w-full max-w-4xl border-l bg-background shadow-lg">
        <div className="flex h-full flex-col">
          {/* Header */}
          <div className="flex items-center justify-between border-b px-6 py-4">
            <div className="flex items-center gap-3">
              <Server className="h-5 w-5 text-indigo-500" />
              <h2 className="text-lg font-semibold">Telemetry Producers</h2>
              <Badge variant="secondary" className="ml-1">
                {filteredSources?.length || 0}
              </Badge>
            </div>
            <Button variant="ghost" size="icon" onClick={onClose}>
              <X className="h-4 w-4" />
              <span className="sr-only">Close</span>
            </Button>
          </div>

          {/* Search */}
          <div className="border-b px-6 py-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Search producers..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>
          </div>

          {/* Content */}
          <div className="flex-1 overflow-hidden">
            <ScrollArea className="h-full">
              <div className="px-6 py-4">
                <div className="text-sm text-muted-foreground mb-6">
                  {filteredSources?.length || 0} producers generating{" "}
                  <span className="font-mono font-medium text-foreground">{schemaData?.schemaKey}</span> telemetry
                </div>

                <div className="rounded-lg border">
                  <Table>
                    <TableHeader>
                      <TableRow className="hover:bg-transparent">
                        <TableHead className="w-[240px] font-semibold">Name</TableHead>
                        <TableHead className="w-[140px] font-semibold">Namespace</TableHead>
                        <TableHead className="w-[120px] font-semibold">Version</TableHead>
                        <TableHead className="w-[140px] font-semibold">First Seen</TableHead>
                        <TableHead className="w-[140px] font-semibold">Last Seen</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {filteredSources?.map((source: any) => (
                        <TableRow key={`${source.name}-${source.namespace}-${source.version}`} className="hover:bg-muted/50">
                          <TableCell className="py-4">
                            <div className="flex items-center gap-3">
                              <Server className="h-4 w-4 text-indigo-500 flex-shrink-0" />
                              <span className="font-medium">{source.name || 'N/A'}</span>
                            </div>
                          </TableCell>
                          <TableCell className="py-4">
                            <div className="flex items-center gap-2">
                              <span className="font-mono text-sm whitespace-nowrap">{source.namespace || 'N/A'}</span>
                            </div>
                          </TableCell>
                          <TableCell className="py-4">
                            <div className="flex items-center gap-2">
                              <Tag className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
                              <Badge variant="outline" className="font-mono">
                                {source.version ? `v${source.version}` : 'N/A'}
                              </Badge>
                            </div>
                          </TableCell>
                          <TableCell className="py-4">
                            <div className="flex items-center gap-2 text-sm">
                              <Clock className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
                              <span className="whitespace-nowrap font-mono">{formatDate(source.firstSeen)}</span>
                            </div>
                          </TableCell>
                          <TableCell className="py-4">
                            <div className="flex items-center gap-2 text-sm">
                              <Clock className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
                              <span className="whitespace-nowrap font-mono">{formatDate(source.lastSeen)}</span>
                            </div>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>

                {filteredSources?.length === 0 && (
                  <div className="text-center py-12">
                    <Server className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                    <h3 className="text-lg font-medium mb-2">No producers found</h3>
                    <p className="text-muted-foreground">
                      {searchQuery ? "Try adjusting your search terms." : "No telemetry producers are available."}
                    </p>
                  </div>
                )}
              </div>
            </ScrollArea>
          </div>
        </div>
      </div>
    </div>
  )
}
