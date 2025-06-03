package schema

import "time"

type SchemaAssignment struct {
	SchemaId string `json:"schemaId"`
	Version  string `json:"version,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type TelemetrySchema struct {
	SchemaId      string      `json:"schemaId"`
	Version       string      `json:"version"`
	ProducerCount int         `json:"producerCount"`
	LastSeen      *time.Time  `json:"lastSeen,omitempty"`
	Producers     []Producer  `json:"producers"`
	Attributes    []Attribute `json:"attributes"`
}
