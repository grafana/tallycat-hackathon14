package schema

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

// Entity represents an OpenTelemetry entity as defined in the OTel Entity data model
type Entity struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Attributes map[string]interface{} `json:"attributes"`
	FirstSeen  time.Time              `json:"firstSeen"`
	LastSeen   time.Time              `json:"lastSeen"`
}

// EntityID generates a unique ID for an entity using deterministic hashing
func (e *Entity) EntityID() string {
	if e.ID != "" {
		return e.ID
	}
	e.ID = GenerateEntityID(e.Type, e.Attributes)
	return e.ID
}

// GenerateEntityID creates a deterministic entity ID from type and attributes
func GenerateEntityID(entityType string, attributes map[string]interface{}) string {
	// Sort attribute keys for consistent ordering
	keys := make([]string, 0, len(attributes))
	for key := range attributes {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Build a consistent string representation
	var parts []string
	parts = append(parts, entityType) // Include entity type to avoid collisions

	for _, key := range keys {
		value := fmt.Sprintf("%v", attributes[key])
		if value != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Use xxhash for fast, deterministic hashing
	h := xxhash.New()
	h.Write([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("%x", h.Sum64())
}

// DetectEntities extracts multiple entities from resource attributes
func DetectEntities(resourceAttributes pcommon.Map) []Entity {
	entityGroups := make(map[string]map[string]interface{})
	now := time.Now()

	// Group attributes by entity type prefix
	resourceAttributes.Range(func(key string, value pcommon.Value) bool {
		parts := strings.Split(key, ".")
		if len(parts) > 0 {
			entityType := parts[0]
			if entityGroups[entityType] == nil {
				entityGroups[entityType] = make(map[string]interface{})
			}
			entityGroups[entityType][key] = convertPCommonValue(value)
		}
		return true
	})

	// Create an entity for each type found
	var entities []Entity
	for entityType, entityAttrs := range entityGroups {
		entity := Entity{
			Type:       entityType,
			Attributes: entityAttrs,
			FirstSeen:  now,
			LastSeen:   now,
		}
		entity.ID = entity.EntityID()
		entities = append(entities, entity)
	}

	return entities
}

// convertPCommonValue converts pcommon.Value to interface{} for storage
func convertPCommonValue(value pcommon.Value) interface{} {
	switch value.Type() {
	case pcommon.ValueTypeStr:
		return value.Str()
	case pcommon.ValueTypeInt:
		return value.Int()
	case pcommon.ValueTypeDouble:
		return value.Double()
	case pcommon.ValueTypeBool:
		return value.Bool()
	case pcommon.ValueTypeBytes:
		return value.Bytes().AsRaw()
	case pcommon.ValueTypeSlice:
		slice := value.Slice()
		result := make([]interface{}, slice.Len())
		for i := 0; i < slice.Len(); i++ {
			result[i] = convertPCommonValue(slice.At(i))
		}
		return result
	case pcommon.ValueTypeMap:
		m := value.Map()
		result := make(map[string]interface{})
		m.Range(func(k string, v pcommon.Value) bool {
			result[k] = convertPCommonValue(v)
			return true
		})
		return result
	default:
		return value.AsString()
	}
}

// MergeEntity merges two entities of the same type and ID, updating timestamps
func MergeEntity(existing, new Entity) Entity {
	merged := existing

	// Update timestamps
	if new.FirstSeen.Before(existing.FirstSeen) {
		merged.FirstSeen = new.FirstSeen
	}
	if new.LastSeen.After(existing.LastSeen) {
		merged.LastSeen = new.LastSeen
	}

	// Merge attributes (new attributes override existing ones)
	for key, value := range new.Attributes {
		merged.Attributes[key] = value
	}

	return merged
}

// GetEntityAttributesByType filters entity attributes by their prefix type
func GetEntityAttributesByType(entities []Entity, entityType string) map[string]interface{} {
	for _, entity := range entities {
		if entity.Type == entityType {
			return entity.Attributes
		}
	}
	return nil
}

// EntityTypesFromAttributes returns all unique entity types found in attributes
func EntityTypesFromAttributes(resourceAttributes pcommon.Map) []string {
	entityTypes := make(map[string]bool)

	resourceAttributes.Range(func(key string, value pcommon.Value) bool {
		parts := strings.Split(key, ".")
		if len(parts) > 0 {
			entityTypes[parts[0]] = true
		}
		return true
	})

	result := make([]string, 0, len(entityTypes))
	for entityType := range entityTypes {
		result = append(result, entityType)
	}
	sort.Strings(result) // For consistent ordering
	return result
}
