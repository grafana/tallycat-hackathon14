package api

import (
	"encoding/json"
	"net/http"

	"github.com/tallycat/tallycat/internal/repository"
	"k8s.io/utils/ptr"
)

// HandleListSchemas returns a paginated, filtered, and searched list of schemas as JSON.
func HandleListSchemas(schemaRepo repository.SchemaProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := ParseListQueryParams(r)
		schemas, total, err := schemaRepo.ListSchemas(ctx, params)
		if err != nil {
			http.Error(w, "failed to list schemas", http.StatusInternalServerError)
			return
		}

		items := make([]SchemaListItem, 0, len(schemas))
		for _, sch := range schemas {
			item := SchemaListItem{
				Name:        sch.SignalKey,
				Description: "Test", // Not stored yet
				Type:        sch.SignalType,
				DataType:    ptr.Deref(sch.MetricType, ""),
				Status:      "Active", // Default for now
				Format:      "OTLP",   // Default for now
				LastUpdated: sch.UpdatedAt,
			}
			if item.Name == "" {
				item.Name = sch.SchemaID
			}
			items = append(items, item)
		}

		resp := ListSchemasResponse{
			Schemas:  items,
			Total:    total,
			Page:     params.Page,
			PageSize: params.PageSize,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
