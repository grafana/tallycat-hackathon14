package schema

import "time"

type SchemaAssignment struct {
	SchemaId string `json:"schemaId"`
	Version  string `json:"version,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Status   string `json:"status,omitempty"`
}

type SchemaAssignmentRow struct {
	SchemaId      string     `json:"schemaId"`
	Status        string     `json:"status"`
	Version       string     `json:"version"`
	ProducerCount int        `json:"producerCount"`
	LastSeen      *time.Time `json:"lastSeen,omitempty"`
}
