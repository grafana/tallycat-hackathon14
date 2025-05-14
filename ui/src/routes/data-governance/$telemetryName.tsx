import { createFileRoute, Link } from '@tanstack/react-router'
import { useParams } from '@tanstack/react-router'
import { TelemetryTypeIcon } from '@/components/telemetry/telemetry-icons'
import { Button } from '@/components/ui/button'
import { ArrowLeft, ChevronRight, Server } from 'lucide-react'
import type { TelemetryType, DataType, Status } from '@/types/telemetry'
import { useState } from 'react'
import { TelemetryOverviewPanel } from '@/components/schema-catalog/TelemetryOverviewPanel'

const mockTelemetryItems = {
  "system.memory.usage": {
    id: "telemetry-001",
    name: "system.memory.usage",
    type: "metric" as TelemetryType,
    dataType: "histogram" as DataType,
    status: "active" as Status,
    description: "Measures the duration of HTTP server requests in milliseconds",
    lastUpdated: "2023-11-15T10:30:00Z",
    schemaVersionCount: 1,
    created: "2023-09-10T08:15:00Z",
    fields: 24,
    source: "OpenTelemetry Collector",
    instrumentationLibrary: "opentelemetry-js",
    format: "OTLP",
    unit: "ms",
    aggregation: "cumulative",
    cardinality: "high",
    tags: ["http", "server", "duration", "request"],
    // NEW: Sources that generate this telemetry
    sources: [
      {
        id: "service-001",
        name: "api-gateway",
        team: "API Team",
        environment: "production",
        health: "healthy",
        version: "2.3.1",
        volume: 1250,
        dailyAverage: 1100,
        peak: 1800,
        contribution: 42,
        compliance: "compliant",
        requiredFieldsPresent: 7,
        requiredFieldsTotal: 7,
        optionalFieldsPresent: 3,
        optionalFieldsTotal: 3,
        lastValidated: "2023-11-15T08:30:00Z",
      },
      {
        id: "service-002",
        name: "user-service",
        team: "User Team",
        environment: "production",
        health: "healthy",
        version: "1.9.0",
        volume: 850,
        dailyAverage: 820,
        peak: 1200,
        contribution: 28,
        compliance: "compliant",
        requiredFieldsPresent: 7,
        requiredFieldsTotal: 7,
        optionalFieldsPresent: 2,
        optionalFieldsTotal: 3,
        lastValidated: "2023-11-14T14:15:00Z",
      },
      {
        id: "service-003",
        name: "order-service",
        team: "Order Team",
        environment: "production",
        health: "warning",
        version: "1.5.2",
        volume: 450,
        dailyAverage: 430,
        peak: 700,
        contribution: 15,
        compliance: "partial",
        requiredFieldsPresent: 7,
        requiredFieldsTotal: 7,
        optionalFieldsPresent: 1,
        optionalFieldsTotal: 3,
        lastValidated: "2023-11-13T11:45:00Z",
      },
      {
        id: "service-004",
        name: "payment-service",
        team: "Payment Team",
        environment: "production",
        health: "critical",
        version: "2.0.1",
        volume: 250,
        dailyAverage: 300,
        peak: 500,
        contribution: 8,
        compliance: "non-compliant",
        requiredFieldsPresent: 6,
        requiredFieldsTotal: 7,
        optionalFieldsPresent: 0,
        optionalFieldsTotal: 3,
        lastValidated: "2023-11-12T09:20:00Z",
      },
      {
        id: "service-005",
        name: "notification-service",
        team: "Notification Team",
        environment: "production",
        health: "healthy",
        version: "1.2.0",
        volume: 150,
        dailyAverage: 140,
        peak: 300,
        contribution: 5,
        compliance: "compliant",
        requiredFieldsPresent: 7,
        requiredFieldsTotal: 7,
        optionalFieldsPresent: 3,
        optionalFieldsTotal: 3,
        lastValidated: "2023-11-14T16:30:00Z",
      },
      {
        id: "service-006",
        name: "api-gateway-staging",
        team: "API Team",
        environment: "staging",
        health: "healthy",
        version: "2.4.0-beta",
        volume: 50,
        dailyAverage: 45,
        peak: 100,
        contribution: 2,
        compliance: "compliant",
        requiredFieldsPresent: 7,
        requiredFieldsTotal: 7,
        optionalFieldsPresent: 3,
        optionalFieldsTotal: 3,
        lastValidated: "2023-11-15T09:10:00Z",
      },
    ],
    // Teams that have services generating this telemetry
    sourceTeams: ["API Team", "User Team", "Order Team", "Payment Team", "Notification Team"],
    // Technical schema details
    schema: [
      { name: "timestamp", type: "timestamp", description: "Time when the metric was recorded", required: true },
      { name: "service.name", type: "string", description: "Name of the service", required: true },
      { name: "service.instance.id", type: "string", description: "Instance identifier", required: true },
      { name: "http.route", type: "string", description: "HTTP route template", required: true },
      { name: "http.method", type: "string", description: "HTTP method", required: true },
      { name: "http.status_code", type: "int", description: "HTTP status code", required: true },
      { name: "duration_ms", type: "double", description: "Duration in milliseconds", required: true },
      { name: "http.request.size", type: "int", description: "Size of the request in bytes", required: false },
      { name: "http.response.size", type: "int", description: "Size of the response in bytes", required: false },
      { name: "http.user_agent", type: "string", description: "User agent string", required: false },
    ],
    // Technical metadata
    metricDetails: {
      type: "Histogram",
      unit: "ms",
      aggregation: "Cumulative",
      metricName: "http.server.request.duration",
      otelCompatible: true,
      buckets: [0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000],
      monotonic: false,
      instrumentationScope: "io.opentelemetry.http",
      semanticConventions: "http",
    },
    // Technical usage information
    usedBy: [
      { name: "API Latency Monitoring", type: "dashboard", id: "dash-001" },
      { name: "SLO Tracking", type: "dashboard", id: "dash-002" },
      { name: "High Latency Alert", type: "alert", id: "alert-001" },
      { name: "Error Rate Correlation", type: "analysis", id: "analysis-001" },
    ],
    // Technical versioning
    history: [
      {
        version: "2.3.0",
        date: "2023-11-15T10:30:00Z",
        author: "Jane Smith",
        changes: "Added http.request.size and http.response.size fields",
        validationStatus: "passed",
      },
      {
        version: "2.2.0",
        date: "2023-10-20T14:15:00Z",
        author: "John Doe",
        changes: "Added http.user_agent field",
        validationStatus: "passed",
      },
      {
        version: "2.1.5",
        date: "2023-10-05T11:30:00Z",
        author: "John Doe",
        changes: "Updated field documentation to match OTel spec v1.11.0",
        validationStatus: "warning",
      },
      {
        version: "2.1.0",
        date: "2023-09-25T09:45:00Z",
        author: "Jane Smith",
        changes: "Added http.status_code field",
        validationStatus: "warning",
      },
      {
        version: "2.0.0",
        date: "2023-09-10T08:15:00Z",
        author: "John Doe",
        changes: "Initial release",
        validationStatus: "error",
      },
    ],
    // Technical examples
    examples: [
      {
        description: "Successful GET request",
        value: {
          timestamp: "2023-11-15T10:30:00.123Z",
          "service.name": "api-gateway",
          "service.instance.id": "instance-001",
          "http.route": "/users/{id}",
          "http.method": "GET",
          "http.status_code": 200,
          duration_ms: 45.2,
          "http.request.size": 1024,
          "http.response.size": 8192,
          "http.user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
        },
      },
      {
        description: "Failed POST request",
        value: {
          timestamp: "2023-11-15T10:35:00.456Z",
          "service.name": "api-gateway",
          "service.instance.id": "instance-002",
          "http.route": "/orders",
          "http.method": "POST",
          "http.status_code": 500,
          duration_ms: 2345.7,
          "http.request.size": 15360,
          "http.response.size": 512,
          "http.user_agent": "PostmanRuntime/7.29.0",
        },
      },
    ],
    // Technical validation rules
    validationRules: [
      { field: "duration_ms", rule: "value >= 0", description: "Duration must be non-negative" },
      {
        field: "http.status_code",
        rule: "value >= 100 && value < 600",
        description: "HTTP status code must be between 100 and 599",
      },
      {
        field: "http.method",
        rule: "value in ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS']",
        description: "HTTP method must be a valid method",
      },
    ],
  },
  // Other telemetry items would be here...
}

export const TelemetryDetails = () => {
  const { telemetryName } = useParams({ from: '/data-governance/$telemetryName' })
  const [_, setActiveTab] = useState("schema")

  const telemetry = mockTelemetryItems[telemetryName as keyof typeof mockTelemetryItems]

  const handleViewAllSources = () => {
    // setIsSourcesPanelOpen(true)
  }

  if (!telemetry) {
    return (
      <div className="flex flex-col items-center justify-center h-[50vh] gap-4">
        <h1 className="text-2xl font-medium">Telemetry signal not found</h1>
        <p className="text-muted-foreground">
          The telemetry signal you're looking for doesn't exist or has been removed.
        </p>
        <Button asChild>
          <Link to="/data-governance/schema-catalog">
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
            <Link to="/data-governance/schema-catalog">
              <ArrowLeft className="h-4 w-4" />
              <span className="sr-only">Back</span>
            </Link>
          </Button>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Link to="/data-governance/schema-catalog" className="hover:text-foreground">
              Telemetry Catalog
            </Link>
            <ChevronRight className="h-4 w-4" />
            <span className="font-medium text-foreground">{telemetry.name}</span>
          </div>
        </div>

        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-md bg-primary/10">
              {TelemetryTypeIcon({ type: telemetry.type })}
            </div>
            <div>
              <h1 className="text-2xl font-medium font-mono">{telemetry.name}</h1>
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <div className="flex items-center gap-1.5">
                  <span className="capitalize">
                    {telemetry.type} ({telemetry.dataType})
                  </span>
                </div>
              </div>
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            {/* OTel Compliant Badge */}
            {/* <Badge
              variant="outline"
              className="bg-green-500/10 text-green-500 border-green-500/20 flex items-center gap-1.5 px-2 py-1"
            >
              <CheckCircle2 className="h-3.5 w-3.5" />
              OTel Compliant
            </Badge> */}

            {/* Sources Badge */}
            <Button
              variant="outline"
              size="sm"
              className="h-7 px-2 text-xs flex items-center gap-1.5"
              onClick={handleViewAllSources}
            >
              <Server className="h-3.5 w-3.5" />
              {telemetry.sources.length} Sources
            </Button>

            {/* View Validation Button */}
            <Button variant="outline" size="sm" className="h-7 px-2 text-xs" onClick={() => setActiveTab("validation")}>
              View Validation
            </Button>
          </div>
        </div>
      
      </div>
      
      <TelemetryOverviewPanel telemetry={telemetry} onViewAllSources={handleViewAllSources} />
    </div>
  )
}

export const Route = createFileRoute('/data-governance/$telemetryName')({
  component: TelemetryDetails,
  validateSearch: (search: Record<string, unknown>) => {
    return search
  },
}) 