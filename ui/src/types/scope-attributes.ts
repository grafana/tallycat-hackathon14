export interface ScopeAttribute {
  name: string
  values: string
  uniqueCount?: number // Optional: for displaying stats
}

export interface ScopeAttributeData {
  scopeName: string
  attributes: ScopeAttribute[]
  totalUniqueValues: number
}

export interface ScopeAttributesState {
  data: ScopeAttributeData | null
  isLoading: boolean
  error: string | null
}
