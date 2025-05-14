package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
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
				Name:               sch.SignalKey,
				Description:        "Test", // Not stored yet
				Type:               sch.SignalType,
				DataType:           ptr.Deref(sch.MetricType, ""),
				Status:             "Active", // Default for now
				Format:             "OTLP",   // Default for now
				LastUpdated:        sch.UpdatedAt,
				SchemaVersionCount: sch.SchemaVersionCount,
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

// HandleGetSchema returns a specific schema by its key
func HandleGetSchema(schemaRepo repository.SchemaProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		signalKey := chi.URLParam(r, "key")

		schema, err := schemaRepo.GetSchemaByKey(ctx, signalKey)
		if err != nil {
			http.Error(w, "failed to get schema", http.StatusInternalServerError)
			return
		}

		if schema == nil {
			http.Error(w, "schema not found", http.StatusNotFound)
			return
		}

		// Convert repository schema to API response
		resp := GetSchemaResponse{
			ID:                     schema.SchemaID,
			Name:                   schema.SignalKey,
			Type:                   schema.SignalType,
			DataType:               ptr.Deref(schema.MetricType, ""),
			Status:                 "active", // Default for now
			Description:            "",       // Not stored yet
			LastUpdated:            schema.UpdatedAt,
			SchemaVersionCount:     schema.SchemaVersionCount,
			Created:                schema.CreatedAt,
			Fields:                 len(schema.FieldNames),
			Source:                 "OpenTelemetry Collector", // Default for now
			InstrumentationLibrary: schema.ScopeName,
			Format:                 "OTLP", // Default for now
			Unit:                   ptr.Deref(schema.Unit, ""),
			Aggregation:            "cumulative",     // Default for now
			Cardinality:            "high",           // Default for now
			Tags:                   []string{},       // Not stored yet
			Sources:                []SchemaSource{}, // Not stored yet
			SourceTeams:            []string{},       // Not stored yet
			Schema:                 make([]SchemaField, 0, len(schema.FieldNames)),
			MetricDetails: MetricDetails{
				Type:                 ptr.Deref(schema.MetricType, ""),
				Unit:                 ptr.Deref(schema.Unit, ""),
				Aggregation:          "cumulative", // Default for now
				MetricName:           schema.SignalKey,
				OtelCompatible:       true,    // Default for now
				Buckets:              []int{}, // Not stored yet
				Monotonic:            false,   // Not stored yet
				InstrumentationScope: schema.ScopeName,
				SemanticConventions:  "http", // Default for now
			},
			UsedBy:          []SchemaUsage{},    // Not stored yet
			History:         []SchemaVersion{},  // Not stored yet
			Examples:        []SchemaExample{},  // Not stored yet
			ValidationRules: []ValidationRule{}, // Not stored yet
		}

		// Convert fields
		for _, name := range schema.FieldNames {
			field := SchemaField{
				Name: name,
				Type: schema.FieldTypes[name],
			}
			resp.Schema = append(resp.Schema, field)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
