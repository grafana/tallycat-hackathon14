import { technicalFacets } from '@/data/technical-facets'
import type { TechnicalFacet } from '@/data/technical-facets'

// This is a mock API call that will be replaced with actual API call later
export const getTechnicalFacets = async (): Promise<TechnicalFacet[]> => {
  // Simulate API delay
  await new Promise((resolve) => setTimeout(resolve, 500))
  return technicalFacets
}
