import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api-client'

interface AssignSchemaVersionParams {
  schemaKey: string
  schemaId: string
  version: string
  description: string
}

export const useAssignSchemaVersion = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      schemaKey,
      schemaId,
      version,
      description,
    }: AssignSchemaVersionParams) => {
      return api.telemetries.assignTelemetrySchemaVersion(schemaKey, {
        schemaId,
        version,
        description,
      })
    },
    onSuccess: () => {
      // Invalidate relevant queries to refresh the data
      queryClient.invalidateQueries({ queryKey: ['schemaAssignments'] })
      queryClient.invalidateQueries({ queryKey: ['schemas'] })
    },
  })
}
