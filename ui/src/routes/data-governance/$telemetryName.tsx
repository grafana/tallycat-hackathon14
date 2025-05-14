import { createFileRoute } from '@tanstack/react-router'
import { useParams } from '@tanstack/react-router'
import { DataTypeIcon } from '@/components/telemetry/telemetry-icons'
import { Badge } from '@/components/ui/badge'
import { getTelemetryTypeBgColor, getStatusBadge } from '@/utils/telemetry'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import type { TelemetryType, DataType, Status } from '@/types/telemetry'

export const TelemetryDetails = () => {
  const { telemetryName } = useParams({ from: '/data-governance/$telemetryName' })

  // TODO: Replace with actual data fetching
  const mockTelemetry = {
    name: telemetryName,
    type: 'metric' as TelemetryType,
    dataType: 'gauge' as DataType,
    description: 'This is a sample telemetry description',
    status: 'active' as Status,
    format: 'prometheus',
    lastUpdated: new Date().toISOString(),
    schemaVersionCount: 3,
  }

  return (
    <div className="mx-auto">
      <div className="flex flex-col gap-6">
        {/* Header Section */}
        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-3">
            <div
              className={`flex h-12 w-12 items-center justify-center rounded-lg ${getTelemetryTypeBgColor(
                mockTelemetry.type,
              )}`}
            >
              <DataTypeIcon dataType={mockTelemetry.dataType} />
            </div>
            <div>
              <h1 className="text-3xl font-medium">{mockTelemetry.name}</h1>
              <p className="text-muted-foreground">{mockTelemetry.description}</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Badge variant="outline" className="capitalize">
              {mockTelemetry.type}
            </Badge>
            {getStatusBadge(mockTelemetry.status) && (
              <Badge variant="outline" className={getStatusBadge(mockTelemetry.status)?.className}>
                {getStatusBadge(mockTelemetry.status)?.label}
              </Badge>
            )}
          </div>
        </div>

        {/* Main Content */}
        <Tabs defaultValue="overview" className="w-full">
          <TabsList>
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="schema">Schema</TabsTrigger>
            <TabsTrigger value="usage">Usage</TabsTrigger>
            <TabsTrigger value="history">History</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="mt-6">
            <div className="grid gap-6 md:grid-cols-2">
              <Card>
                <CardHeader>
                  <CardTitle>Details</CardTitle>
                  <CardDescription>Basic information about this telemetry</CardDescription>
                </CardHeader>
                <CardContent>
                  <dl className="grid gap-4">
                    <div>
                      <dt className="text-sm font-medium text-muted-foreground">Data Type</dt>
                      <dd className="flex items-center gap-1.5">
                        <DataTypeIcon dataType={mockTelemetry.dataType} />
                        <span>{mockTelemetry.dataType}</span>
                      </dd>
                    </div>
                    <div>
                      <dt className="text-sm font-medium text-muted-foreground">Format</dt>
                      <dd className="font-mono text-sm">{mockTelemetry.format}</dd>
                    </div>
                    <div>
                      <dt className="text-sm font-medium text-muted-foreground">Last Updated</dt>
                      <dd>{new Date(mockTelemetry.lastUpdated).toLocaleString()}</dd>
                    </div>
                    <div>
                      <dt className="text-sm font-medium text-muted-foreground">Schema Versions</dt>
                      <dd>{mockTelemetry.schemaVersionCount}</dd>
                    </div>
                  </dl>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Metadata</CardTitle>
                  <CardDescription>Additional metadata and tags</CardDescription>
                </CardHeader>
                <CardContent>
                  <p className="text-muted-foreground">No metadata available</p>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          <TabsContent value="schema" className="mt-6">
            <Card>
              <CardHeader>
                <CardTitle>Schema Definition</CardTitle>
                <CardDescription>Current schema version and structure</CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-muted-foreground">Schema details will be displayed here</p>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="usage" className="mt-6">
            <Card>
              <CardHeader>
                <CardTitle>Usage Statistics</CardTitle>
                <CardDescription>How this telemetry is being used</CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-muted-foreground">Usage statistics will be displayed here</p>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="history" className="mt-6">
            <Card>
              <CardHeader>
                <CardTitle>Version History</CardTitle>
                <CardDescription>Previous versions and changes</CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-muted-foreground">Version history will be displayed here</p>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}

export const Route = createFileRoute('/data-governance/$telemetryName')({
  component: TelemetryDetails,
  validateSearch: (search: Record<string, unknown>) => {
    return search
  },
}) 