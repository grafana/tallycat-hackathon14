package schema

import (
	"fmt"
	"strings"
)

// SchemaID represents a unique identifier for a schema
type SchemaID struct {
	SignalType   string
	ScopeName    string
	ScopeVersion string
	Hash         uint64
}

// String returns the string representation of the schema ID
func (s *SchemaID) String() string {
	return fmt.Sprintf("%s:%s:%s:%x", s.SignalType, s.ScopeName, s.ScopeVersion, s.Hash)
}

// ParseSchemaID parses a schema ID string into its components
func ParseSchemaID(id string) (*SchemaID, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid schema ID format: %s", id)
	}

	hash, err := parseHexUint64(parts[3])
	if err != nil {
		return nil, fmt.Errorf("invalid hash in schema ID: %w", err)
	}

	return &SchemaID{
		SignalType:   parts[0],
		ScopeName:    parts[1],
		ScopeVersion: parts[2],
		Hash:         hash,
	}, nil
}

// parseHexUint64 parses a hex string into a uint64
func parseHexUint64(s string) (uint64, error) {
	var result uint64
	_, err := fmt.Sscanf(s, "%x", &result)
	return result, err
}
