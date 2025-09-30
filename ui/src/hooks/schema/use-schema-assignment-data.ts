import { useState, useMemo } from 'react'
import { useSchemaAssignments } from '@/hooks'
import { Status } from '@/types/telemetry'
import type { TelemetrySchema } from '@/types/telemetry-schema'

interface UseSchemaAssignmentDataReturn {
  searchQuery: string
  setSearchQuery: (value: string) => void
  activeStatus: string[]
  setActiveStatus: (status: string[]) => void
  currentPage: number
  setCurrentPage: (page: number) => void
  pageSize: number
  setPageSize: (size: number) => void
  tableData: TelemetrySchema[]
  isLoading: boolean
  error: Error | null
  totalCount: number
}

export const useSchemaAssignmentData = (
  schemaKey: string,
): UseSchemaAssignmentDataReturn => {
  const [searchQuery, setSearchQuery] = useState('')
  const [activeStatus, setActiveStatus] = useState<string[]>([])
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)

  const { data, isLoading, error } = useSchemaAssignments({
    schemaKey,
    search: searchQuery,
    status: activeStatus,
    page: currentPage,
    pageSize,
  })

  const tableData = useMemo(
    () =>
      (data?.items ?? []).map((item) => ({
        id: item.schemaId,
        name: item.schemaId,
        status:
          item.version && item.version !== 'Unassigned'
            ? Status.Active
            : Status.Experimental,
        version: item.version === 'Unassigned' ? null : item.version,
        entities: Object.fromEntries(
          Array(item.entityCount).fill(0).map((_, i) => [`entity${i}`, {
            id: `entity${i}`,
            type: 'unknown',
            attributes: {},
            firstSeen: item.lastSeen,
            lastSeen: item.lastSeen,
          }])
        ),
        attributes: [],
        lastSeen: item.lastSeen,
        discoveredAt: '',
        resourceAttributes: [],
        instrumentationAttributes: [],
        telemetryAttributes: [],
      })),
    [data?.items],
  )

  return {
    searchQuery,
    setSearchQuery,
    activeStatus,
    setActiveStatus,
    currentPage,
    setCurrentPage,
    pageSize,
    setPageSize,
    tableData,
    isLoading,
    error,
    totalCount: data?.total ?? 0,
  }
}
