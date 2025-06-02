import { useQuery } from '@tanstack/react-query'
import { getTechnicalFacets } from '@/services/technical-facets'

export const useTechnicalFacets = () => {
  return useQuery({
    queryKey: ['technicalFacets'],
    queryFn: getTechnicalFacets,
  })
}
