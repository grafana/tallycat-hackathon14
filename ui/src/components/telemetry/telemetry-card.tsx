import { ChevronDown, Clock } from 'lucide-react'
import { Link } from '@tanstack/react-router'
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  CardFooter,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { Telemetry } from '@/types/telemetry'
import { DataTypeIcon, getDataType } from './telemetry-icons'
import { getTelemetryTypeBgColor, getStatusBadge } from '@/utils/telemetry'
import { formatDate, DateFormat } from '@/lib/utils'

interface TelemetryCardProps {
  item: Telemetry
}

export const TelemetryCard = ({ item }: TelemetryCardProps) => {
  const statusBadge = getStatusBadge(undefined)
  const dataType = getDataType(item)

  return (
    <Card className="overflow-hidden">
      <CardHeader className="flex flex-row items-start justify-between space-y-0 pb-2">
        <div className="flex items-center space-x-2">
          <div
            className={`flex h-8 w-8 items-center justify-center rounded-md ${getTelemetryTypeBgColor(item.telemetryType)}`}
          >
            <DataTypeIcon dataType={dataType} />
          </div>
          <CardTitle className="text-base">
            <Link
              to={`/data-governance/telemetry-catalog`}
              className="hover:text-primary hover:underline"
            >
              {item.schemaKey}
            </Link>
          </CardTitle>
        </div>
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
      </CardHeader>
      <CardContent>
        <div className="flex items-center justify-between py-1">
          <span className="text-sm text-muted-foreground">Type</span>
          <div className="flex items-center gap-1.5">
            <DataTypeIcon dataType={dataType} />
            <Badge variant="outline" className="capitalize">
              {dataType}
            </Badge>
          </div>
        </div>
        <div className="flex items-center justify-between py-1">
          <span className="text-sm text-muted-foreground">Status</span>
          {statusBadge && (
            <Badge variant="outline" className={statusBadge.className}>
              {statusBadge.label}
            </Badge>
          )}
        </div>
        <div className="flex items-center justify-between py-1">
          <span className="text-sm text-muted-foreground">Format</span>
          <span className="text-sm font-mono">{item.protocol}</span>
        </div>
        <div className="mt-2">
          <p className="text-sm text-muted-foreground line-clamp-2">
            {item.brief}
          </p>
        </div>
      </CardContent>
      <CardFooter className="border-t bg-muted/30 px-6 py-3">
        <div className="flex w-full items-center justify-between text-xs text-muted-foreground">
          {/* <div className="flex items-center gap-1">
            <Tag className="h-3 w-3" />
            <span>{item.tags?.length || 0} tags</span>
          </div> */}
          <div className="flex items-center gap-1">
            <Clock className="h-3 w-3" />
            <span>{formatDate(item.updatedAt, DateFormat.short)}</span>
          </div>
        </div>
      </CardFooter>
    </Card>
  )
}
