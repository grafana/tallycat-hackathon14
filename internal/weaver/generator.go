package weaver

import (
	"fmt"
	"strings"

	"github.com/tallycat/tallycat/internal/schema"
)

// GenerateYAML generates a Weaver format YAML string from telemetry schema data
func GenerateYAML(telemetry *schema.Telemetry, telemetrySchema *schema.TelemetrySchema) (string, error) {
	if telemetry == nil {
		return "", fmt.Errorf("telemetry cannot be nil")
	}

	// Build the YAML structure
	var yamlLines []string

	// Start with the groups section
	yamlLines = append(yamlLines, "groups:")
	yamlLines = append(yamlLines, fmt.Sprintf("  - id: metric.%s", telemetry.SchemaKey))
	yamlLines = append(yamlLines, "    type: metric")
	yamlLines = append(yamlLines, fmt.Sprintf("    metric_name: %s", telemetry.SchemaKey))

	// Add brief if available
	if telemetry.Brief != "" {
		yamlLines = append(yamlLines, fmt.Sprintf("    brief: %s", telemetry.Brief))
	}

	// Add instrument (metric type)
	yamlLines = append(yamlLines, fmt.Sprintf("    instrument: %s", strings.ToLower(string(telemetry.MetricType))))

	// Add unit
	if telemetry.MetricUnit != "" {
		yamlLines = append(yamlLines, fmt.Sprintf("    unit: %s", telemetry.MetricUnit))
	}

	// Add attributes section
	yamlLines = append(yamlLines, "    attributes:")

	// Filter and format attributes - only include DataPoint attributes as per frontend logic
	var dataPointAttributes []schema.Attribute
	var attributesToUse []schema.Attribute

	// Determine which attributes to use
	if telemetrySchema != nil && len(telemetrySchema.Attributes) > 0 {
		attributesToUse = telemetrySchema.Attributes
	} else {
		attributesToUse = telemetry.Attributes
	}

	// Filter for DataPoint source attributes
	for _, attr := range attributesToUse {
		if attr.Source == schema.AttributeSourceDataPoint {
			dataPointAttributes = append(dataPointAttributes, attr)
		}
	}

	// Format each attribute
	for _, attr := range dataPointAttributes {
		yamlLines = append(yamlLines, formatAttribute(attr)...)
	}

	// If no DataPoint attributes found, add an empty comment
	if len(dataPointAttributes) == 0 {
		yamlLines = append(yamlLines, "      # No DataPoint attributes found")
	}

	return strings.Join(yamlLines, "\n"), nil
}

// formatAttribute formats a single attribute into YAML lines
func formatAttribute(attr schema.Attribute) []string {
	var lines []string

	// Add the attribute ID
	lines = append(lines, fmt.Sprintf("      - id: %s", attr.Name))

	// Add the attribute type - convert from internal type to Weaver type
	weaverType := convertAttributeType(attr.Type)
	lines = append(lines, fmt.Sprintf("        type: %s", weaverType))

	// Add requirement level - default to recommended as per frontend
	requirementLevel := "recommended"
	if attr.RequirementLevel != "" {
		requirementLevel = strings.ToLower(string(attr.RequirementLevel))
	}
	lines = append(lines, fmt.Sprintf("        requirement_level: %s", requirementLevel))

	// Add brief if available
	if attr.Brief != "" {
		lines = append(lines, fmt.Sprintf("        brief: %s", attr.Brief))
	}

	return lines
}

// convertAttributeType converts internal attribute types to Weaver-compatible types
func convertAttributeType(attrType schema.AttributeType) string {
	switch attrType {
	case schema.AttributeTypeStr:
		return "string"
	case schema.AttributeTypeBool:
		return "boolean"
	case schema.AttributeTypeInt:
		return "int"
	case schema.AttributeTypeDouble:
		return "double"
	case schema.AttributeTypeMap:
		return "string" // Maps are typically represented as strings in Weaver
	case schema.AttributeTypeSlice:
		return "string[]" // Arrays of strings
	case schema.AttributeTypeBytes:
		return "string"
	case schema.AttributeTypeEmpty:
		return "string" // Default to string for empty/unknown types
	default:
		return "string" // Default fallback
	}
}
