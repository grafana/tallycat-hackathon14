import { useState, useEffect } from 'react'
import type { TelemetryItem, FilterState, SortField, SortDirection, ViewMode } from '@/types/telemetry'
import { mockTelemetryItems } from '@/data/mock-telemetry'

export const useTelemetry = () => {
  const [searchQuery, setSearchQuery] = useState('')
  const [viewMode, setViewMode] = useState<ViewMode>('list')
  const [activeFilters, setActiveFilters] = useState<FilterState>({})
  const [activeFilterCount, setActiveFilterCount] = useState(0)
  const [sortField, setSortField] = useState<SortField>('name')
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc')
  const [activeTab, setActiveTab] = useState('all')
  const [filteredItems, setFilteredItems] = useState<TelemetryItem[]>(mockTelemetryItems)

  const toggleFilter = (facetId: string, value: string) => {
    setActiveFilters((prev) => {
      const currentValues = prev[facetId] || []
      const newValues = currentValues.includes(value)
        ? currentValues.filter((v) => v !== value)
        : [...currentValues, value]

      // Update filter count
      const newCount = Object.values({
        ...prev,
        [facetId]: newValues,
      }).reduce((acc, curr) => acc + curr.length, 0)
      setActiveFilterCount(newCount)

      return {
        ...prev,
        [facetId]: newValues,
      }
    })
  }

  const removeFilter = (facetId: string, value: string) => {
    setActiveFilters((prev) => {
      const currentValues = prev[facetId] || []
      const newValues = currentValues.filter((v) => v !== value)

      // Update filter count
      const newCount = Object.values({
        ...prev,
        [facetId]: newValues,
      }).reduce((acc, curr) => acc + curr.length, 0)
      setActiveFilterCount(newCount)

      return {
        ...prev,
        [facetId]: newValues,
      }
    })
  }

  const clearAllFilters = () => {
    setActiveFilters({})
    setActiveFilterCount(0)
  }

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === "asc" ? "desc" : "asc")
    } else {
      setSortField(field)
      setSortDirection("asc")
    }
  }

  useEffect(() => {
    let results = [...mockTelemetryItems]

    // Filter by telemetry type tab
    if (activeTab !== "all") {
      results = results.filter((item) => item.type === activeTab)
    }

    // Apply search query
    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      results = results.filter(
        (item) =>
          item.name.toLowerCase().includes(query) ||
          item.description.toLowerCase().includes(query) ||
          (item.tags && item.tags.some((tag) => tag.toLowerCase().includes(query))),
      )
    }

    // Apply all active filters
    Object.entries(activeFilters).forEach(([facetId, selectedValues]) => {
      if (selectedValues.length > 0) {
        results = results.filter((item) => {
          const itemValue = item[facetId as keyof typeof item]
          if (Array.isArray(itemValue)) {
            return selectedValues.some((value) => itemValue.includes(value))
          }
          return selectedValues.includes(String(itemValue))
        })
      }
    })

    // Apply sorting
    results.sort((a, b) => {
      const aValue = a[sortField as keyof typeof a]
      const bValue = b[sortField as keyof typeof b]

      if (typeof aValue === "string" && typeof bValue === "string") {
        return sortDirection === "asc" ? aValue.localeCompare(bValue) : bValue.localeCompare(aValue)
      }

      // For dates
      if (sortField === "lastUpdated") {
        return sortDirection === "asc"
          ? new Date(a.lastUpdated).getTime() - new Date(b.lastUpdated).getTime()
          : new Date(b.lastUpdated).getTime() - new Date(a.lastUpdated).getTime()
      }

      return 0
    })

    setFilteredItems(results)
  }, [searchQuery, activeFilters, sortField, sortDirection, activeTab])

  return {
    searchQuery,
    setSearchQuery,
    viewMode,
    setViewMode,
    activeFilters,
    activeFilterCount,
    toggleFilter,
    removeFilter,
    clearAllFilters,
    sortField,
    sortDirection,
    handleSort,
    activeTab,
    setActiveTab,
    filteredItems,
  }
} 