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
import { TelemetryEntitiesTable } from '@/components/telemetry/telemetry-entities-table'
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
  const [filteredEntities, setFilteredEntities] = useState(viewingSchema?.entities || {})

  // Update filtered entities when search query or viewingSchema changes
  useEffect(() => {
    if (viewingSchema?.entities) {
      if (!searchQuery) {
        setFilteredEntities(viewingSchema.entities)
        return
      }

      const query = searchQuery.toLowerCase()
      const filtered = Object.fromEntries(
        Object.entries(viewingSchema.entities).filter(([_, entity]) => {
          const entityType = entity.type.toLowerCase()
          const entityId = entity.id.toLowerCase()
          const attributeValues = Object.values(entity.attributes)
            .map(val => String(val).toLowerCase())
            .join(' ')

          return entityType.includes(query) || 
                 entityId.includes(query) || 
                 attributeValues.includes(query)
        })
      )
      setFilteredEntities(filtered)
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
            attributes and entities
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
                  value="entities"
                  className="flex items-center gap-2"
                >
                  Entities
                  <Badge variant="secondary" className="ml-1">
                    {Object.keys(filteredEntities).length}
                  </Badge>
                </TabsTrigger>
                <TabsTrigger value="weaver">Weaver Definition</TabsTrigger>
              </TabsList>

              <TabsContent value="schema" className="mt-4">
                <AttributesView attributes={viewingSchema.attributes} telemetry={telemetry} />
              </TabsContent>

              <TabsContent value="entities" className="mt-4 space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-lg font-semibold flex items-center gap-2">
                      <Users className="h-5 w-5 text-primary" />
                      Telemetry Entities
                    </h3>
                    <p className="text-sm text-muted-foreground">
                      Entities currently producing this schema variant
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="relative">
                      <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                      <Input
                        placeholder="Search entities..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="pl-9 w-64"
                      />
                    </div>
                  </div>
                </div>

                <TelemetryEntitiesTable entities={filteredEntities} />
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
