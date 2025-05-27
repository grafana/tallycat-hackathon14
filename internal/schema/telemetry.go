package schema

import (
	"time"
)

type MetricType string

const (
	MetricTypeEmpty                MetricType = "Empty"
	MetricTypeGauge                MetricType = "Gauge"
	MetricTypeSum                  MetricType = "Sum"
	MetricTypeHistogram            MetricType = "Histogram"
	MetricTypeExponentialHistogram MetricType = "ExponentialHistogram"
	MetricTypeSummary              MetricType = "Summary"
)

type TelemetryProtocol string

const (
	TelemetryProtocolOTLP       TelemetryProtocol = "OTLP"
	TelemetryProtocolPrometheus TelemetryProtocol = "Prometheus"
)

type MetricTemporality string

const (
	MetricTemporalityCumulative  MetricTemporality = "Cumulative"
	MetricTemporalityDelta       MetricTemporality = "Delta"
	MetricTemporalityUnspecified MetricTemporality = "Unspecified"
)

type TelemetryType string

const (
	TelemetryTypeMetric TelemetryType = "Metric"
)

type Telemetry struct {
	SchemaID          string            `json:"schemaId"`
	SchemaVersion     string            `json:"schemaVersion"`
	SchemaURL         string            `json:"schemaURL,omitempty"`
	SchemaKey         string            `json:"schemaKey"`
	TelemetryType     TelemetryType     `json:"telemetryType"`
	MetricUnit        string            `json:"metricUnit"`
	MetricType        MetricType        `json:"metricType"`
	MetricTemporality MetricTemporality `json:"metricTemporality"`
	Attributes        []Attribute       `json:"attributes"`
	Brief             string            `json:"brief,omitempty"`
	Note              string            `json:"note,omitempty"`
	Protocol          TelemetryProtocol `json:"protocol"`
	SeenCount         int               `json:"seenCount"`
	CreatedAt         time.Time         `json:"createdAt"`
	UpdatedAt         time.Time         `json:"updatedAt"`
	// Producers maps producer keys to their information
	Producers map[string]*Producer `json:"producers"`
}
