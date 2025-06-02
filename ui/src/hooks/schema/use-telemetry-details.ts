import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api-client'

interface UseTelemetryDetailsOptions {
  telemetryName: string
}

export const useTelemetryDetails = ({
  telemetryName,
}: UseTelemetryDetailsOptions) => {
  return useQuery({
    queryKey: ['telemetry', telemetryName],
    queryFn: async () => {
      try {
        return await api.telemetries.getByKey(telemetryName)
      } catch (error) {
        if (
          error instanceof Error &&
          'status' in error &&
          (error as { status?: number }).status === 404
        ) {
          return null
        }
        throw error
      }
    },
  })
}
