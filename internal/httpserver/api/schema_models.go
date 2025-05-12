package api

import "time"

// SchemaListItem is the API response item for a schema in the catalog
// (Fields: name, description, type, data_type, status, format, last_updated)
type SchemaListItem struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	DataType    string    `json:"data_type"`
	Status      string    `json:"status"`
	Format      string    `json:"format"`
	LastUpdated time.Time `json:"last_updated"`
}

type ListSchemasResponse struct {
	Schemas  []SchemaListItem `json:"schemas"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}
