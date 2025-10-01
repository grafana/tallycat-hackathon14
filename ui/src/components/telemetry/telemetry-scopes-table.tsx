import { Clock, Globe, Package, Tag } from 'lucide-react'
import type { TelemetryScope } from '@/types/telemetry'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { DateFormat, formatDate, formatScopeName } from '@/lib/utils'

interface TelemetryScopesTableProps {
  scopes: Array<TelemetryScope>
  className?: string
}

export const TelemetryScopesTable = ({
  scopes,
  className = '',
}: TelemetryScopesTableProps) => {
  return (
    <div className={`rounded-lg border ${className}`}>
      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead className="w-[300px] font-semibold">Name</TableHead>
            <TableHead className="w-[120px] font-semibold">Version</TableHead>
            <TableHead className="w-[200px] font-semibold">Schema URL</TableHead>
            <TableHead className="w-[140px] font-semibold">First Seen</TableHead>
            <TableHead className="w-[140px] font-semibold">Last Seen</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {scopes.map((scope) => {
            // Get key attributes to display if any exist
            const keyAttributes = Object.entries(scope.attributes)
              .slice(0, 3)
              .map(([key, value]) => `${key}: ${value}`)
              .join(', ')

            return (
              <TableRow
                key={scope.id}
                className="hover:bg-muted/50"
              >
                <TableCell className="py-4">
                  <div className="flex flex-col gap-1">
                    <div className="flex items-center gap-2">
                      <Package className="h-4 w-4 text-blue-500 flex-shrink-0" />
                      <span className="font-medium text-sm" title={scope.name}>
                        {formatScopeName(scope.name)}
                      </span>
                    </div>
                    {keyAttributes && (
                      <div className="flex items-center gap-2 ml-6">
                        <Tag className="h-3 w-3 text-muted-foreground flex-shrink-0" />
                        <span className="text-xs text-muted-foreground" title={keyAttributes}>
                          {keyAttributes}
                        </span>
                      </div>
                    )}
                  </div>
                </TableCell>
                <TableCell className="py-4">
                  <Badge variant="outline" className="font-mono text-xs">
                    {scope.version || 'N/A'}
                  </Badge>
                </TableCell>
                <TableCell className="py-4">
                  {scope.schemaURL ? (
                    <div className="flex items-center gap-2">
                      <Globe className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
                      <span className="text-sm font-mono truncate" title={scope.schemaURL}>
                        {scope.schemaURL}
                      </span>
                    </div>
                  ) : (
                    <span className="text-sm text-muted-foreground">No schema URL</span>
                  )}
                </TableCell>
                <TableCell className="py-4">
                  <div className="flex items-center gap-2 text-sm">
                    <Clock className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
                    <span className="whitespace-nowrap font-mono">
                      {formatDate(scope.firstSeen, DateFormat.shortDateTime)}
                    </span>
                  </div>
                </TableCell>
                <TableCell className="py-4">
                  <div className="flex items-center gap-2 text-sm">
                    <Clock className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
                    <span className="whitespace-nowrap font-mono">
                      {formatDate(scope.lastSeen, DateFormat.shortDateTime)}
                    </span>
                  </div>
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}
