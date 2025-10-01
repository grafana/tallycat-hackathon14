package schema

import "time"

type SchemaAssignment struct {
	SchemaId string `json:"schemaId"`
	Version  string `json:"version,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type TelemetrySchema struct {
	SchemaId    string             `json:"schemaId"`
	Version     string             `json:"version"`
	EntityCount int                `json:"entityCount"`
	LastSeen    *time.Time         `json:"lastSeen,omitempty"`
	Entities    map[string]*Entity `json:"entities"`
	Scope       *Scope             `json:"scope"`
	Attributes  []Attribute        `json:"attributes"`
}
