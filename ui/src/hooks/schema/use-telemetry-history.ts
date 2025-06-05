import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api-client'
import type { TelemetryHistory } from '@/types/telemetry'
import type { ListResponse } from '@/lib/api-client'

interface UseTelemetryHistoryParams {
  telemetryKey: string
  page?: number
  pageSize?: number
  enabled?: boolean
}

export function useTelemetryHistory({
  telemetryKey,
  page = 1,
  pageSize = 10,
  enabled = true,
}: UseTelemetryHistoryParams) {
  return useQuery({
    queryKey: ['telemetry-history', telemetryKey, page, pageSize],
    queryFn: async () => {
      const searchParams = new URLSearchParams({
        page: page.toString(),
        pageSize: pageSize.toString(),
      })
      return apiClient.get<ListResponse<TelemetryHistory>>(
        `/api/v1/telemetries/${telemetryKey}/history?${searchParams.toString()}`
      )
    },
    enabled: enabled && !!telemetryKey,
  })
} 