// Entity Attribute Aggregation Utilities
// Functions to aggregate entity attributes from telemetry data for the Interactive Entity Catalog

import type { Telemetry } from '@/types/telemetry'
import type { EntityAttribute, EntityAttributeData } from '@/types/entity-attributes'

/**
 * Aggregates entity attributes for a specific entity type from all telemetries
 * 
 * @param telemetries Array of telemetry objects with populated entities
 * @param entityType The entity type to aggregate attributes for (e.g., "host", "container")
 * @returns Array of EntityAttribute objects with aggregated values
 */
export const aggregateEntityAttributes = (
  telemetries: Telemetry[],
  entityType: string
): EntityAttribute[] => {
  // Use Map to collect unique values for each attribute name
  const attributeMap = new Map<string, Set<string>>()
  
  // Iterate through all telemetries and their entities
  for (const telemetry of telemetries) {
    if (!telemetry.entities) continue
    
    // Check each entity in the telemetry
    for (const entity of Object.values(telemetry.entities)) {
      if (entity.type !== entityType) continue
      
      // Process each attribute of the matching entity
      for (const [attributeName, attributeValue] of Object.entries(entity.attributes || {})) {
        // Skip empty or null values
        if (attributeValue === null || attributeValue === undefined || attributeValue === '') {
          continue
        }
        
        // Initialize Set for this attribute if not exists
        if (!attributeMap.has(attributeName)) {
          attributeMap.set(attributeName, new Set<string>())
        }
        
        // Add the value to the Set (automatically handles deduplication)
        attributeMap.get(attributeName)!.add(String(attributeValue))
      }
    }
  }
  
  // Convert Map to EntityAttribute array with sorted values
  const attributes: EntityAttribute[] = Array.from(attributeMap.entries())
    .map(([name, valueSet]) => ({
      name,
      values: Array.from(valueSet).sort().join(', ')
    }))
    .sort((a, b) => a.name.localeCompare(b.name)) // Sort attributes by name
  
  return attributes
}

/**
 * Creates complete EntityAttributeData for a specific entity type
 * 
 * @param telemetries Array of telemetry objects
 * @param entityType The entity type to create data for
 * @returns Complete EntityAttributeData object
 */
export const createEntityAttributeData = (
  telemetries: Telemetry[],
  entityType: string
): EntityAttributeData => {
  const attributes = aggregateEntityAttributes(telemetries, entityType)
  
  return {
    entityType,
    attributes
  }
}

/**
 * Aggregates attributes for all entity types found in telemetries
 * 
 * @param telemetries Array of telemetry objects
 * @returns Map of entity type to EntityAttributeData
 */
export const aggregateAllEntityAttributes = (
  telemetries: Telemetry[]
): Map<string, EntityAttributeData> => {
  // First, collect all unique entity types
  const entityTypes = new Set<string>()
  
  for (const telemetry of telemetries) {
    if (!telemetry.entities) continue
    
    for (const entity of Object.values(telemetry.entities)) {
      entityTypes.add(entity.type)
    }
  }
  
  // Create attribute data for each entity type
  const result = new Map<string, EntityAttributeData>()
  
  for (const entityType of entityTypes) {
    const attributeData = createEntityAttributeData(telemetries, entityType)
    result.set(entityType, attributeData)
  }
  
  return result
}

/**
 * Validates that telemetries have the required structure for attribute aggregation
 * 
 * @param telemetries Array of telemetry objects
 * @returns Validation result with any issues found
 */
export const validateTelemetriesForAttributes = (telemetries: Telemetry[]) => {
  const issues: string[] = []
  let totalEntities = 0
  let entitiesWithAttributes = 0
  
  for (const telemetry of telemetries) {
    if (!telemetry.entities) {
      issues.push(`Telemetry ${telemetry.schemaKey} has no entities`)
      continue
    }
    
    for (const entity of Object.values(telemetry.entities)) {
      totalEntities++
      
      if (!entity.type) {
        issues.push(`Entity ${entity.id} has no type`)
      }
      
      if (entity.attributes && Object.keys(entity.attributes).length > 0) {
        entitiesWithAttributes++
      }
    }
  }
  
  return {
    isValid: issues.length === 0,
    issues,
    stats: {
      totalEntities,
      entitiesWithAttributes,
      coveragePercentage: totalEntities > 0 ? (entitiesWithAttributes / totalEntities) * 100 : 0
    }
  }
}

/**
 * Gets summary statistics for entity attributes
 * 
 * @param attributeData EntityAttributeData object
 * @returns Summary statistics
 */
export const getEntityAttributeStats = (attributeData: EntityAttributeData) => {
  const totalAttributes = attributeData.attributes.length
  const attributesWithMultipleValues = attributeData.attributes.filter(
    attr => attr.values.includes(',')
  ).length
  
  const totalUniqueValues = attributeData.attributes.reduce((sum, attr) => {
    return sum + attr.values.split(', ').length
  }, 0)
  
  const averageValuesPerAttribute = totalAttributes > 0 
    ? Math.round((totalUniqueValues / totalAttributes) * 10) / 10 
    : 0
  
  return {
    totalAttributes,
    attributesWithMultipleValues,
    totalUniqueValues,
    averageValuesPerAttribute,
    diversityPercentage: totalAttributes > 0 
      ? Math.round((attributesWithMultipleValues / totalAttributes) * 100) 
      : 0
  }
}
