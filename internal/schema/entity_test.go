package schema

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestGenerateEntityID(t *testing.T) {
	tests := []struct {
		name       string
		entityType string
		attributes map[string]interface{}
		want       string
	}{
		{
			name:       "service entity with basic attributes",
			entityType: "service",
			attributes: map[string]interface{}{
				"service.name":        "payment-api",
				"service.version":     "1.0.0",
				"service.instance.id": "pod-123",
			},
			want: "c8f7c7c7c7c7c7c7", // This will be a deterministic hash
		},
		{
			name:       "host entity",
			entityType: "host",
			attributes: map[string]interface{}{
				"host.name": "worker-01",
				"host.arch": "amd64",
			},
			want: "a1b2c3d4e5f6g7h8", // This will be a deterministic hash
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateEntityID(tt.entityType, tt.attributes)

			// Verify it's not empty and has expected length (16 hex chars)
			assert.NotEmpty(t, got)
			assert.Len(t, got, 16)

			// Verify deterministic - same input should produce same output
			got2 := GenerateEntityID(tt.entityType, tt.attributes)
			assert.Equal(t, got, got2)
		})
	}
}

func TestDetectEntities(t *testing.T) {
	tests := []struct {
		name       string
		attributes map[string]interface{}
		wantTypes  []string
		wantCount  int
	}{
		{
			name: "service only attributes",
			attributes: map[string]interface{}{
				"service.name":        "payment-api",
				"service.version":     "1.0.0",
				"service.instance.id": "pod-123",
			},
			wantTypes: []string{"service"},
			wantCount: 1,
		},
		{
			name: "multiple entity types",
			attributes: map[string]interface{}{
				"service.name":       "payment-api",
				"service.version":    "1.0.0",
				"k8s.pod.name":       "payment-pod-123",
				"k8s.pod.uid":        "abc-def-456",
				"k8s.namespace.name": "production",
				"host.name":          "worker-01",
				"host.arch":          "amd64",
				"container.id":       "docker-789",
			},
			wantTypes: []string{"container", "host", "k8s", "service"},
			wantCount: 4,
		},
		{
			name: "no dotted attributes",
			attributes: map[string]interface{}{
				"simple_attr": "value",
				"another":     "value2",
			},
			wantTypes: []string{"another", "simple_attr"},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create pcommon.Map from test attributes
			resourceAttrs := pcommon.NewMap()
			for key, value := range tt.attributes {
				switch v := value.(type) {
				case string:
					resourceAttrs.PutStr(key, v)
				case int:
					resourceAttrs.PutInt(key, int64(v))
				case bool:
					resourceAttrs.PutBool(key, v)
				case float64:
					resourceAttrs.PutDouble(key, v)
				}
			}

			entities := DetectEntities(resourceAttrs)

			// Check count
			assert.Len(t, entities, tt.wantCount)

			// Check entity types
			gotTypes := make([]string, len(entities))
			for i, entity := range entities {
				gotTypes[i] = entity.Type
			}

			// Sort for comparison
			assert.ElementsMatch(t, tt.wantTypes, gotTypes)

			// Verify each entity has proper structure
			for _, entity := range entities {
				assert.NotEmpty(t, entity.ID)
				assert.NotEmpty(t, entity.Type)
				assert.NotEmpty(t, entity.Attributes)
				assert.False(t, entity.FirstSeen.IsZero())
				assert.False(t, entity.LastSeen.IsZero())
			}
		})
	}
}

func TestEntityID(t *testing.T) {
	entity := Entity{
		Type: "service",
		Attributes: map[string]interface{}{
			"service.name":    "test-service",
			"service.version": "1.0.0",
		},
		FirstSeen: time.Now(),
		LastSeen:  time.Now(),
	}

	// First call should generate and set ID
	id1 := entity.EntityID()
	assert.NotEmpty(t, id1)
	assert.Equal(t, id1, entity.ID)

	// Second call should return the same ID
	id2 := entity.EntityID()
	assert.Equal(t, id1, id2)
}

func TestMergeEntity(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-1 * time.Hour)
	later := now.Add(1 * time.Hour)

	existing := Entity{
		ID:   "test-id",
		Type: "service",
		Attributes: map[string]interface{}{
			"service.name":    "test-service",
			"service.version": "1.0.0",
		},
		FirstSeen: now,
		LastSeen:  now,
	}

	new := Entity{
		ID:   "test-id",
		Type: "service",
		Attributes: map[string]interface{}{
			"service.name":        "test-service",
			"service.version":     "1.1.0",        // Updated version
			"service.instance.id": "new-instance", // New attribute
		},
		FirstSeen: earlier, // Earlier first seen
		LastSeen:  later,   // Later last seen
	}

	merged := MergeEntity(existing, new)

	// Check timestamps
	assert.Equal(t, earlier, merged.FirstSeen)
	assert.Equal(t, later, merged.LastSeen)

	// Check attributes are merged
	assert.Equal(t, "test-service", merged.Attributes["service.name"])
	assert.Equal(t, "1.1.0", merged.Attributes["service.version"])            // Updated
	assert.Equal(t, "new-instance", merged.Attributes["service.instance.id"]) // Added
}

func TestGetEntityAttributesByType(t *testing.T) {
	entities := []Entity{
		{
			Type: "service",
			Attributes: map[string]interface{}{
				"service.name":    "test-service",
				"service.version": "1.0.0",
			},
		},
		{
			Type: "host",
			Attributes: map[string]interface{}{
				"host.name": "worker-01",
				"host.arch": "amd64",
			},
		},
	}

	// Test existing entity type
	serviceAttrs := GetEntityAttributesByType(entities, "service")
	require.NotNil(t, serviceAttrs)
	assert.Equal(t, "test-service", serviceAttrs["service.name"])
	assert.Equal(t, "1.0.0", serviceAttrs["service.version"])

	// Test non-existing entity type
	containerAttrs := GetEntityAttributesByType(entities, "container")
	assert.Nil(t, containerAttrs)
}

func TestEntityTypesFromAttributes(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("service.name", "test-service")
	resourceAttrs.PutStr("service.version", "1.0.0")
	resourceAttrs.PutStr("k8s.pod.name", "test-pod")
	resourceAttrs.PutStr("host.name", "worker-01")
	resourceAttrs.PutStr("simple_attr", "value")

	entityTypes := EntityTypesFromAttributes(resourceAttrs)

	expected := []string{"host", "k8s", "service", "simple_attr"}
	assert.ElementsMatch(t, expected, entityTypes)
}

func TestConvertPCommonValue(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() pcommon.Value
		expected interface{}
	}{
		{
			name: "string value",
			setup: func() pcommon.Value {
				v := pcommon.NewValueStr("test-string")
				return v
			},
			expected: "test-string",
		},
		{
			name: "int value",
			setup: func() pcommon.Value {
				v := pcommon.NewValueInt(42)
				return v
			},
			expected: int64(42),
		},
		{
			name: "bool value",
			setup: func() pcommon.Value {
				v := pcommon.NewValueBool(true)
				return v
			},
			expected: true,
		},
		{
			name: "double value",
			setup: func() pcommon.Value {
				v := pcommon.NewValueDouble(3.14)
				return v
			},
			expected: 3.14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := tt.setup()
			result := convertPCommonValue(value)
			assert.Equal(t, tt.expected, result)
		})
	}
}
