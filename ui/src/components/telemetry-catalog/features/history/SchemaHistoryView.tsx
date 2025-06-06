"use client"
import { Badge } from "@/components/ui/badge"
import type { Telemetry } from "@/types/telemetry"
import { useTelemetryHistory } from "@/hooks/schema/use-telemetry-history"
import { formatDate, DateFormat } from "@/lib/utils"
import { LoadingState } from "@/components/telemetry-catalog/components/states/LoadingState"
import { ErrorState } from "@/components/telemetry-catalog/components/states/ErrorState"
import { Clock, Sparkles, GitBranch } from "lucide-react"
import type { TelemetryHistory } from "@/types/telemetry"

interface SchemaHistoryViewProps {
    telemetry: Telemetry
}

const EmptyState = () => (
    <div className="flex flex-col items-center justify-center py-8 px-6 text-center">
        <div className="relative mb-4">
            <div className="flex items-center justify-center w-12 h-12 rounded-full bg-muted/50 mb-2">
                <Clock className="h-6 w-6 text-muted-foreground/60" />
            </div>
            <div className="absolute -top-1 -right-1 w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center">
                <Sparkles className="h-3 w-3 text-primary" />
            </div>
        </div>

        <h4 className="text-lg font-semibold mb-1">No Version History Yet</h4>
        <p className="text-muted-foreground mb-4 max-w-md">
            This schema is just getting started! As changes are made and new versions are created, you'll see a
            beautiful timeline of its evolution here.
        </p>

        <div className="flex items-center gap-4 text-sm text-muted-foreground">
            <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-primary"></div>
                <span>Version releases</span>
            </div>
            <div className="flex items-center gap-2">
                <GitBranch className="h-3 w-3" />
                <span>Change tracking</span>
            </div>
            <div className="flex items-center gap-2">
                <Clock className="h-3 w-3" />
                <span>Timeline view</span>
            </div>
        </div>

        <div className="mt-4 p-3 bg-muted/30 rounded-lg">
            <p className="text-xs text-muted-foreground">
                💡 <strong>Coming soon:</strong> Version history will automatically populate as your schema evolves
            </p>
        </div>
    </div>
)

interface TimelineItemProps {
    version: TelemetryHistory
    isLast: boolean
}

const TimelineItem = ({ version, isLast }: TimelineItemProps) => (
    <div className="relative pl-6">
        {!isLast && (
            <div className="absolute left-2 top-2 h-full w-0.5 bg-border" />
        )}
        <div className="absolute left-0 top-2 h-4 w-4 rounded-full border-2 border-primary bg-background" />
        <div className="space-y-1">
            <div className="flex items-center gap-2">
                <Badge variant="outline" className="font-mono text-xs">
                    v{version.version}
                </Badge>
                <span className="text-sm text-muted-foreground">
                    {formatDate(version.timestamp, DateFormat.datetime)}
                </span>
            </div>
            <div className="flex items-center justify-between">
                <p className="font-medium">{version.summary}</p>
                {version.status && (
                    <Badge variant="secondary" className="ml-2">
                        {version.status}
                    </Badge>
                )}
            </div>
        </div>
    </div>
)

interface TimelineProps {
    items: TelemetryHistory[]
    schemaKey: string
}

const Timeline = ({ items, schemaKey }: TimelineProps) => (
    <div className="space-y-6">
        <div>
            <h3 className="text-lg font-medium">Version History</h3>
            <p className="text-sm text-muted-foreground">
                Track changes to {schemaKey} over time
            </p>
        </div>
        <div className="space-y-8">
            {items.map((version, index) => (
                <TimelineItem
                    key={version.id}
                    version={version}
                    isLast={index === items.length - 1}
                />
            ))}
        </div>
    </div>
)

export function SchemaHistoryView({ telemetry }: SchemaHistoryViewProps) {
    const { data, isLoading, error } = useTelemetryHistory({
        telemetryKey: telemetry.schemaKey,
    })

    if (isLoading) {
        return <LoadingState />
    }

    if (error) {
        return <ErrorState />
    }

    return !data?.items.length ? (
        <EmptyState />
    ) : (
        <Timeline items={data.items} schemaKey={telemetry.schemaKey} />
    )
}
