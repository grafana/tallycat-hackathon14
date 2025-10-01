import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api-client'
import type { ListScopesParams } from '@/lib/api-client'

export const useTelemetryScopes = (telemetryKey: string, params: ListScopesParams) => {
  return useQuery({
    queryKey: ['telemetry-scopes', telemetryKey, params],
    queryFn: () => api.scopes.listByTelemetry(telemetryKey, params),
    placeholderData: (previousData) => previousData,
    enabled: !!telemetryKey,
  })
}
