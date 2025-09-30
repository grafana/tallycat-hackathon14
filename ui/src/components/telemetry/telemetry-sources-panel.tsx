'use client'

import { useState, useEffect } from 'react'
import { X, Search, Server } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import type { Telemetry, TelemetryEntity } from '@/types/telemetry'
import { TelemetryEntitiesTable } from './telemetry-entities-table'

interface TelemetryEntitiesPanelProps {
  schemaData: Telemetry | null
  isOpen: boolean
  onClose: () => void
}

export function TelemetryEntitiesPanel({
  schemaData,
  isOpen,
  onClose,
}: TelemetryEntitiesPanelProps) {
  // State for the sources panel
  const [searchQuery, setSearchQuery] = useState('')
  const [filteredEntities, setFilteredEntities] = useState<Record<string, TelemetryEntity>>({})

  useEffect(() => {
    if (schemaData?.entities) {
      setFilteredEntities(schemaData.entities)
    }
  }, [schemaData])

  // Reset state when panel is opened
  useEffect(() => {
    if (isOpen) {
      setSearchQuery('')
    }
  }, [isOpen])

  useEffect(() => {
    if (schemaData?.entities) {
      if (!searchQuery) {
        setFilteredEntities(schemaData.entities)
        return
      }

      const query = searchQuery.toLowerCase()
      const filtered = Object.fromEntries(
        Object.entries(schemaData.entities).filter(([_, entity]) => {
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
  }, [searchQuery, schemaData])

  if (!isOpen) return null

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 bg-background/10 backdrop-blur-sm"
      onClick={handleBackdropClick}
    >
      <div className="fixed inset-y-0 right-0 w-full max-w-4xl border-l bg-background shadow-lg">
        <div className="flex h-full flex-col">
          {/* Header */}
          <div className="flex items-center justify-between border-b px-6 py-4">
            <div className="flex items-center gap-3">
              <Server className="h-5 w-5 text-indigo-500" />
              <h2 className="text-lg font-semibold">Telemetry Entities</h2>
              <Badge variant="secondary" className="ml-1">
                {Object.keys(filteredEntities).length || 0}
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
                placeholder="Search entities..."
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
                  {Object.keys(filteredEntities).length || 0} entities generating{' '}
                  <span className="font-mono font-medium text-foreground">
                    {schemaData?.schemaKey}
                  </span>{' '}
                  telemetry
                </div>

                <TelemetryEntitiesTable entities={filteredEntities} />

                {Object.keys(filteredEntities).length === 0 && (
                  <div className="text-center py-12">
                    <Server className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                    <h3 className="text-lg font-medium mb-2">
                      No entities found
                    </h3>
                    <p className="text-muted-foreground">
                      {searchQuery
                        ? 'Try adjusting your search terms.'
                        : 'No telemetry entities are available.'}
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
