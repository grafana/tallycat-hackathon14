package weaver

import (
	"fmt"
	"strings"

	"github.com/tallycat/tallycat/internal/schema"
)

// quoteYAMLString ensures a string is properly quoted for YAML output
// Always quotes strings to avoid issues with special characters like colons
func quoteYAMLString(s string) string {
	if s == "" {
		return `""`
	}
	// Always quote to avoid YAML parsing issues with special characters
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(s, `"`, `\"`))
}

// buildGroupID creates a group ID in the format: {telemetryType}.{sanitized_scope_name}.{schemaKey}
// Falls back to "unknown" for scope name when telemetry.Scope is nil or scope name is empty/UNKNOWN
func buildGroupID(telemetryType string, telemetry *schema.Telemetry) string {
	var scopeName string
	if telemetry.Scope != nil && telemetry.Scope.Name != "" {
		scopeName = schema.SanitizeScopeName(telemetry.Scope.Name)
	} else {
		scopeName = "unknown"
	}

	return fmt.Sprintf("%s.%s.%s", telemetryType, scopeName, telemetry.SchemaKey)
}

// GenerateYAML generates a Weaver format YAML string from telemetry schema data
func GenerateYAML(telemetry *schema.Telemetry, telemetrySchema *schema.TelemetrySchema) (string, error) {
	if telemetry == nil {
		return "", fmt.Errorf("telemetry cannot be nil")
	}

	// Generate different YAML based on telemetry type
	switch telemetry.TelemetryType {
	case schema.TelemetryTypeLog:
		return generateLogEventYAML(telemetry, telemetrySchema)
	case schema.TelemetryTypeMetric:
		return generateMetricYAML(telemetry, telemetrySchema)
	case schema.TelemetryTypeSpan:
		return generateSpanYAML(telemetry, telemetrySchema)
	default:
		// Default to metric for backwards compatibility
		return generateMetricYAML(telemetry, telemetrySchema)
	}
}

// generateMetricYAML generates YAML for metric telemetries (existing logic)
func generateMetricYAML(telemetry *schema.Telemetry, telemetrySchema *schema.TelemetrySchema) (string, error) {
	// Build the YAML structure
	var yamlLines []string

	// Start with the groups section
	yamlLines = append(yamlLines, "groups:")
	yamlLines = append(yamlLines, fmt.Sprintf("  - id: %s", buildGroupID("metric", telemetry)))
	yamlLines = append(yamlLines, "    type: metric")
	yamlLines = append(yamlLines, fmt.Sprintf("    metric_name: %s", telemetry.SchemaKey))

	yamlLines = append(yamlLines, fmt.Sprintf("    brief: %s", quoteYAMLString(telemetry.Brief)))
	yamlLines = append(yamlLines, "    stability: stable")

	// Add instrument (metric type)
	yamlLines = append(yamlLines, fmt.Sprintf("    instrument: %s", convertMetricTypeToInstrument(telemetry.MetricType)))

	// Add unit - always include even if empty (required by Weaver schema)
	yamlLines = append(yamlLines, fmt.Sprintf("    unit: %s", quoteYAMLString(telemetry.MetricUnit)))

	// Filter and format attributes - only include DataPoint attributes as per frontend logic
	dataPointAttributes := make(map[string]schema.Attribute)
	var attributesToUse []schema.Attribute

	// Determine which attributes to use
	if telemetrySchema != nil && len(telemetrySchema.Attributes) > 0 {
		attributesToUse = telemetrySchema.Attributes
	} else {
		attributesToUse = telemetry.Attributes
	}

	// Filter for DataPoint source attributes
	for _, attr := range attributesToUse {
		switch attr.Source {
		case schema.AttributeSourceResource, schema.AttributeSourceScope:
			attr.RequirementLevel = schema.RequirementLevelRequired
			fallthrough
		case schema.AttributeSourceDataPoint:
			_, ok := dataPointAttributes[attr.Name]
			if !ok {
				dataPointAttributes[attr.Name] = attr
				continue
			}
			if attr.RequirementLevel == schema.RequirementLevelRequired {
				dataPointAttributes[attr.Name] = attr
			}
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

// generateLogEventYAML generates YAML for log telemetries as events
func generateLogEventYAML(telemetry *schema.Telemetry, telemetrySchema *schema.TelemetrySchema) (string, error) {
	// Build the YAML structure
	var yamlLines []string

	// Determine event name: use LogEventName if available, otherwise use SchemaKey
	eventName := telemetry.LogEventName
	if eventName == "" {
		eventName = telemetry.SchemaKey
	}

	// Start with the groups section
	yamlLines = append(yamlLines, "groups:")
	yamlLines = append(yamlLines, fmt.Sprintf("  - id: %s", buildGroupID("event", telemetry)))
	yamlLines = append(yamlLines, "    type: event")
	yamlLines = append(yamlLines, fmt.Sprintf("    name: %s", eventName))
	yamlLines = append(yamlLines, fmt.Sprintf("    brief: %s", quoteYAMLString(telemetry.Brief)))
	yamlLines = append(yamlLines, "    stability: stable")

	// Collect all attributes (LogRecord source + log-specific attributes)
	var allAttributes []schema.Attribute
	var attributesToUse []schema.Attribute

	// Determine which attributes to use
	if telemetrySchema != nil && len(telemetrySchema.Attributes) > 0 {
		attributesToUse = telemetrySchema.Attributes
	} else {
		attributesToUse = telemetry.Attributes
	}

	// Filter for LogRecord source attributes
	for _, attr := range attributesToUse {
		if attr.Source == schema.AttributeSourceLogRecord {
			allAttributes = append(allAttributes, attr)
		}
		if attr.Source == schema.AttributeSourceResource || attr.Source == schema.AttributeSourceScope {
			attr.RequirementLevel = schema.RequirementLevelRequired
			allAttributes = append(allAttributes, attr)
		}
	}

	// Add log-specific attributes
	logAttributes := generateLogSpecificAttributes(telemetry)
	allAttributes = append(allAttributes, logAttributes...)

	// Always add attributes section for events (at minimum log severity attributes)
	yamlLines = append(yamlLines, "    attributes:")

	// Format each attribute
	for _, attr := range allAttributes {
		yamlLines = append(yamlLines, formatAttribute(attr)...)
	}

	return strings.Join(yamlLines, "\n"), nil
}

// generateSpanYAML generates YAML for span telemetries following SpanSemanticConvention
func generateSpanYAML(telemetry *schema.Telemetry, telemetrySchema *schema.TelemetrySchema) (string, error) {
	// Build the YAML structure
	var yamlLines []string

	// Start with the groups section
	yamlLines = append(yamlLines, "groups:")
	yamlLines = append(yamlLines, fmt.Sprintf("  - id: %s", buildGroupID("span", telemetry)))
	yamlLines = append(yamlLines, "    type: span")
	yamlLines = append(yamlLines, fmt.Sprintf("    brief: %s", quoteYAMLString(telemetry.Brief)))
	yamlLines = append(yamlLines, "    stability: stable")

	// Add span_kind (required for SpanSemanticConvention)
	spanKind := convertSpanKindToWeaver(telemetry.SpanKind)
	yamlLines = append(yamlLines, fmt.Sprintf("    span_kind: %s", spanKind))

	// Collect span attributes
	var spanAttributes []schema.Attribute
	var attributesToUse []schema.Attribute

	// Determine which attributes to use
	if telemetrySchema != nil && len(telemetrySchema.Attributes) > 0 {
		attributesToUse = telemetrySchema.Attributes
	} else {
		attributesToUse = telemetry.Attributes
	}

	// Filter for Span source attributes (span-level attributes)
	for _, attr := range attributesToUse {
		if attr.Source == schema.AttributeSourceSpan {
			spanAttributes = append(spanAttributes, attr)
		}
		if attr.Source == schema.AttributeSourceResource || attr.Source == schema.AttributeSourceScope {
			attr.RequirementLevel = schema.RequirementLevelRequired
			spanAttributes = append(spanAttributes, attr)
		}
	}

	// Only add attributes section if there are span attributes
	if len(spanAttributes) > 0 {
		yamlLines = append(yamlLines, "    attributes:")

		// Format each attribute
		for _, attr := range spanAttributes {
			yamlLines = append(yamlLines, formatAttribute(attr)...)
		}
	}

	return strings.Join(yamlLines, "\n"), nil
}

// generateLogSpecificAttributes creates standard log attributes from telemetry fields
func generateLogSpecificAttributes(telemetry *schema.Telemetry) []schema.Attribute {
	var attributes []schema.Attribute

	// Always add log severity number
	attributes = append(attributes, schema.Attribute{
		Name:             "log.severity.number",
		Type:             schema.AttributeTypeInt,
		RequirementLevel: schema.RequirementLevelRecommended,
		Brief:            "Log severity number",
		Source:           schema.AttributeSourceLogRecord,
	})

	// Always add log severity text
	attributes = append(attributes, schema.Attribute{
		Name:             "log.severity.text",
		Type:             schema.AttributeTypeStr,
		RequirementLevel: schema.RequirementLevelRecommended,
		Brief:            "Log severity text",
		Source:           schema.AttributeSourceLogRecord,
	})

	return attributes
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

	lines = append(lines, "        stability: stable")

	// Add brief - always include even if empty (required by Weaver schema)
	lines = append(lines, fmt.Sprintf("        brief: %s", quoteYAMLString(attr.Brief)))

	return lines
}

// GenerateMultiMetricYAML generates a Weaver format YAML string from multiple metric telemetry schema data
// Only processes telemetries with TelemetryType = TelemetryTypeMetric
func GenerateMultiMetricYAML(telemetries []schema.Telemetry, schemas map[string]*schema.TelemetrySchema) (string, error) {
	// Filter for metrics only
	var metricTelemetries []schema.Telemetry
	for _, telemetry := range telemetries {
		if telemetry.TelemetryType == schema.TelemetryTypeMetric {
			metricTelemetries = append(metricTelemetries, telemetry)
		}
	}

	return generateMultiTelemetryYAML(metricTelemetries, schemas)
}

// GenerateMultiLogYAML generates a Weaver format YAML string from multiple log telemetry schema data
// Only processes telemetries with TelemetryType = TelemetryTypeLog
func GenerateMultiLogYAML(telemetries []schema.Telemetry, schemas map[string]*schema.TelemetrySchema) (string, error) {
	// Filter for logs only
	var logTelemetries []schema.Telemetry
	for _, telemetry := range telemetries {
		if telemetry.TelemetryType == schema.TelemetryTypeLog {
			logTelemetries = append(logTelemetries, telemetry)
		}
	}

	return generateMultiTelemetryYAML(logTelemetries, schemas)
}

// GenerateMultiSpanYAML generates a Weaver format YAML string from multiple span telemetry schema data
// Only processes telemetries with TelemetryType = TelemetryTypeSpan
func GenerateMultiSpanYAML(telemetries []schema.Telemetry, schemas map[string]*schema.TelemetrySchema) (string, error) {
	// Filter for spans only
	var spanTelemetries []schema.Telemetry
	for _, telemetry := range telemetries {
		if telemetry.TelemetryType == schema.TelemetryTypeSpan {
			spanTelemetries = append(spanTelemetries, telemetry)
		}
	}

	return generateMultiTelemetryYAML(spanTelemetries, schemas)
}

// generateMultiTelemetryYAML is a helper function that generates YAML from a list of telemetries
// This function does not filter by type - it processes all provided telemetries
func generateMultiTelemetryYAML(telemetries []schema.Telemetry, schemas map[string]*schema.TelemetrySchema) (string, error) {
	if len(telemetries) == 0 {
		return "", nil
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

		// Generate YAML for this single telemetry using the existing function
		singleYAML, err := GenerateYAML(&telemetry, telemetrySchema)
		if err != nil {
			return "", fmt.Errorf("failed to generate YAML for telemetry %s: %w", telemetry.SchemaKey, err)
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

// GenerateRegistryManifest generates a registry_manifest.yaml content for a producer
func GenerateRegistryManifest(producerName, producerVersion string) string {
	var lines []string

	lines = append(lines, fmt.Sprintf("name: %s", producerName))
	lines = append(lines, fmt.Sprintf("description: Schema for %s, version %s", producerName, producerVersion))
	lines = append(lines, fmt.Sprintf("semconv_version: %s", producerVersion))
	lines = append(lines, fmt.Sprintf("schema_base_url: http://github.com/nicolastakashi/tallycat/%s---%s", producerName, producerVersion))

	return strings.Join(lines, "\n") + "\n"
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

// convertSpanKindToWeaver converts internal span kinds to Weaver-compatible span_kind values
func convertSpanKindToWeaver(spanKind schema.SpanKind) string {
	switch spanKind {
	case schema.SpanKindClient:
		return "client"
	case schema.SpanKindServer:
		return "server"
	case schema.SpanKindProducer:
		return "producer"
	case schema.SpanKindConsumer:
		return "consumer"
	case schema.SpanKindInternal:
		return "internal"
	default:
		return "internal" // Default to internal for unknown span kinds
	}
}
