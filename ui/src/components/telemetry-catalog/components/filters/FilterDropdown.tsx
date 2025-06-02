import React from 'react'
import { Filter } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useTechnicalFacets } from '@/hooks'
import type { TechnicalFacet, FacetOption } from '@/data/technical-facets'

interface FilterDropdownProps {
  activeFilters: Record<string, string[]>
  activeFilterCount: number
  onToggleFilter: (facetId: string, value: string) => void
  isLoading: boolean
  error: Error | null
}

export const FilterDropdown: React.FC<FilterDropdownProps> = ({
  activeFilters,
  activeFilterCount,
  onToggleFilter,
  isLoading,
  error,
}) => {
  const { data: technicalFacets } = useTechnicalFacets()

  if (isLoading) {
    return (
      <Button variant="outline" size="sm" className="h-9" disabled>
        <Filter className="mr-2 h-4 w-4" />
        Loading...
      </Button>
    )
  }

  if (error) {
    return (
      <Button
        variant="outline"
        size="sm"
        className="h-9 text-destructive"
        disabled
      >
        <Filter className="mr-2 h-4 w-4" />
        Error loading filters
      </Button>
    )
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="h-9">
          <Filter className="mr-2 h-4 w-4" />
          Filter
          {activeFilterCount > 0 && (
            <Badge variant="secondary" className="ml-1 h-5 min-w-5 px-1">
              {activeFilterCount}
            </Badge>
          )}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-[220px]">
        <ScrollArea className="h-[500px]">
          {technicalFacets?.map((facet: TechnicalFacet) => (
            <div key={facet.id} className="px-2 py-1.5">
              <DropdownMenuLabel className="px-0">
                {facet.name}
              </DropdownMenuLabel>
              <DropdownMenuSeparator className="mb-1" />
              {facet.options.map((option: FacetOption) => (
                <DropdownMenuCheckboxItem
                  key={option.id}
                  checked={(activeFilters[facet.id] || []).includes(option.id)}
                  onCheckedChange={() => onToggleFilter(facet.id, option.id)}
                >
                  {option.name}
                </DropdownMenuCheckboxItem>
              ))}
              {facet !== technicalFacets[technicalFacets.length - 1] && (
                <DropdownMenuSeparator className="mt-1" />
              )}
            </div>
          ))}
        </ScrollArea>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
