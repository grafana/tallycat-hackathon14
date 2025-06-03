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
import type { TelemetryProducer } from '@/types/telemetry'
import { formatDate, DateFormat } from '@/lib/utils'

interface TelemetryProducersTableProps {
  producers: TelemetryProducer[]
  className?: string
}

export function TelemetryProducersTable({
  producers,
  className = '',
}: TelemetryProducersTableProps) {
  return (
    <div className={`rounded-lg border ${className}`}>
      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead className="w-[240px] font-semibold">Name</TableHead>
            <TableHead className="w-[140px] font-semibold">Namespace</TableHead>
            <TableHead className="w-[120px] font-semibold">Version</TableHead>
            <TableHead className="w-[140px] font-semibold">First Seen</TableHead>
            <TableHead className="w-[140px] font-semibold">Last Seen</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {producers.map((producer) => (
            <TableRow
              key={`${producer.name}-${producer.namespace}-${producer.version}`}
              className="hover:bg-muted/50"
            >
              <TableCell className="py-4">
                <div className="flex items-center gap-3">
                  <Server className="h-4 w-4 text-indigo-500 flex-shrink-0" />
                  <span className="font-medium">{producer.name || 'N/A'}</span>
                </div>
              </TableCell>
              <TableCell className="py-4">
                <div className="flex items-center gap-2">
                  <span className="font-mono text-sm whitespace-nowrap">
                    {producer.namespace || 'N/A'}
                  </span>
                </div>
              </TableCell>
              <TableCell className="py-4">
                <div className="flex items-center gap-2">
                  <Tag className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
                  <Badge variant="outline" className="font-mono">
                    {producer.version ? `v${producer.version}` : 'N/A'}
                  </Badge>
                </div>
              </TableCell>
              <TableCell className="py-4">
                <div className="flex items-center gap-2 text-sm">
                  <Clock className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
                  <span className="whitespace-nowrap font-mono">
                    {formatDate(producer.firstSeen, DateFormat.shortDateTime)}
                  </span>
                </div>
              </TableCell>
              <TableCell className="py-4">
                <div className="flex items-center gap-2 text-sm">
                  <Clock className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
                  <span className="whitespace-nowrap font-mono">
                    {formatDate(producer.lastSeen, DateFormat.shortDateTime)}
                  </span>
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
} 