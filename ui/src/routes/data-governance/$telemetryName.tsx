import { Link, createFileRoute, useParams  } from '@tanstack/react-router'
import { ArrowLeft, ChevronRight } from 'lucide-react'
import { useState } from 'react'
import { TelemetryTypeIcon } from '@/components/telemetry/telemetry-icons'
import { Button } from '@/components/ui/button'
import { useTelemetryDetails } from '@/hooks'
import { useTelemetryScopes } from '@/hooks/schema'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardContent } from '@/components/ui/card'
import { TelemetryEntitiesPanel } from '@/components/telemetry/telemetry-sources-panel'
import { TelemetryScopesPanel } from '@/components/telemetry/telemetry-scopes-panel'
import { VersionAssignmentView } from '@/components/telemetry-catalog/VersionAssignment'
import { TelemetryOverviewPanel } from '@/components/telemetry-catalog/features/telemetry/TelemetryOverviewPanel'
import { SchemaHistoryView } from '@/components/telemetry-catalog/features/history/SchemaHistoryView'
// import { SchemaDefinitionView } from '@/components/telemetry-catalog/features/schema-definition/SchemaDefinitionView'

export const TelemetryDetails = () => {
  const { telemetryName } = useParams({
    from: '/data-governance/$telemetryName',
  })
  const [activeTab, setActiveTab] = useState('schemas')
  const {
    data: telemetry,
    isLoading,
    error,
  } = useTelemetryDetails({ telemetryName })
  const [isEntitiesPanelOpen, setIsEntitiesPanelOpen] = useState(false)
  const [isScopesPanelOpen, setIsScopesPanelOpen] = useState(false)

  // Fetch scopes data for this specific telemetry
  const { data: scopesData } = useTelemetryScopes(telemetry?.schemaKey || '')

  const handleViewAllSources = () => {
    setIsEntitiesPanelOpen(true)
  }

  const handleViewAllScopes = () => {
    setIsScopesPanelOpen(true)
  }

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center h-[50vh] gap-4">
        <p className="text-muted-foreground">Loading telemetry details...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center h-[50vh] gap-4">
        <p className="text-destructive">
          Error loading telemetry details. Please try again later.
        </p>
      </div>
    )
  }

  if (!telemetry) {
    return (
      <div className="flex flex-col items-center justify-center h-[50vh] gap-4">
        <h1 className="text-2xl font-medium">Telemetry signal not found</h1>
        <p className="text-muted-foreground">
          The telemetry signal you're looking for doesn't exist or has been
          removed.
        </p>
        <Button asChild>
          <Link to="/data-governance/telemetry-catalog">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Telemetry Catalog
          </Link>
        </Button>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="sm" asChild className="h-8 w-8 p-0">
            <Link to="/data-governance/telemetry-catalog">
              <ArrowLeft className="h-4 w-4" />
              <span className="sr-only">Back</span>
            </Link>
          </Button>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Link
              to="/data-governance/telemetry-catalog"
              className="hover:text-foreground"
            >
              Telemetry Catalog
            </Link>
            <ChevronRight className="h-4 w-4" />
            <span className="font-medium text-foreground">
              {telemetry.schemaKey}
            </span>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-md bg-primary/10">
            {TelemetryTypeIcon({ type: telemetry.telemetryType })}
          </div>
          <div>
            <h1 className="text-2xl font-medium font-mono">
              {telemetry.schemaKey}
            </h1>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <div className="flex items-center gap-1.5">
                <span className="capitalize">
                  {telemetry.telemetryType} ({telemetry.telemetryType})
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <TelemetryOverviewPanel
        telemetry={telemetry}
        scopesCount={scopesData?.total || 0}
        onViewAllSources={handleViewAllSources}
        onViewAllScopes={handleViewAllScopes}
      />

      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="mb-4">
          {/* <TabsTrigger value="attributes">Attributes</TabsTrigger> */}
          <TabsTrigger value="schemas">Schemas</TabsTrigger>
          <TabsTrigger value="history">History</TabsTrigger>
        </TabsList>

        {/* <TabsContent value="attributes" className="mt-0 space-y-4">
          <Card>
            <CardContent className="p-6">
              <SchemaDefinitionView schemaData={telemetry} />
            </CardContent>
          </Card>
        </TabsContent> */}

        <TabsContent value="schemas" className="mt-0 space-y-4">
          <Card>
            <CardContent className="p-6 min-h-[300px]">
              <VersionAssignmentView telemetry={telemetry} />
            </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="history" className="mt-0 space-y-4">
        <Card>
            <CardContent className="p-6 min-h-[300px]">
              <SchemaHistoryView telemetry={telemetry} />
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <TelemetryEntitiesPanel
        schemaData={telemetry}
        isOpen={isEntitiesPanelOpen}
        onClose={() => setIsEntitiesPanelOpen(false)}
      />

      <TelemetryScopesPanel
        telemetryKey={telemetry?.schemaKey || ''}
        isOpen={isScopesPanelOpen}
        onClose={() => setIsScopesPanelOpen(false)}
      />
    </div>
  )
}

export const Route = createFileRoute('/data-governance/$telemetryName')({
  component: TelemetryDetails,
  validateSearch: (search: Record<string, unknown>) => {
    return search
  },
})
