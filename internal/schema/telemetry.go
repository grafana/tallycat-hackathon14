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
	TelemetryTypeLog    TelemetryType = "Log"
)

type Telemetry struct {
	SchemaID      string        `json:"schemaId"`
	SchemaVersion string        `json:"schemaVersion"`
	SchemaURL     string        `json:"schemaURL,omitempty"`
	SchemaKey     string        `json:"schemaKey"`
	TelemetryType TelemetryType `json:"telemetryType"`
	// Metric fields
	MetricUnit        string            `json:"metricUnit"`
	MetricType        MetricType        `json:"metricType"`
	MetricTemporality MetricTemporality `json:"metricTemporality"`
	Brief             string            `json:"brief,omitempty"`
	//Log fields
	LogSeverityNumber         int    `json:"logSeverityNumber"`
	LogSeverityText           string `json:"logSeverityText"`
	LogBody                   string `json:"logBody"`
	LogFlags                  int    `json:"logFlags"`
	LogTraceID                string `json:"logTraceID"`
	LogSpanID                 string `json:"logSpanID"`
	LogEventName              string `json:"logEventName"`
	LogDroppedAttributesCount int    `json:"logDroppedAttributesCount"`

	Attributes []Attribute       `json:"attributes"`
	Note       string            `json:"note,omitempty"`
	Protocol   TelemetryProtocol `json:"protocol"`
	SeenCount  int               `json:"seenCount"`
	CreatedAt  time.Time         `json:"createdAt"`
	UpdatedAt  time.Time         `json:"updatedAt"`
	// Producers maps producer keys to their information
	Producers map[string]*Producer `json:"producers"`
}

type TelemetryHistory struct {
	Id        int       `json:"id"`
	SchemaKey string    `json:"schemaKey"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Author    *string   `json:"author,omitempty"`
	Summary   string    `json:"summary"`
	Status    string    `json:"status"`
	Snapshot  []byte    `json:"snapshot"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
