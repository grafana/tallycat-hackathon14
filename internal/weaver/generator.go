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

	brief := telemetry.Brief
	if brief == "" {
		brief = `""`
	}
	yamlLines = append(yamlLines, fmt.Sprintf("    brief: %s", brief))

	// Add instrument (metric type)
	yamlLines = append(yamlLines, fmt.Sprintf("    instrument: %s", convertMetricTypeToInstrument(telemetry.MetricType)))

	// Add unit - always include even if empty (required by Weaver schema)
	unit := telemetry.MetricUnit
	if unit == "" {
		unit = `""`
	}
	yamlLines = append(yamlLines, fmt.Sprintf("    unit: %s", unit))

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

	// Only add attributes section if there are DataPoint attributes
	if len(dataPointAttributes) > 0 {
		yamlLines = append(yamlLines, "    attributes:")

		// Format each attribute
		for _, attr := range dataPointAttributes {
			yamlLines = append(yamlLines, formatAttribute(attr)...)
		}
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

	// Add brief - always include even if empty (required by Weaver schema)
	brief := attr.Brief
	if brief == "" {
		brief = `""`
	}
	lines = append(lines, fmt.Sprintf("        brief: %s", brief))

	return lines
}

// GenerateMultiMetricYAML generates a Weaver format YAML string from multiple telemetry schema data
func GenerateMultiMetricYAML(telemetries []schema.Telemetry, schemas map[string]*schema.TelemetrySchema) (string, error) {
	if len(telemetries) == 0 {
		return "groups: []", nil
	}

	// Build the YAML structure
	var yamlLines []string

	// Start with the groups section
	yamlLines = append(yamlLines, "groups:")

	// Process each telemetry
	for _, telemetry := range telemetries {
		// Get the corresponding schema if available
		var telemetrySchema *schema.TelemetrySchema
		if schemas != nil {
			telemetrySchema = schemas[telemetry.SchemaID]
		}

		// Generate YAML for this single metric using the existing function
		singleYAML, err := GenerateYAML(&telemetry, telemetrySchema)
		if err != nil {
			return "", fmt.Errorf("failed to generate YAML for metric %s: %w", telemetry.SchemaKey, err)
		}

		// Parse the single YAML and extract the group content
		lines := strings.Split(singleYAML, "\n")

		// Skip the "groups:" line and add the group content
		for i, line := range lines {
			if i == 0 && strings.TrimSpace(line) == "groups:" {
				continue // Skip the groups header
			}
			yamlLines = append(yamlLines, line)
		}
	}

	return strings.Join(yamlLines, "\n"), nil
}

// convertMetricTypeToInstrument converts internal metric types to OpenTelemetry Weaver instrument names
func convertMetricTypeToInstrument(metricType schema.MetricType) string {
	switch metricType {
	case schema.MetricTypeGauge:
		return "gauge"
	case schema.MetricTypeSum:
		return "counter" // Sum metrics are represented as counters in Weaver
	case schema.MetricTypeHistogram:
		return "histogram"
	case schema.MetricTypeExponentialHistogram:
		return "histogram" // ExponentialHistogram is still a histogram in Weaver
	case schema.MetricTypeSummary:
		return "histogram" // Summary is typically represented as histogram in Weaver
	case schema.MetricTypeEmpty:
		return "gauge" // Default to gauge for empty/unknown types
	default:
		return "gauge" // Default fallback
	}
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
