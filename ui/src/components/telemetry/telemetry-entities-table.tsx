import { Server, Clock, Tag } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { TelemetryEntity } from '@/types/telemetry'
import { formatDate, DateFormat } from '@/lib/utils'

interface TelemetryEntitiesTableProps {
  entities: Record<string, TelemetryEntity>
  className?: string
}

export function TelemetryEntitiesTable({
  entities,
  className = '',
}: TelemetryEntitiesTableProps) {
  const entityList = Object.values(entities)
  return (
    <div className={`rounded-lg border ${className}`}>
      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead className="w-[140px] font-semibold">Type</TableHead>
            <TableHead className="w-[320px] font-semibold">Key Attributes</TableHead>
            <TableHead className="w-[140px] font-semibold">First Seen</TableHead>
            <TableHead className="w-[140px] font-semibold">Last Seen</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {entityList.map((entity) => {
            // Get key attributes to display (show more since we have more space)
            const keyAttributes = Object.entries(entity.attributes)
              .slice(0, 5)
              .map(([key, value]) => `${key}: ${value}`)
              .join(', ')

            return (
              <TableRow
                key={entity.id}
                className="hover:bg-muted/50"
              >
                <TableCell className="py-4">
                  <div className="flex items-center gap-2">
                    <Server className="h-4 w-4 text-indigo-500 flex-shrink-0" />
                    <Badge variant="outline" className="font-mono">
                      {entity.type}
                    </Badge>
                  </div>
                </TableCell>
                <TableCell className="py-4">
                  <div className="flex items-center gap-2">
                    <Tag className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
                    <span className="text-sm" title={keyAttributes}>
                      {keyAttributes || 'No attributes'}
                    </span>
                  </div>
                </TableCell>
                <TableCell className="py-4">
                  <div className="flex items-center gap-2 text-sm">
                    <Clock className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
                    <span className="whitespace-nowrap font-mono">
                      {formatDate(entity.firstSeen, DateFormat.shortDateTime)}
                    </span>
                  </div>
                </TableCell>
                <TableCell className="py-4">
                  <div className="flex items-center gap-2 text-sm">
                    <Clock className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
                    <span className="whitespace-nowrap font-mono">
                      {formatDate(entity.lastSeen, DateFormat.shortDateTime)}
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