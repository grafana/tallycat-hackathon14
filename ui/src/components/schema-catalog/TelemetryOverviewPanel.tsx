"use client"
import {
  Info,
  BarChart2,
  PieChart,
  Timer,
  Activity,
  Calendar,
  Clock,
  Database,
} from "lucide-react"
import { type Telemetry } from "@/types/telemetry"

interface TelemetryOverviewPanelProps {
  telemetry: Telemetry
  onViewAllSources: () => void
}

export function TelemetryOverviewPanel({ telemetry }: TelemetryOverviewPanelProps) {
  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return new Intl.DateTimeFormat("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    }).format(date)
  }

  // Calculate source health counts

  // Calculate total volume

  return (
    <div className="bg-gradient-to-br from-background to-muted rounded-xl border shadow-sm overflow-hidden">
      {/* Description Section */}
      <div className="p-6 border-b">
        <div className="flex items-center gap-2 mb-2">
          <Info className="h-5 w-5 text-primary" />
          <h2 className="text-xl font-medium">Description</h2>
        </div>
        <p className="text-base leading-relaxed">{telemetry.brief || "No description available"}</p>
      </div>

      {/* Main Content Area - 3 columns (removed Source & Format column) */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-0">
        {/* Technical Details Column 1 - Metrics & Structure */}
        <div className="p-5 border-r">
          <h3 className="text-sm font-medium text-muted-foreground mb-3 flex items-center gap-1.5">
            <BarChart2 className="h-4 w-4 text-green-500" />
            Metrics & Structure
          </h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5">
                <Database className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-sm">Attributes</span>
              </div>
              <span className="text-sm">{telemetry.attributes.length}</span>
            </div>

            {/* <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5">
                <BarChart2 className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-sm">Cardinality</span>
              </div>
              <span className="text-sm capitalize">{telemetry.cardinality}</span>
            </div> */}

            {telemetry.metricType && (
              <>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-1.5">
                    <PieChart className="h-3.5 w-3.5 text-muted-foreground" />
                    <span className="text-sm">Type</span>
                  </div>
                  <span className="text-sm">{telemetry.metricType}</span>
                </div>

                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-1.5">
                    <Timer className="h-3.5 w-3.5 text-muted-foreground" />
                    <span className="text-sm">Unit</span>
                  </div>
                  <span className="text-sm">{telemetry.metricUnit}</span>
                </div>

                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-1.5">
                    <Activity className="h-3.5 w-3.5 text-muted-foreground" />
                    <span className="text-sm">Aggregation</span>
                  </div>
                  <span className="text-sm">{telemetry.metricTemporality}</span>
                </div>
              </>
            )}

            {/* {telemetry.telemetryType === "trace" && telemetry.spanKind && (
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-1.5">
                  <Activity className="h-3.5 w-3.5 text-muted-foreground" />
                  <span className="text-sm">Span Kind</span>
                </div>
                <span className="text-sm">{telemetry.spanKind}</span>
              </div>
            )} */}
          </div>
        </div>

        {/* Timeline Column - Renamed to "History" and removed tags */}
        <div className="p-5 border-r">
          <h3 className="text-sm font-medium text-muted-foreground mb-3 flex items-center gap-1.5">
            <Clock className="h-4 w-4 text-purple-500" />
            History
          </h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5">
                <Calendar className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-sm">Created</span>
              </div>
              <span className="text-sm">{formatDate(telemetry.createdAt)}</span>
            </div>

            <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5">
                <Clock className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-sm">Updated</span>
              </div>
              <span className="text-sm">{formatDate(telemetry.updatedAt)}</span>
            </div>

            <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5">
                <Activity className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-sm">Version</span>
              </div>
              <span className="text-sm font-medium">v{telemetry.schemaVersion || "1.0.0"}</span>
            </div>
          </div>
        </div>

        {/* Data Producers Column - Renamed from "Sources" */}
        <div className="p-5">
          <h3 className="text-sm font-medium text-muted-foreground mb-3 flex items-center gap-1.5">
            <Database className="h-4 w-4 text-indigo-500" />
            Data Producers
          </h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5">
                <Database className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-sm">Services</span>
              </div>
              <span className="text-sm font-medium">{1} services</span>
            </div>

            {/* <div>
              <div className="flex items-center justify-between mb-1.5">
                <div className="flex items-center gap-1.5">
                  <Activity className="h-3.5 w-3.5 text-muted-foreground" />
                  <span className="text-sm">Health</span>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <div className="flex items-center gap-1">
                        <CheckCircle2 className="h-3.5 w-3.5 text-green-500" />
                        <span className="text-sm">{sourceHealthCounts.healthy}</span>
                      </div>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>{sourceHealthCounts.healthy} healthy producers</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>

                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <div className="flex items-center gap-1">
                        <AlertTriangle className="h-3.5 w-3.5 text-yellow-500" />
                        <span className="text-sm">{sourceHealthCounts.warning}</span>
                      </div>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>{sourceHealthCounts.warning} producers with warnings</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>

                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <div className="flex items-center gap-1">
                        <AlertCircle className="h-3.5 w-3.5 text-red-500" />
                        <span className="text-sm">{sourceHealthCounts.critical}</span>
                      </div>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>{sourceHealthCounts.critical} critical producers</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
            </div> */}

            {/* <div>
              <div className="flex items-center justify-between mb-1.5">
                <div className="flex items-center gap-1.5">
                  <BarChart2 className="h-3.5 w-3.5 text-muted-foreground" />
                  <span className="text-sm">Volume</span>
                </div>
                <span className="text-sm">{totalVolume.toLocaleString()} events/min</span>
              </div>
              <Progress value={100} className="h-1.5" />
            </div>

            <div className="pt-1">
              <Button variant="ghost" size="sm" className="w-full justify-between" onClick={onViewAllSources}>
                <span>View all producers</span>
                <ChevronRight className="h-4 w-4" />
              </Button>
            </div> */}
          </div>
        </div>
      </div>
    </div>
  )
}
