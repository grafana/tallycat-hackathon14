package schema

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

// Scope represents an OpenTelemetry scope as defined in the OTel Scope data model
type Scope struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Version    string                 `json:"version"`
	SchemaURL  string                 `json:"schemaURL"`
	Attributes map[string]interface{} `json:"attributes"`
	FirstSeen  time.Time              `json:"firstSeen"`
	LastSeen   time.Time              `json:"lastSeen"`
}

// ScopeID generates a unique ID for an scope using deterministic hashing
func (s *Scope) ScopeID() string {
	if s.ID != "" {
		return s.ID
	}
	s.ID = GenerateScopeID(s.Name, s.Version, s.SchemaURL, s.Attributes)
	return s.ID
}

// GenerateEntityID creates a deterministic entity ID from type and attributes
func GenerateScopeID(name string, version string, schemaURL string, attributes map[string]interface{}) string {
	// Sort attribute keys for consistent ordering
	keys := make([]string, 0, len(attributes))
	for key := range attributes {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Build a consistent string representation
	var parts []string
	parts = append(parts, name)
	parts = append(parts, version)
	parts = append(parts, schemaURL)

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

// DetectScopes extracts scopes from an InstrumentationScope object
func DetectScopes(scope pcommon.InstrumentationScope, schemaURL string) Scope {
	now := time.Now()

	// Convert scope attributes to map using the same function from entity.go
	attributes := make(map[string]interface{})
	scope.Attributes().Range(func(key string, value pcommon.Value) bool {
		// Use the same conversion logic as in entity.go
		switch value.Type() {
		case pcommon.ValueTypeStr:
			attributes[key] = value.Str()
		case pcommon.ValueTypeInt:
			attributes[key] = value.Int()
		case pcommon.ValueTypeDouble:
			attributes[key] = value.Double()
		case pcommon.ValueTypeBool:
			attributes[key] = value.Bool()
		case pcommon.ValueTypeBytes:
			attributes[key] = value.Bytes().AsRaw()
		default:
			attributes[key] = value.AsString()
		}
		return true
	})

	// Use default name if scope name is empty
	scopeName := scope.Name()
	if scopeName == "" {
		scopeName = "UNKNOWN"
	}

	// Transform the scope name to human-readable format
	transformer := NewScopeTransformer()
	transformedScopeName := transformer.Transform(scopeName)

	// Use default version if scope version is empty
	scopeVersion := scope.Version()
	if scopeVersion == "" {
		scopeVersion = "UNKNOWN"
	}

	// Create scope object with transformed name
	s := Scope{
		Name:       transformedScopeName,
		Version:    scopeVersion,
		SchemaURL:  schemaURL,
		Attributes: attributes,
		FirstSeen:  now,
		LastSeen:   now,
	}
	s.ID = s.ScopeID()

	return s
}
