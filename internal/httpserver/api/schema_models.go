package api

import (
	"github.com/tallycat/tallycat/internal/schema"
)

type ListSchemasResponse struct {
	Items    []schema.Telemetry `json:"items"`
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
}

type ListResponse[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}
