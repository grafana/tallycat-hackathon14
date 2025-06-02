import { Search } from 'lucide-react'
import { Input } from '@/components/ui/input'
import React from 'react'

interface SearchBarProps {
  value: string
  onChange: (value: string) => void
}

export const SearchBar: React.FC<SearchBarProps> = ({ value, onChange }) => (
  <div className="relative w-full sm:max-w-md">
    <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
    <Input
      type="search"
      placeholder="Search telemetry signals..."
      className="w-full pl-9 pr-4"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  </div>
)
