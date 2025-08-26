package weaver

import (
	"strings"
	"testing"

	"github.com/tallycat/tallycat/internal/schema"
)

func TestGenerateYAML_NilTelemetry(t *testing.T) {

	yaml, err := GenerateYAML(nil, nil)

	if err == nil {
		t.Error("Expected error for nil telemetry, got nil")
	}
	if yaml != "" {
		t.Error("Expected empty YAML for nil telemetry")
	}
}

func TestGenerateYAML_BasicTelemetry(t *testing.T) {

	telemetry := &schema.Telemetry{
		SchemaKey:     "http.server.duration",
		Brief:         "Measures the duration of HTTP server requests",
		MetricType:    schema.MetricTypeHistogram,
		MetricUnit:    "ms",
		TelemetryType: schema.TelemetryTypeMetric,
		Attributes: []schema.Attribute{
			{
				Name:             "http.method",
				Type:             schema.AttributeTypeStr,
				Source:           schema.AttributeSourceDataPoint,
				RequirementLevel: schema.RequirementLevelRequired,
				Brief:            "HTTP request method",
			},
			{
				Name:   "service.name",
				Type:   schema.AttributeTypeStr,
				Source: schema.AttributeSourceResource, // Should be filtered out
			},
		},
	}

	yaml, err := GenerateYAML(telemetry, nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the basic structure
	expectedLines := []string{
		"groups:",
		"  - id: metric.http.server.duration",
		"    type: metric",
		"    metric_name: http.server.duration",
		"    brief: Measures the duration of HTTP server requests",
		"    instrument: histogram",
		"    unit: ms",
		"    attributes:",
		"      - id: http.method",
		"        type: string",
		"        requirement_level: required",
		"        brief: HTTP request method",
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(yaml, expectedLine) {
			t.Errorf("Expected YAML to contain '%s', but it didn't.\nActual YAML:\n%s", expectedLine, yaml)
		}
	}

	// Verify that resource attributes are filtered out
	if strings.Contains(yaml, "service.name") {
		t.Error("YAML should not contain resource attributes")
	}
}

func TestGenerateYAML_WithTelemetrySchema(t *testing.T) {

	telemetry := &schema.Telemetry{
		SchemaKey:     "test.metric",
		MetricType:    schema.MetricTypeGauge,
		MetricUnit:    "1",
		Brief:         "Test metric",
		TelemetryType: schema.TelemetryTypeMetric,
		Attributes:    []schema.Attribute{}, // Empty in telemetry
	}

	telemetrySchema := &schema.TelemetrySchema{
		Attributes: []schema.Attribute{
			{
				Name:   "status_code",
				Type:   schema.AttributeTypeInt,
				Source: schema.AttributeSourceDataPoint,
			},
		},
	}

	yaml, err := GenerateYAML(telemetry, telemetrySchema)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use attributes from telemetrySchema when available
	if !strings.Contains(yaml, "status_code") {
		t.Error("YAML should contain attributes from telemetrySchema")
	}
	if !strings.Contains(yaml, "type: int") {
		t.Error("YAML should contain converted integer type")
	}
}

func TestGenerateYAML_NoDataPointAttributes(t *testing.T) {

	telemetry := &schema.Telemetry{
		SchemaKey:     "test.metric",
		MetricType:    schema.MetricTypeSum,
		MetricUnit:    "bytes",
		TelemetryType: schema.TelemetryTypeMetric,
		Attributes: []schema.Attribute{
			{
				Name:   "service.name",
				Type:   schema.AttributeTypeStr,
				Source: schema.AttributeSourceResource, // Not DataPoint
			},
			{
				Name:   "library.name",
				Type:   schema.AttributeTypeStr,
				Source: schema.AttributeSourceScope, // Not DataPoint
			},
		},
	}

	yaml, err := GenerateYAML(telemetry, nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should contain the comment for no DataPoint attributes
	if !strings.Contains(yaml, "# No DataPoint attributes found") {
		t.Error("YAML should contain comment for no DataPoint attributes")
	}
}

func TestGenerateYAML_EmptyValues(t *testing.T) {

	telemetry := &schema.Telemetry{
		SchemaKey:     "minimal.metric",
		MetricType:    schema.MetricTypeGauge,
		TelemetryType: schema.TelemetryTypeMetric,
		// Brief and MetricUnit are empty
		Attributes: []schema.Attribute{
			{
				Name:   "test.attr",
				Type:   schema.AttributeTypeStr,
				Source: schema.AttributeSourceDataPoint,
				// Brief and RequirementLevel are empty
			},
		},
	}

	yaml, err := GenerateYAML(telemetry, nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should handle empty values gracefully
	expectedLines := []string{
		"metric_name: minimal.metric",
		"instrument: gauge",
		"requirement_level: recommended",
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(yaml, expectedLine) {
			t.Errorf("Expected YAML to contain '%s', but it didn't.\nActual YAML:\n%s", expectedLine, yaml)
		}
	}
}
