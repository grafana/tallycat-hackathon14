import { Telemetry } from '@/types/telemetry'
import { ScopeAttribute, ScopeAttributeData } from '@/types/scope-attributes'

/**
 * Aggregates scope attributes for a specific scope name from telemetries
 */
export const aggregateScopeAttributes = (
  telemetries: Telemetry[],
  scopeName: string
): ScopeAttribute[] => {
  // Filter telemetries that have the specified scope name
  const relevantTelemetries = telemetries.filter(
    t => t.scope !== null && t.scope.name === scopeName
  )

  if (relevantTelemetries.length === 0) {
    return []
  }

  // Collect all attributes from all scopes with this name
  const attributeMap = new Map<string, Set<string>>()

  relevantTelemetries.forEach(telemetry => {
    const scope = telemetry.scope!
    
    if (scope.attributes && typeof scope.attributes === 'object') {
      Object.entries(scope.attributes).forEach(([attrName, attrValue]) => {
        if (!attributeMap.has(attrName)) {
          attributeMap.set(attrName, new Set())
        }
        
        // Convert value to string and add to set
        const valueStr = attrValue !== null && attrValue !== undefined 
          ? String(attrValue) 
          : ''
        
        if (valueStr.trim()) {
          attributeMap.get(attrName)!.add(valueStr)
        }
      })
    }
  })

  // Transform map to ScopeAttribute array
  const attributes: ScopeAttribute[] = Array.from(attributeMap.entries()).map(([name, valueSet]) => {
    const valuesArray = Array.from(valueSet).sort()
    
    return {
      name,
      values: valuesArray.join(', '),
      uniqueCount: valuesArray.length,
    }
  })

  // Sort attributes by name for consistent display
  return attributes.sort((a, b) => a.name.localeCompare(b.name))
}

/**
 * Creates complete scope attribute data for a given scope name
 */
export const createScopeAttributeData = (
  telemetries: Telemetry[],
  scopeName: string
): ScopeAttributeData => {
  const attributes = aggregateScopeAttributes(telemetries, scopeName)
  const totalUniqueValues = attributes.reduce((sum, attr) => sum + (attr.uniqueCount || 0), 0)

  return {
    scopeName,
    attributes,
    totalUniqueValues,
  }
}

/**
 * Aggregates attributes for all scopes from telemetries
 */
export const aggregateAllScopeAttributes = (telemetries: Telemetry[]): Record<string, ScopeAttributeData> => {
  // Get unique scope names
  const scopeNames = new Set(
    telemetries
      .filter(t => t.scope !== null)
      .map(t => t.scope!.name)
  )

  // Create attribute data for each scope
  const result: Record<string, ScopeAttributeData> = {}
  
  scopeNames.forEach(scopeName => {
    result[scopeName] = createScopeAttributeData(telemetries, scopeName)
  })

  return result
}

/**
 * Validates that telemetries have the expected structure for scope attribute aggregation
 */
export const validateTelemetriesForScopeAttributes = (telemetries: Telemetry[], scopeName: string): {
  valid: boolean
  totalTelemetries: number
  relevantTelemetries: number
  scopesWithAttributes: number
  totalAttributes: number
} => {
  const totalTelemetries = telemetries.length
  const relevantTelemetries = telemetries.filter(
    t => t.scope !== null && t.scope.name === scopeName
  )
  
  const scopesWithAttributes = relevantTelemetries.filter(
    t => t.scope!.attributes && Object.keys(t.scope!.attributes).length > 0
  ).length

  const totalAttributes = relevantTelemetries.reduce((count, t) => {
    return count + (t.scope!.attributes ? Object.keys(t.scope!.attributes).length : 0)
  }, 0)

  return {
    valid: relevantTelemetries.length > 0,
    totalTelemetries,
    relevantTelemetries: relevantTelemetries.length,
    scopesWithAttributes,
    totalAttributes,
  }
}

/**
 * Gets statistics about scope attribute data
 */
export const getScopeAttributeStats = (attributeData: ScopeAttributeData) => {
  const totalAttributes = attributeData.attributes.length
  const totalUniqueValues = attributeData.totalUniqueValues
  const averageValuesPerAttribute = totalAttributes > 0 
    ? Math.round((totalUniqueValues / totalAttributes) * 100) / 100 
    : 0

  return {
    scopeName: attributeData.scopeName,
    totalAttributes,
    totalUniqueValues,
    averageValuesPerAttribute,
  }
}
