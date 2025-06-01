import { ChevronDown } from 'lucide-react'
import { Link } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import type { Telemetry } from '@/types/telemetry'
import { DataTypeIcon } from './telemetry-icons'
import { getTelemetryTypeBgColor, getStatusBadge } from '@/utils/telemetry'
import { formatDate, DateFormat } from '@/lib/utils'

interface TelemetryTableRowProps {
  item: Telemetry
}

export const TelemetryTableRow = ({ item }: TelemetryTableRowProps) => {
  const statusBadge = getStatusBadge(item.status)

  return (
    <tr className="border-b">
      <td className="px-4 py-3">
        <div className="flex items-center gap-3">
          <div
            className={`flex h-8 w-8 items-center justify-center rounded-md ${getTelemetryTypeBgColor(
              item.type,
            )}`}
          >
            <DataTypeIcon dataType={item.dataType} />
          </div>
          <div>
            <Link
              to={`/data-governance/schema-catalog`}
              className="font-medium hover:text-primary hover:underline"
            >
              {item.name}
            </Link>
            <p className="text-xs text-muted-foreground line-clamp-1">{item.description}</p>
          </div>
        </div>
      </td>
      <td className="px-4 py-3">
        <Badge variant="outline" className="capitalize">
          {item.type}
        </Badge>
      </td>
      <td className="px-4 py-3">
        <div className="flex items-center gap-1.5">
          <DataTypeIcon dataType={item.dataType} />
          <span className="text-sm">{item.dataType}</span>
        </div>
      </td>
      <td className="px-4 py-3">
        {statusBadge && (
          <Badge variant="outline" className={statusBadge.className}>
            {statusBadge.label}
          </Badge>
        )}
      </td>
      <td className="hidden px-4 py-3 md:table-cell">
        <span className="font-mono text-xs">{item.format}</span>
      </td>
      <td className="hidden px-4 py-3 lg:table-cell">{formatDate(item.lastUpdated, DateFormat.short)}</td>
      <td className="px-4 py-3">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="h-8 w-8">
              <ChevronDown className="h-4 w-4" />
              <span className="sr-only">Open menu</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem>View Details</DropdownMenuItem>
            <DropdownMenuItem>Edit</DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem>Export</DropdownMenuItem>
            <DropdownMenuItem className="text-red-500">Delete</DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </td>
    </tr>
  )
} 