import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api-client'

export const useTelemetryScopes = (telemetryKey: string) => {
  return useQuery({
    queryKey: ['telemetry-scopes', telemetryKey],
    queryFn: () => api.scopes.listByTelemetry(telemetryKey),
    placeholderData: (previousData) => previousData,
    enabled: !!telemetryKey,
  })
}
