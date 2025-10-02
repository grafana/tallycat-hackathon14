package weaver

import (
	"fmt"
	"log"

	"github.com/tallycat/tallycat/internal/schema"
)

// ExampleGenerator demonstrates the Weaver YAML generation
func ExampleGenerateYAML() {

	// Create a sample telemetry schema
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
				Name:             "http.status_code",
				Type:             schema.AttributeTypeInt,
				Source:           schema.AttributeSourceDataPoint,
				RequirementLevel: schema.RequirementLevelRecommended,
				Brief:            "HTTP response status code",
			},
			{
				Name:   "service.name",
				Type:   schema.AttributeTypeStr,
				Source: schema.AttributeSourceResource,
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
		log.Fatal(err)
	}

	fmt.Println(yaml)

	// Output:
	// groups:
	//   - id: metric.django.http.server.duration
	//     type: metric
	//     metric_name: http.server.duration
	//     brief: "Measures the duration of HTTP server requests"
	//     stability: stable
	//     instrument: histogram
	//     unit: "ms"
	//     attributes:
	//       - id: http.method
	//         type: string
	//         requirement_level: required
	//         stability: stable
	//         brief: "HTTP request method"
	//       - id: http.status_code
	//         type: int
	//         requirement_level: recommended
	//         stability: stable
	//         brief: "HTTP response status code"
	//       - id: service.name
	//         type: string
	//         requirement_level: required
	//         stability: stable
	//         brief: ""
}
