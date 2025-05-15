package api

import "time"

// SchemaListItem is the API response item for a schema in the catalog
// (Fields: name, description, type, data_type, status, format, last_updated)
type SchemaListItem struct {
	ID                     string    `json:"id"`
	Name                   string    `json:"name"`
	Type                   string    `json:"type"`
	DataType               string    `json:"dataType"`
	Status                 string    `json:"status"`
	Description            string    `json:"description"`
	LastUpdated            time.Time `json:"lastUpdated"`
	SchemaVersionCount     int       `json:"schemaVersionCount"`
	Created                time.Time `json:"created"`
	Fields                 int       `json:"fields"`
	Source                 string    `json:"source"`
	InstrumentationLibrary string    `json:"instrumentationLibrary"`
	Format                 string    `json:"format"`
	Unit                   string    `json:"unit"`
	Aggregation            string    `json:"aggregation"`
	Cardinality            string    `json:"cardinality"`
	Tags                   []string  `json:"tags"`
	SourceTeams            []string  `json:"sourceTeams"`
}

type ListSchemasResponse struct {
	Items    []SchemaListItem `json:"items"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

type GetSchemaResponse struct {
	ID                     string           `json:"id"`
	Name                   string           `json:"name"`
	Type                   string           `json:"type"`
	DataType               string           `json:"dataType"`
	Status                 string           `json:"status"`
	Description            string           `json:"description"`
	LastUpdated            time.Time        `json:"lastUpdated"`
	SchemaVersionCount     int              `json:"schemaVersionCount"`
	Created                time.Time        `json:"created"`
	Fields                 int              `json:"fields"`
	Source                 string           `json:"source"`
	InstrumentationLibrary string           `json:"instrumentationLibrary"`
	Format                 string           `json:"format"`
	Unit                   string           `json:"unit"`
	Aggregation            string           `json:"aggregation"`
	Cardinality            string           `json:"cardinality"`
	Tags                   []string         `json:"tags"`
	Sources                []SchemaSource   `json:"sources"`
	SourceTeams            []string         `json:"sourceTeams"`
	Schema                 []SchemaField    `json:"schema"`
	MetricDetails          MetricDetails    `json:"metricDetails"`
	UsedBy                 []SchemaUsage    `json:"usedBy"`
	History                []SchemaVersion  `json:"history"`
	Examples               []SchemaExample  `json:"examples"`
	ValidationRules        []ValidationRule `json:"validationRules"`
}

type SchemaSource struct {
	ID                    string    `json:"id"`
	Name                  string    `json:"name"`
	Team                  string    `json:"team"`
	Environment           string    `json:"environment"`
	Health                string    `json:"health"`
	Version               string    `json:"version"`
	Volume                int       `json:"volume"`
	DailyAverage          int       `json:"dailyAverage"`
	Peak                  int       `json:"peak"`
	Contribution          int       `json:"contribution"`
	Compliance            string    `json:"compliance"`
	RequiredFieldsPresent int       `json:"requiredFieldsPresent"`
	RequiredFieldsTotal   int       `json:"requiredFieldsTotal"`
	OptionalFieldsPresent int       `json:"optionalFieldsPresent"`
	OptionalFieldsTotal   int       `json:"optionalFieldsTotal"`
	LastValidated         time.Time `json:"lastValidated"`
}

type SchemaField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

type MetricDetails struct {
	Type                 string `json:"type"`
	Unit                 string `json:"unit"`
	Aggregation          string `json:"aggregation"`
	MetricName           string `json:"metricName"`
	OtelCompatible       bool   `json:"otelCompatible"`
	Buckets              []int  `json:"buckets"`
	Monotonic            bool   `json:"monotonic"`
	InstrumentationScope string `json:"instrumentationScope"`
	SemanticConventions  string `json:"semanticConventions"`
}

type SchemaUsage struct {
	Name string `json:"name"`
	Type string `json:"type"`
	ID   string `json:"id"`
}

type SchemaVersion struct {
	Version          string    `json:"version"`
	Date             time.Time `json:"date"`
	Author           string    `json:"author"`
	Changes          string    `json:"changes"`
	ValidationStatus string    `json:"validationStatus"`
}

type SchemaExample struct {
	Description string                 `json:"description"`
	Value       map[string]interface{} `json:"value"`
}

type ValidationRule struct {
	Field       string `json:"field"`
	Rule        string `json:"rule"`
	Description string `json:"description"`
}
