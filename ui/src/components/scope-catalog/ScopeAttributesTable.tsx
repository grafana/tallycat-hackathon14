import { Database } from 'lucide-react'
import { useScopeCatalog } from '@/hooks/scope'
import { Telemetry, TelemetryType } from '@/types/telemetry'

interface ScopeTelemetryItem {
  schemaKey: string
  telemetryType: TelemetryType
}

interface ScopeTelemetriesTableProps {
  telemetries: ScopeTelemetryItem[]
}

interface ScopeTelemetriesExpansionProps {
  scopeName: string
  isExpanded: boolean
}

// Get badge color for telemetry type
const getTelemetryTypeBadgeColor = (type: TelemetryType): string => {
  switch (type) {
    case TelemetryType.Metric:
      return 'bg-green-100 text-green-700'
    case TelemetryType.Log:
      return 'bg-orange-100 text-orange-700'
    case TelemetryType.Span:
      return 'bg-purple-100 text-purple-700'
    case TelemetryType.Profile:
      return 'bg-red-100 text-red-700'
    default:
      return 'bg-gray-100 text-gray-700'
  }
}

// Component to display the telemetries table
const ScopeTelemetriesTable = ({ telemetries }: ScopeTelemetriesTableProps) => {
  if (telemetries.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        <Database className="mx-auto h-8 w-8 mb-2 opacity-50" />
        <div className="text-sm">No telemetries found for this scope</div>
      </div>
    )
  }

  return (
    <div className="border rounded-lg overflow-hidden bg-background">
      <table className="w-full">
        <thead className="bg-muted/30">
          <tr>
            <th className="text-left p-3 font-medium text-muted-foreground text-sm">
              Schema Key
            </th>
            <th className="text-left p-3 font-medium text-muted-foreground text-sm">
              Type
            </th>
          </tr>
        </thead>
        <tbody>
          {telemetries.map((telemetry, index) => (
            <tr
              key={`${telemetry.schemaKey}-${telemetry.telemetryType}`}
              className={`border-t ${index % 2 === 0 ? 'bg-background' : 'bg-muted/10'}`}
            >
              <td className="p-3">
                <div className="font-mono text-sm text-foreground">
                  {telemetry.schemaKey}
                </div>
              </td>
              <td className="p-3">
                <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getTelemetryTypeBadgeColor(telemetry.telemetryType)}`}>
                  {telemetry.telemetryType}
                </span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

// Main expansion component that shows telemetries for a scope
export const ScopeTelemetriesExpansion = ({ 
  scopeName, 
  isExpanded 
}: ScopeTelemetriesExpansionProps) => {
  const { rows } = useScopeCatalog()

  // Don't render anything if not expanded
  if (!isExpanded) {
    return null
  }

  // We need to get the telemetries from the raw data
  // For now, we'll use a simple approach and get telemetries from the hook
  const { rows: catalogRows } = useScopeCatalog()
  
  // This is a placeholder - we need to get the actual telemetries
  // We'll update this with proper data access
  const telemetries: ScopeTelemetryItem[] = []

  // Render the telemetries table
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-foreground">
          Scope Telemetries
        </h4>
        <div className="text-xs text-muted-foreground">
          {telemetries.length} telemetries
        </div>
      </div>
      <ScopeTelemetriesTable telemetries={telemetries} />
    </div>
  )
}
