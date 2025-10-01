'use client'

import { useState, useEffect } from 'react'
import { X, Search, Package } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { TelemetryScopesTable } from './telemetry-scopes-table'
import { useTelemetryScopes } from '@/hooks/schema'
import type { TelemetryScope } from '@/types/telemetry'

interface TelemetryScopesPanelProps {
  telemetryKey: string
  isOpen: boolean
  onClose: () => void
}

export function TelemetryScopesPanel({
  telemetryKey,
  isOpen,
  onClose,
}: TelemetryScopesPanelProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [filteredScopes, setFilteredScopes] = useState<TelemetryScope[]>([])

  // Fetch scopes data for this specific telemetry
  const { data: scopesData, isLoading } = useTelemetryScopes(telemetryKey)

  useEffect(() => {
    if (scopesData?.items) {
      setFilteredScopes(scopesData.items)
    }
  }, [scopesData])

  // Reset state when panel is opened
  useEffect(() => {
    if (isOpen) {
      setSearchQuery('')
    }
  }, [isOpen])

  useEffect(() => {
    if (scopesData?.items) {
      if (!searchQuery) {
        // Show all scopes when no search query
        setFilteredScopes(scopesData.items)
        return
      }

      const query = searchQuery.toLowerCase()
      const filtered = scopesData.items.filter(scope => {
        const scopeName = scope.name.toLowerCase()
        const scopeVersion = scope.version.toLowerCase()
        const scopeSchemaURL = scope.schemaURL.toLowerCase()
        const attributeValues = Object.values(scope.attributes)
          .map(val => String(val).toLowerCase())
          .join(' ')

        return scopeName.includes(query) || 
               scopeVersion.includes(query) || 
               scopeSchemaURL.includes(query) ||
               attributeValues.includes(query)
      })
      setFilteredScopes(filtered)
    }
  }, [searchQuery, scopesData])

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
              <Package className="h-5 w-5 text-blue-500" />
              <h2 className="text-lg font-semibold">Telemetry Scopes</h2>
              <Badge variant="secondary" className="ml-1">
                {filteredScopes.length || 0}
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
                placeholder="Search scopes..."
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
                  {filteredScopes.length || 0} instrumentation scopes generating{' '}
                  <span className="font-mono font-medium text-foreground">
                    {telemetryKey}
                  </span>{' '}
                  telemetry
                </div>

                {isLoading ? (
                  <div className="text-center py-12">
                    <Package className="h-12 w-12 text-muted-foreground mx-auto mb-4 animate-pulse" />
                    <h3 className="text-lg font-medium mb-2">
                      Loading scopes...
                    </h3>
                    <p className="text-muted-foreground">
                      Please wait while we fetch the scope data.
                    </p>
                  </div>
                ) : filteredScopes.length > 0 ? (
                  <TelemetryScopesTable scopes={filteredScopes} />
                ) : (
                  <div className="text-center py-12">
                    <Package className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                    <h3 className="text-lg font-medium mb-2">
                      No scopes found
                    </h3>
                    <p className="text-muted-foreground">
                      {searchQuery
                        ? 'Try adjusting your search terms.'
                        : 'No telemetry scopes are available for this schema.'}
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
