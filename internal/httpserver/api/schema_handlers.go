package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tallycat/tallycat/internal/repository"
	"github.com/tallycat/tallycat/internal/schema"
)

// HandleListSchemas returns a paginated, filtered, and searched list of schemas as JSON.
func HandleListSchemas(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := ParseListQueryParams(r)
		schemas, total, err := schemaRepo.ListSchemas(ctx, params)
		if err != nil {
			slog.Error("failed to list schemas", "error", err)
			http.Error(w, "failed to list schemas", http.StatusInternalServerError)
			return
		}

		resp := ListSchemasResponse{
			Items:    schemas,
			Total:    total,
			Page:     params.Page,
			PageSize: params.PageSize,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func HandleGetSchema(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(schema)
	}
}

func HandleAssignSchemaVersion(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		assignment := schema.SchemaAssignment{}
		err := json.NewDecoder(r.Body).Decode(&assignment)
		if err != nil {
			http.Error(w, "failed to decode request body", http.StatusBadRequest)
			return
		}

		err = schemaRepo.AssignSchemaVersion(ctx, assignment)
		if err != nil {
			http.Error(w, "failed to assign schema version", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(assignment)
	}
}

// HandleListSchemaAssignmentsForKey returns a paged, filtered list of schema assignments for a given schemaKey.
func HandleListSchemaAssignmentsForKey(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		schemaKey := chi.URLParam(r, "key")
		params := ParseListQueryParams(r)

		assignments, total, err := schemaRepo.ListSchemaAssignmentsForKey(ctx, schemaKey, params)
		if err != nil {
			slog.Error("failed to list schema assignments", "error", err)
			http.Error(w, "failed to list schema assignments", http.StatusInternalServerError)
			return
		}

		resp := struct {
			Items    any `json:"items"`
			Total    int `json:"total"`
			Page     int `json:"page"`
			PageSize int `json:"pageSize"`
		}{
			Items:    assignments,
			Total:    total,
			Page:     params.Page,
			PageSize: params.PageSize,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
