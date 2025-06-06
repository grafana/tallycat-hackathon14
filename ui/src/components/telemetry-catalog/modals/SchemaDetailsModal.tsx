'use client'

import { useState, useEffect } from 'react'
import {
  Database,
  Users,
  Search,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { AttributesView } from '@/components/telemetry-catalog/features/schema-definition/AttributesView'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type { TelemetrySchema } from '@/types/telemetry-schema'
import type { Telemetry } from '@/types/telemetry'
import { TelemetryProducersTable } from '@/components/telemetry/telemetry-producers-table'
import { WeaverDefinition } from '@/components/telemetry-catalog/features/weaver-definition/WeaverDefinition'

interface SchemaDetailsModalProps {
  viewingSchema: TelemetrySchema | null
  onClose: () => void
  telemetry: Telemetry
  isLoading?: boolean
}

export function SchemaDetailsModal({
  viewingSchema,
  onClose,
  telemetry,
  isLoading = false,
}: SchemaDetailsModalProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [filteredProducers, setFilteredProducers] = useState(viewingSchema?.producers || [])

  // Update filtered producers when search query or viewingSchema changes
  useEffect(() => {
    if (viewingSchema?.producers) {
      const filtered = viewingSchema.producers.filter((producer) => {
        if (!searchQuery) return true

        const name = producer.name?.toLowerCase() || ''
        const namespace = producer.namespace?.toLowerCase() || ''
        const query = searchQuery.toLowerCase()

        return name.includes(query) || namespace.includes(query)
      })
      setFilteredProducers(filtered)
    }
  }, [searchQuery, viewingSchema])

  return (
    <Dialog open={!!viewingSchema} onOpenChange={onClose}>
      <DialogContent className="w-[90vw] max-w-3xl md:w-[60vw] md:max-w-4xl px-8 py-6 max-h-[80vh] overflow-hidden">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Database className="h-5 w-5" />
            Schema Details: {viewingSchema?.id}
          </DialogTitle>
          <DialogDescription>
            Detailed information about this schema variant including all
            attributes and producers
          </DialogDescription>
        </DialogHeader>
        <div className="overflow-y-auto max-h-[60vh] mt-4">
          {isLoading ? (
            <div className="flex items-center justify-center h-32">
              <p className="text-muted-foreground">Loading schema details...</p>
            </div>
          ) : viewingSchema ? (
            <Tabs defaultValue="schema" className="w-full h-full">
              <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="schema">Schema Definition</TabsTrigger>
                <TabsTrigger
                  value="producers"
                  className="flex items-center gap-2"
                >
                  Producers
                  <Badge variant="secondary" className="ml-1">
                    {filteredProducers.length}
                  </Badge>
                </TabsTrigger>
                <TabsTrigger value="weaver">Weaver Definition</TabsTrigger>
              </TabsList>

              <TabsContent value="schema" className="mt-4">
                <AttributesView attributes={viewingSchema.attributes} />
              </TabsContent>

              <TabsContent value="producers" className="mt-4 space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-lg font-semibold flex items-center gap-2">
                      <Users className="h-5 w-5 text-primary" />
                      Telemetry Producers
                    </h3>
                    <p className="text-sm text-muted-foreground">
                      Services currently producing this schema variant
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="relative">
                      <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                      <Input
                        placeholder="Search producers..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="pl-9 w-64"
                      />
                    </div>
                  </div>
                </div>

                <TelemetryProducersTable producers={filteredProducers} />
              </TabsContent>

              <TabsContent value="weaver" className="mt-4">
                <WeaverDefinition telemetry={telemetry} schema={viewingSchema} />
              </TabsContent>

            </Tabs>
          ) : null}
        </div>
      </DialogContent>
    </Dialog>
  )
}
