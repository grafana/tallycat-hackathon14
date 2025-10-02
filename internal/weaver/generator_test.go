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
		Scope: &schema.Scope{
			Name:      "django",
			Version:   "1.0.0",
			SchemaURL: "https://opentelemetry.io/schemas/1.0.0",
			Attributes: map[string]interface{}{
				"service.name": "django",
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
		"  - id: metric.django.http.server.duration",
		"    type: metric",
		"    metric_name: http.server.duration",
		"    brief: \"Measures the duration of HTTP server requests\"",
		"    stability: stable",
		"    instrument: histogram",
		"    unit: \"ms\"",
		"    attributes:",
		"      - id: http.method",
		"        type: string",
		"        requirement_level: required",
		"        stability: stable",
		"        brief: \"HTTP request method\"",
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(yaml, expectedLine) {
			t.Errorf("Expected YAML to contain '%s', but it didn't.\nActual YAML:\n%s", expectedLine, yaml)
		}
	}

	// Verify that resource attributes are now included
	if !strings.Contains(yaml, "service.name") {
		t.Error("YAML should contain resource attributes")
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

	// Should contain attributes section with resource and scope attributes
	if !strings.Contains(yaml, "attributes:") {
		t.Error("YAML should contain attributes section with resource and scope attributes")
	}

	// Should contain resource and scope attributes
	if !strings.Contains(yaml, "service.name") {
		t.Error("YAML should contain resource attributes")
	}
	if !strings.Contains(yaml, "library.name") {
		t.Error("YAML should contain scope attributes")
	}

	// Should not contain the old comment
	if strings.Contains(yaml, "# No DataPoint attributes found") {
		t.Error("YAML should not contain the old comment for no DataPoint attributes")
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

func TestGenerateMultiMetricYAML_EmptyTelemetries(t *testing.T) {
	yaml, err := GenerateMultiMetricYAML([]schema.Telemetry{}, map[string]*schema.TelemetrySchema{})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := ""
	if yaml != expected {
		t.Errorf("Expected YAML to be '%s', got '%s'", expected, yaml)
	}
}

func TestGenerateMultiMetricYAML_SingleMetric(t *testing.T) {
	telemetries := []schema.Telemetry{
		{
			SchemaID:      "metric1_schema_id",
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
			},
			Scope: &schema.Scope{
				Name:      "django",
				Version:   "1.0.0",
				SchemaURL: "https://opentelemetry.io/schemas/1.0.0",
				Attributes: map[string]interface{}{
					"service.name": "django",
				},
			},
		},
	}

	schemas := map[string]*schema.TelemetrySchema{}

	yaml, err := GenerateMultiMetricYAML(telemetries, schemas)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the basic structure
	expectedLines := []string{
		"groups:",
		"  - id: metric.django.http.server.duration",
		"    type: metric",
		"    metric_name: http.server.duration",
		"    brief: \"Measures the duration of HTTP server requests\"",
		"    stability: stable",
		"    instrument: histogram",
		"    unit: \"ms\"",
		"    attributes:",
		"      - id: http.method",
		"        type: string",
		"        requirement_level: required",
		"        stability: stable",
		"        brief: \"HTTP request method\"",
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(yaml, expectedLine) {
			t.Errorf("Expected YAML to contain '%s', but it didn't.\nActual YAML:\n%s", expectedLine, yaml)
		}
	}
}

func TestGenerateMultiMetricYAML_MultipleMetrics(t *testing.T) {
	telemetries := []schema.Telemetry{
		{
			SchemaID:      "metric1_schema_id",
			SchemaKey:     "http.server.duration",
			Brief:         "HTTP server request duration",
			MetricType:    schema.MetricTypeHistogram,
			MetricUnit:    "ms",
			TelemetryType: schema.TelemetryTypeMetric,
			Attributes: []schema.Attribute{
				{
					Name:   "http.method",
					Type:   schema.AttributeTypeStr,
					Source: schema.AttributeSourceDataPoint,
				},
			},
			Scope: &schema.Scope{
				Name:      "django",
				Version:   "1.0.0",
				SchemaURL: "https://opentelemetry.io/schemas/1.0.0",
				Attributes: map[string]interface{}{
					"service.name": "django",
				},
			},
		},
		{
			SchemaID:      "metric2_schema_id",
			SchemaKey:     "http.server.requests",
			Brief:         "HTTP server request count",
			MetricType:    schema.MetricTypeSum,
			MetricUnit:    "1",
			TelemetryType: schema.TelemetryTypeMetric,
			Attributes: []schema.Attribute{
				{
					Name:   "http.status_code",
					Type:   schema.AttributeTypeInt,
					Source: schema.AttributeSourceDataPoint,
				},
			},
			Scope: &schema.Scope{
				Name:      "django",
				Version:   "1.0.0",
				SchemaURL: "https://opentelemetry.io/schemas/1.0.0",
				Attributes: map[string]interface{}{
					"service.name": "django",
				},
			},
		},
	}

	schemas := map[string]*schema.TelemetrySchema{}

	yaml, err := GenerateMultiMetricYAML(telemetries, schemas)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should contain both metrics
	expectedLines := []string{
		"groups:",
		"  - id: metric.django.http.server.duration",
		"    instrument: histogram",
		"  - id: metric.django.http.server.requests",
		"    instrument: counter",
		"      - id: http.method",
		"      - id: http.status_code",
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(yaml, expectedLine) {
			t.Errorf("Expected YAML to contain '%s', but it didn't.\nActual YAML:\n%s", expectedLine, yaml)
		}
	}
}

func TestGenerateMultiMetricYAML_WithSchemas(t *testing.T) {
	telemetries := []schema.Telemetry{
		{
			SchemaID:      "metric1_schema_id",
			SchemaKey:     "test.metric",
			MetricType:    schema.MetricTypeGauge,
			MetricUnit:    "1",
			Brief:         "Test metric",
			TelemetryType: schema.TelemetryTypeMetric,
			Attributes:    []schema.Attribute{}, // Empty in telemetry
		},
	}

	schemas := map[string]*schema.TelemetrySchema{
		"metric1_schema_id": {
			SchemaId: "metric1_schema_id",
			Attributes: []schema.Attribute{
				{
					Name:   "custom_attribute",
					Type:   schema.AttributeTypeStr,
					Source: schema.AttributeSourceDataPoint,
				},
			},
		},
	}

	yaml, err := GenerateMultiMetricYAML(telemetries, schemas)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use attributes from schema
	if !strings.Contains(yaml, "custom_attribute") {
		t.Error("YAML should contain attributes from telemetrySchema")
	}
}

func TestConvertMetricTypeToInstrument(t *testing.T) {
	tests := []struct {
		metricType schema.MetricType
		expected   string
	}{
		{schema.MetricTypeGauge, "gauge"},
		{schema.MetricTypeSum, "counter"},
		{schema.MetricTypeHistogram, "histogram"},
		{schema.MetricTypeExponentialHistogram, "histogram"},
		{schema.MetricTypeSummary, "histogram"},
		{schema.MetricTypeEmpty, "gauge"},
	}

	for _, tt := range tests {
		t.Run(string(tt.metricType), func(t *testing.T) {
			result := convertMetricTypeToInstrument(tt.metricType)
			if result != tt.expected {
				t.Errorf("convertMetricTypeToInstrument(%s) = %s, expected %s", tt.metricType, result, tt.expected)
			}
		})
	}
}

func TestGenerateYAML_SumMetricUsesCounter(t *testing.T) {
	telemetry := &schema.Telemetry{
		SchemaKey:     "test.counter",
		Brief:         "Test counter metric",
		MetricType:    schema.MetricTypeSum,
		MetricUnit:    "1",
		TelemetryType: schema.TelemetryTypeMetric,
		Attributes: []schema.Attribute{
			{
				Name:   "test.attr",
				Type:   schema.AttributeTypeStr,
				Source: schema.AttributeSourceDataPoint,
			},
		},
	}

	yaml, err := GenerateYAML(telemetry, nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use "counter" instead of "sum"
	if !strings.Contains(yaml, "instrument: counter") {
		t.Error("YAML should contain 'instrument: counter' for Sum metric type")
	}

	// Should NOT contain "instrument: sum"
	if strings.Contains(yaml, "instrument: sum") {
		t.Error("YAML should not contain 'instrument: sum'")
	}
}

func TestGenerateYAML_AttributeWithoutBrief(t *testing.T) {
	telemetry := &schema.Telemetry{
		SchemaKey:     "test.metric",
		Brief:         "Test metric",
		MetricType:    schema.MetricTypeGauge,
		MetricUnit:    "1",
		TelemetryType: schema.TelemetryTypeMetric,
		Attributes: []schema.Attribute{
			{
				Name:   "test.attribute",
				Type:   schema.AttributeTypeStr,
				Source: schema.AttributeSourceDataPoint,
				// Brief is intentionally not set (empty string)
			},
		},
	}

	yaml, err := GenerateYAML(telemetry, nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should always include brief field, even if empty
	if !strings.Contains(yaml, "brief: ") {
		t.Error("YAML should contain 'brief: ' field for all attributes")
	}

	// Should contain the attribute with empty brief (quoted)
	expectedLines := []string{
		"      - id: test.attribute",
		"        type: string",
		"        requirement_level: recommended",
		`        brief: ""`,
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(yaml, expectedLine) {
			t.Errorf("Expected YAML to contain '%s', but it didn't.\nActual YAML:\n%s", expectedLine, yaml)
		}
	}
}

func TestGenerateYAML_MetricWithoutUnit(t *testing.T) {
	telemetry := &schema.Telemetry{
		SchemaKey:  "test.metric",
		Brief:      "Test metric",
		MetricType: schema.MetricTypeGauge,
		// MetricUnit is intentionally not set (empty string)
		TelemetryType: schema.TelemetryTypeMetric,
		Scope: &schema.Scope{
			Name:      "django",
			Version:   "1.0.0",
			SchemaURL: "https://opentelemetry.io/schemas/1.0.0",
			Attributes: map[string]interface{}{
				"service.name": "django",
			},
		},
	}

	yaml, err := GenerateYAML(telemetry, nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should always include unit field, even if empty
	if !strings.Contains(yaml, "unit: ") {
		t.Error("YAML should contain 'unit: ' field for all metrics")
	}

	// Should contain the metric with empty unit (quoted)
	expectedLines := []string{
		"  - id: metric.django.test.metric",
		"    type: metric",
		"    metric_name: test.metric",
		"    brief: \"Test metric\"",
		"    stability: stable",
		"    instrument: gauge",
		`    unit: ""`,
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(yaml, expectedLine) {
			t.Errorf("Expected YAML to contain '%s', but it didn't.\nActual YAML:\n%s", expectedLine, yaml)
		}
	}
}

func TestGenerateRegistryManifest(t *testing.T) {
	producerName := "otelcontribcol"
	producerVersion := "0.134.0-dev"

	manifest := GenerateRegistryManifest(producerName, producerVersion)

	expectedLines := []string{
		"name: otelcontribcol",
		"description: Schema for otelcontribcol, version 0.134.0-dev",
		"semconv_version: 0.134.0-dev",
		"schema_base_url: http://github.com/nicolastakashi/tallycat/otelcontribcol---0.134.0-dev",
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(manifest, expectedLine) {
			t.Errorf("Expected manifest to contain '%s', but it didn't.\nActual manifest:\n%s", expectedLine, manifest)
		}
	}

	// Verify the manifest is valid YAML format (each line should be key: value)
	lines := strings.Split(strings.TrimSpace(manifest), "\n")
	if len(lines) != 4 {
		t.Errorf("Expected 4 lines in manifest, got %d", len(lines))
	}

	for _, line := range lines {
		if !strings.Contains(line, ": ") {
			t.Errorf("Expected line to contain ': ' separator, got: %s", line)
		}
	}
}

func TestGenerateYAML_LogEvent_BasicTelemetry(t *testing.T) {
	telemetry := &schema.Telemetry{
		SchemaKey:         "user.login",
		Brief:             "User login event",
		TelemetryType:     schema.TelemetryTypeLog,
		LogEventName:      "user.login",
		LogSeverityNumber: 9,
		LogSeverityText:   "INFO",
		LogBody:           "User logged in successfully",
		Attributes: []schema.Attribute{
			{
				Name:             "user.id",
				Type:             schema.AttributeTypeStr,
				Source:           schema.AttributeSourceLogRecord,
				RequirementLevel: schema.RequirementLevelRequired,
				Brief:            "User identifier",
			},
			{
				Name:   "service.name",
				Type:   schema.AttributeTypeStr,
				Source: schema.AttributeSourceResource, // Should be filtered out
			},
		},
		Scope: &schema.Scope{
			Name:      "django",
			Version:   "1.0.0",
			SchemaURL: "https://opentelemetry.io/schemas/1.0.0",
			Attributes: map[string]interface{}{
				"service.name": "django",
			},
		},
	}

	yaml, err := GenerateYAML(telemetry, nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the basic structure for log events
	expectedLines := []string{
		"groups:",
		"  - id: event.django.user.login",
		"    type: event",
		"    name: user.login",
		"    brief: \"User login event\"",
		"    stability: stable",
		"    attributes:",
		"      - id: user.id",
		"        type: string",
		"        requirement_level: required",
		"        stability: stable",
		"        brief: \"User identifier\"",
		"      - id: log.severity.number",
		"        type: int",
		"        requirement_level: recommended",
		"        stability: stable",
		"        brief: \"Log severity number\"",
		"      - id: log.severity.text",
		"        type: string",
		"        requirement_level: recommended",
		"        stability: stable",
		"        brief: \"Log severity text\"",
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(yaml, expectedLine) {
			t.Errorf("Expected YAML to contain '%s', but it didn't.\nActual YAML:\n%s", expectedLine, yaml)
		}
	}

	// Verify that resource attributes are now included
	if !strings.Contains(yaml, "service.name") {
		t.Error("YAML should contain resource attributes")
	}
}

func TestGenerateYAML_LogEvent_NoEventName(t *testing.T) {
	telemetry := &schema.Telemetry{
		SchemaKey:     "error.occurred",
		Brief:         "Error event",
		TelemetryType: schema.TelemetryTypeLog,
		// LogEventName is empty, should use SchemaKey
		LogSeverityNumber: 17,
		LogSeverityText:   "ERROR",
		Attributes: []schema.Attribute{
			{
				Name:   "error.type",
				Type:   schema.AttributeTypeStr,
				Source: schema.AttributeSourceLogRecord,
			},
		},
		Scope: &schema.Scope{
			Name:      "django",
			Version:   "1.0.0",
			SchemaURL: "https://opentelemetry.io/schemas/1.0.0",
			Attributes: map[string]interface{}{
				"service.name": "django",
			},
		},
	}

	yaml, err := GenerateYAML(telemetry, nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use SchemaKey when LogEventName is empty (no 'name' property should be included)
	expectedLines := []string{
		"  - id: event.django.error.occurred",
		"    type: event",
		"    brief:",
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(yaml, expectedLine) {
			t.Errorf("Expected YAML to contain '%s', but it didn't.\nActual YAML:\n%s", expectedLine, yaml)
		}
	}
}

func TestGenerateYAML_LogEvent_NoLogRecordAttributes(t *testing.T) {
	telemetry := &schema.Telemetry{
		SchemaKey:     "simple.event",
		TelemetryType: schema.TelemetryTypeLog,
		LogEventName:  "simple.event",
		Attributes: []schema.Attribute{
			{
				Name:   "service.name",
				Type:   schema.AttributeTypeStr,
				Source: schema.AttributeSourceResource, // Not LogRecord
			},
		},
		Scope: &schema.Scope{
			Name:      "django",
			Version:   "1.0.0",
			SchemaURL: "https://opentelemetry.io/schemas/1.0.0",
			Attributes: map[string]interface{}{
				"service.name": "django",
			},
		},
	}

	yaml, err := GenerateYAML(telemetry, nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should still contain log severity attributes even if no LogRecord attributes
	expectedLines := []string{
		"groups:",
		"  - id: event.django.simple.event",
		"    type: event",
		"    name: simple.event",
		"    attributes:",
		"      - id: log.severity.number",
		"      - id: log.severity.text",
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(yaml, expectedLine) {
			t.Errorf("Expected YAML to contain '%s', but it didn't.\nActual YAML:\n%s", expectedLine, yaml)
		}
	}

	// Should not contain body property since LogBody is empty
	if strings.Contains(yaml, "body:") {
		t.Error("YAML should not contain body property when LogBody is empty")
	}
}

func TestGenerateYAML_LogEvent_WithTelemetrySchema(t *testing.T) {
	telemetry := &schema.Telemetry{
		SchemaKey:     "test.event",
		TelemetryType: schema.TelemetryTypeLog,
		LogEventName:  "test.event",
		Brief:         "Test event",
		Attributes:    []schema.Attribute{}, // Empty in telemetry
	}

	telemetrySchema := &schema.TelemetrySchema{
		Attributes: []schema.Attribute{
			{
				Name:   "custom.attribute",
				Type:   schema.AttributeTypeStr,
				Source: schema.AttributeSourceLogRecord,
				Brief:  "Custom log attribute",
			},
		},
	}

	yaml, err := GenerateYAML(telemetry, telemetrySchema)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use attributes from telemetrySchema when available
	if !strings.Contains(yaml, "custom.attribute") {
		t.Error("YAML should contain attributes from telemetrySchema")
	}
	if !strings.Contains(yaml, "Custom log attribute") {
		t.Error("YAML should contain attribute brief from telemetrySchema")
	}
}

func TestGenerateMultiMetricYAML_MixedTelemetries(t *testing.T) {
	telemetries := []schema.Telemetry{
		{
			SchemaID:      "metric1_schema_id",
			SchemaKey:     "http.server.duration",
			Brief:         "HTTP server request duration",
			MetricType:    schema.MetricTypeHistogram,
			MetricUnit:    "ms",
			TelemetryType: schema.TelemetryTypeMetric,
			Attributes: []schema.Attribute{
				{
					Name:   "http.method",
					Type:   schema.AttributeTypeStr,
					Source: schema.AttributeSourceDataPoint,
				},
			},
			Scope: &schema.Scope{
				Name:      "django",
				Version:   "1.0.0",
				SchemaURL: "https://opentelemetry.io/schemas/1.0.0",
				Attributes: map[string]interface{}{
					"service.name": "django",
				},
			},
		},
		{
			SchemaID:          "log1_schema_id",
			SchemaKey:         "user.login",
			Brief:             "User login event",
			TelemetryType:     schema.TelemetryTypeLog,
			LogEventName:      "user.login",
			LogSeverityNumber: 9,
			LogSeverityText:   "INFO",
			Attributes: []schema.Attribute{
				{
					Name:   "user.id",
					Type:   schema.AttributeTypeStr,
					Source: schema.AttributeSourceLogRecord,
				},
			},
			Scope: &schema.Scope{
				Name:      "django",
				Version:   "1.0.0",
				SchemaURL: "https://opentelemetry.io/schemas/1.0.0",
				Attributes: map[string]interface{}{
					"service.name": "django",
				},
			},
		},
	}

	schemas := map[string]*schema.TelemetrySchema{}

	yaml, err := GenerateMultiMetricYAML(telemetries, schemas)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should contain ONLY the metric, not the log event
	expectedLines := []string{
		"groups:",
		"  - id: metric.django.http.server.duration",
		"    type: metric",
		"    instrument: histogram",
		"      - id: http.method",
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(yaml, expectedLine) {
			t.Errorf("Expected YAML to contain '%s', but it didn't.\nActual YAML:\n%s", expectedLine, yaml)
		}
	}

	// Should NOT contain log-related content
	unexpectedLines := []string{
		"  - id: event.django.user.login",
		"    type: event",
		"    name: user.login",
		"      - id: user.id",
		"      - id: log.severity.number",
	}

	for _, unexpectedLine := range unexpectedLines {
		if strings.Contains(yaml, unexpectedLine) {
			t.Errorf("Expected YAML to NOT contain '%s', but it did.\nActual YAML:\n%s", unexpectedLine, yaml)
		}
	}
}

func TestGenerateMultiLogYAML_EmptyTelemetries(t *testing.T) {
	var telemetries []schema.Telemetry
	schemas := map[string]*schema.TelemetrySchema{}

	yaml, err := GenerateMultiLogYAML(telemetries, schemas)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := ""
	if yaml != expected {
		t.Errorf("Expected '%s', got '%s'", expected, yaml)
	}
}

func TestGenerateMultiLogYAML_SingleLog(t *testing.T) {
	telemetries := []schema.Telemetry{
		{
			SchemaID:          "log1_schema_id",
			SchemaKey:         "user.login",
			Brief:             "User login event",
			TelemetryType:     schema.TelemetryTypeLog,
			LogEventName:      "user.login",
			LogSeverityNumber: 9,
			LogSeverityText:   "INFO",
			LogBody:           "User logged in successfully",
			Attributes: []schema.Attribute{
				{
					Name:   "user.id",
					Type:   schema.AttributeTypeStr,
					Source: schema.AttributeSourceLogRecord,
					Brief:  "User identifier",
				},
			},
			Scope: &schema.Scope{
				Name:      "django",
				Version:   "1.0.0",
				SchemaURL: "https://opentelemetry.io/schemas/1.0.0",
				Attributes: map[string]interface{}{
					"service.name": "django",
				},
			},
		},
	}

	schemas := map[string]*schema.TelemetrySchema{}

	yaml, err := GenerateMultiLogYAML(telemetries, schemas)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedLines := []string{
		"groups:",
		"  - id: event.django.user.login",
		"    type: event",
		"    name: user.login",
		"    brief: \"User login event\"",
		"      - id: user.id",
		"        type: string",
		"        brief: \"User identifier\"",
		"      - id: log.severity.number",
		"        type: int",
		"        brief: \"Log severity number\"",
		"      - id: log.severity.text",
		"        type: string",
		"        brief: \"Log severity text\"",
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(yaml, expectedLine) {
			t.Errorf("Expected YAML to contain '%s', but it didn't.\nActual YAML:\n%s", expectedLine, yaml)
		}
	}
}

func TestGenerateMultiLogYAML_MixedTelemetries(t *testing.T) {
	telemetries := []schema.Telemetry{
		{
			SchemaID:      "metric1_schema_id",
			SchemaKey:     "http.server.duration",
			Brief:         "HTTP server request duration",
			MetricType:    schema.MetricTypeHistogram,
			MetricUnit:    "ms",
			TelemetryType: schema.TelemetryTypeMetric,
			Attributes: []schema.Attribute{
				{
					Name:   "http.method",
					Type:   schema.AttributeTypeStr,
					Source: schema.AttributeSourceDataPoint,
				},
			},
			Scope: &schema.Scope{
				Name:      "django",
				Version:   "1.0.0",
				SchemaURL: "https://opentelemetry.io/schemas/1.0.0",
				Attributes: map[string]interface{}{
					"service.name": "django",
				},
			},
		},
		{
			SchemaID:          "log1_schema_id",
			SchemaKey:         "user.login",
			Brief:             "User login event",
			TelemetryType:     schema.TelemetryTypeLog,
			LogEventName:      "user.login",
			LogSeverityNumber: 9,
			LogSeverityText:   "INFO",
			Attributes: []schema.Attribute{
				{
					Name:   "user.id",
					Type:   schema.AttributeTypeStr,
					Source: schema.AttributeSourceLogRecord,
				},
			},
			Scope: &schema.Scope{
				Name:      "django",
				Version:   "1.0.0",
				SchemaURL: "https://opentelemetry.io/schemas/1.0.0",
				Attributes: map[string]interface{}{
					"service.name": "django",
				},
			},
		},
	}

	schemas := map[string]*schema.TelemetrySchema{}

	yaml, err := GenerateMultiLogYAML(telemetries, schemas)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should contain ONLY the log event, not the metric
	expectedLines := []string{
		"groups:",
		"  - id: event.django.user.login",
		"    type: event",
		"    name: user.login",
		"      - id: user.id",
		"      - id: log.severity.number",
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(yaml, expectedLine) {
			t.Errorf("Expected YAML to contain '%s', but it didn't.\nActual YAML:\n%s", expectedLine, yaml)
		}
	}

	// Should NOT contain metric-related content
	unexpectedLines := []string{
		"  - id: metric.django.http.server.duration",
		"    type: metric",
		"    instrument: histogram",
		"      - id: http.method",
	}

	for _, unexpectedLine := range unexpectedLines {
		if strings.Contains(yaml, unexpectedLine) {
			t.Errorf("Expected YAML to NOT contain '%s', but it did.\nActual YAML:\n%s", unexpectedLine, yaml)
		}
	}
}

func TestGenerateMultiLogYAML_MultipleLogs(t *testing.T) {
	telemetries := []schema.Telemetry{
		{
			SchemaID:          "log1_schema_id",
			SchemaKey:         "user.login",
			Brief:             "User login event",
			TelemetryType:     schema.TelemetryTypeLog,
			LogEventName:      "user.login",
			LogSeverityNumber: 9,
			LogSeverityText:   "INFO",
			Attributes: []schema.Attribute{
				{
					Name:   "user.id",
					Type:   schema.AttributeTypeStr,
					Source: schema.AttributeSourceLogRecord,
				},
			},
			Scope: &schema.Scope{
				Name:      "django",
				Version:   "1.0.0",
				SchemaURL: "https://opentelemetry.io/schemas/1.0.0",
				Attributes: map[string]interface{}{
					"service.name": "django",
				},
			},
		},
		{
			SchemaID:          "log2_schema_id",
			SchemaKey:         "user.logout",
			Brief:             "User logout event",
			TelemetryType:     schema.TelemetryTypeLog,
			LogEventName:      "user.logout",
			LogSeverityNumber: 9,
			LogSeverityText:   "INFO",
			Attributes: []schema.Attribute{
				{
					Name:   "session.id",
					Type:   schema.AttributeTypeStr,
					Source: schema.AttributeSourceLogRecord,
				},
			},
			Scope: &schema.Scope{
				Name:      "django",
				Version:   "1.0.0",
				SchemaURL: "https://opentelemetry.io/schemas/1.0.0",
				Attributes: map[string]interface{}{
					"service.name": "django",
				},
			},
		},
	}

	schemas := map[string]*schema.TelemetrySchema{}

	yaml, err := GenerateMultiLogYAML(telemetries, schemas)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedLines := []string{
		"groups:",
		"  - id: event.django.user.login",
		"    type: event",
		"    name: user.login",
		"      - id: user.id",
		"  - id: event.django.user.logout",
		"    type: event",
		"    name: user.logout",
		"      - id: session.id",
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(yaml, expectedLine) {
			t.Errorf("Expected YAML to contain '%s', but it didn't.\nActual YAML:\n%s", expectedLine, yaml)
		}
	}
}
