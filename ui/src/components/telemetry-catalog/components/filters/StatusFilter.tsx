import { CheckCircle2, SlidersHorizontal } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Status } from '@/types/telemetry'

interface StatusFilterProps {
  activeStatus: string[]
  onStatusChange: (status: string[]) => void
}

export const StatusFilter = ({
  activeStatus,
  onStatusChange,
}: StatusFilterProps) => {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className={`h-9 gap-1 ${activeStatus.length > 0 ? 'bg-primary/10 border-primary/30 text-primary' : ''}`}
        >
          <SlidersHorizontal className="h-4 w-4" />
          Filter
          {activeStatus.length > 0 && (
            <Badge variant="secondary" className="ml-1">
              {activeStatus.length}
            </Badge>
          )}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-56">
        <DropdownMenuLabel>Filter by Status</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {Object.values(Status).map((status) => (
          <DropdownMenuItem
            key={status}
            className="cursor-pointer"
            onClick={() => onStatusChange([status])}
          >
            <div className="flex items-center gap-2">
              {activeStatus.includes(status) ? (
                <CheckCircle2 className="h-4 w-4 text-primary" />
              ) : (
                <div className="h-4 w-4" />
              )}
              <span className="capitalize">{status}</span>
            </div>
          </DropdownMenuItem>
        ))}
        <DropdownMenuSeparator />
        <DropdownMenuItem
          className="cursor-pointer text-muted-foreground"
          onClick={() => onStatusChange([])}
        >
          Clear all filters
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
