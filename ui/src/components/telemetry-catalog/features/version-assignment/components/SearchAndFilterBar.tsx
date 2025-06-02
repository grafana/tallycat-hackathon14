import { Search } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { StatusFilter } from '@/components/telemetry-catalog/components/filters/StatusFilter'

interface SearchAndFilterBarProps {
  searchQuery: string
  onSearchChange: (value: string) => void
  activeStatus: string[]
  onStatusChange: (status: string[]) => void
}

export const SearchAndFilterBar = ({
  searchQuery,
  onSearchChange,
  activeStatus,
  onStatusChange,
}: SearchAndFilterBarProps) => {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-4">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search schema IDs..."
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
            className="pl-9 pr-4"
          />
        </div>
        <StatusFilter
          activeStatus={activeStatus}
          onStatusChange={onStatusChange}
        />
      </div>
    </div>
  )
}
