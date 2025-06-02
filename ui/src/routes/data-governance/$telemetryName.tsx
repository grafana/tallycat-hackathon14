import { createFileRoute, Link } from '@tanstack/react-router'
import { useParams } from '@tanstack/react-router'
import { TelemetryTypeIcon } from '@/components/telemetry/telemetry-icons'
import { Button } from '@/components/ui/button'
import { ArrowLeft, ChevronRight, Server } from 'lucide-react'
import { useState } from 'react'
import { useTelemetryDetails } from '@/hooks'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardContent } from '@/components/ui/card'
import { TelemetryProducersPanel } from '@/components/telemetry/telemetry-sources-panel'
import { VersionAssignmentView } from '@/components/telemetry-catalog/VersionAssignment'
import { TelemetryOverviewPanel } from '@/components/telemetry-catalog/features/telemetry/TelemetryOverviewPanel'
import { SchemaDefinitionView } from '@/components/telemetry-catalog/features/schema-definition/SchemaDefinitionView'

export const TelemetryDetails = () => {
  const { telemetryName } = useParams({
    from: '/data-governance/$telemetryName',
  })
  const [activeTab, setActiveTab] = useState('schema')
  const {
    data: telemetry,
    isLoading,
    error,
  } = useTelemetryDetails({ telemetryName })
  const [isProducersPanelOpen, setIsProducersPanelOpen] = useState(false)

  const handleViewAllSources = () => {
    setIsProducersPanelOpen(true)
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

        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
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
          <div className="flex flex-wrap gap-2">
            {/* Sources Badge */}
            <Button
              variant="outline"
              size="sm"
              className="h-7 px-2 text-xs flex items-center gap-1.5"
              onClick={handleViewAllSources}
            >
              <Server className="h-3.5 w-3.5" />
              {Object.keys(telemetry.producers).length} Producers
            </Button>

            {/* View Validation Button */}
            <Button
              variant="outline"
              size="sm"
              className="h-7 px-2 text-xs"
              onClick={() => setActiveTab('validation')}
            >
              View Validation
            </Button>
          </div>
        </div>
      </div>

      <TelemetryOverviewPanel
        telemetry={telemetry}
        onViewAllSources={handleViewAllSources}
      />

      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="mb-4">
          <TabsTrigger value="schema">Schema</TabsTrigger>
          <TabsTrigger value="versions">Versions</TabsTrigger>
        </TabsList>

        <TabsContent value="schema" className="mt-0 space-y-4">
          <Card>
            <CardContent className="p-6">
              <SchemaDefinitionView schemaData={telemetry} />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="versions" className="mt-0 space-y-4">
          <Card>
            <CardContent className="p-6">
              <VersionAssignmentView telemetry={telemetry} />
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <TelemetryProducersPanel
        schemaData={telemetry}
        isOpen={isProducersPanelOpen}
        onClose={() => setIsProducersPanelOpen(false)}
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
