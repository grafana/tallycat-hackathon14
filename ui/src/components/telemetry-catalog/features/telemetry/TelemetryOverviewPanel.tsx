'use client'
import {
  Info,
  BarChart2,
  PieChart,
  Timer,
  Activity,
  Calendar,
  Clock,
  Database,
  Package,
} from 'lucide-react'
import { type Telemetry } from '@/types/telemetry'
import { formatDate, DateFormat } from '@/lib/utils'

interface TelemetryOverviewPanelProps {
  telemetry: Telemetry
  scopesCount?: number
  onViewAllSources: () => void
  onViewAllScopes?: () => void
}

export function TelemetryOverviewPanel({
  telemetry,
  scopesCount = 0,
  onViewAllSources,
  onViewAllScopes,
}: TelemetryOverviewPanelProps) {
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
        <p className="text-base leading-relaxed">
          {telemetry.brief || 'No description available'}
        </p>
      </div>

      {/* Main Content Area - 4 columns */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-0">
        {/* Technical Details Column 1 - Metrics & Structure */}
        <div className="p-5 border-r">
          <h3 className="text-sm font-medium text-muted-foreground mb-3 flex items-center gap-1.5">
            <BarChart2 className="h-4 w-4 text-green-500" />
            {telemetry.telemetryType} & Structure
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

            {telemetry.telemetryType === "Span" && telemetry.spanKind && (
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-1.5">
                  <Activity className="h-3.5 w-3.5 text-muted-foreground" />
                  <span className="text-sm">Span Kind</span>
                </div>
                <span className="text-sm">{telemetry.spanKind}</span>
              </div>
            )}
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
              <span className="text-sm">
                {formatDate(telemetry.createdAt, DateFormat.datetime)}
              </span>
            </div>

            <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5">
                <Clock className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-sm">Updated</span>
              </div>
              <span className="text-sm">
                {formatDate(telemetry.updatedAt, DateFormat.datetime)}
              </span>
            </div>
          </div>
        </div>

        {/* Entities Column */}
        <div className="p-5 border-r">
          <h3 className="text-sm font-medium text-muted-foreground mb-3 flex items-center gap-1.5">
            <Database className="h-4 w-4 text-indigo-500" />
            Entities
          </h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5">
                <Database className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-sm">Entities</span>
              </div>
              <span className="text-sm font-medium">
                {Object.keys(telemetry.entities).length}
                {Object.keys(telemetry.entities).length > 1
                  ? ' Entities'
                  : ' Entity'}
              </span>
            </div>

            {onViewAllSources && Object.keys(telemetry.entities).length > 0 && (
              <div className="pt-1">
                <button
                  onClick={onViewAllSources}
                  className="text-xs text-muted-foreground hover:text-foreground transition-colors"
                >
                  View all entities →
                </button>
              </div>
            )}

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
                      <p>{sourceHealthCounts.healthy} healthy entities</p>
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
                      <p>{sourceHealthCounts.warning} entities with warnings</p>
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
                      <p>{sourceHealthCounts.critical} critical entities</p>
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
                <span>View all entities</span>
                <ChevronRight className="h-4 w-4" />
              </Button>
            </div> */}
          </div>
        </div>

        {/* Scopes Column */}
        <div className="p-5">
          <h3 className="text-sm font-medium text-muted-foreground mb-3 flex items-center gap-1.5">
            <Package className="h-4 w-4 text-blue-500" />
            Scopes
          </h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5">
                <Package className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-sm">Scopes</span>
              </div>
              <span className="text-sm font-medium">
                {scopesCount}
                {scopesCount > 1 ? ' Scopes' : ' Scope'}
              </span>
            </div>

            {onViewAllScopes && scopesCount > 0 && (
              <div className="pt-1">
                <button
                  onClick={onViewAllScopes}
                  className="text-xs text-muted-foreground hover:text-foreground transition-colors"
                >
                  View all scopes →
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
