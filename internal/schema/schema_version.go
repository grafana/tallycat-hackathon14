package schema

import "time"

type SchemaAssignment struct {
	SchemaId string `json:"schemaId"`
	Version  string `json:"version,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type SchemaAssignmentRow struct {
	SchemaId      string     `json:"schemaId"`
	Version       string     `json:"version"`
	ProducerCount int        `json:"producerCount"`
	LastSeen      *time.Time `json:"lastSeen,omitempty"`
}
