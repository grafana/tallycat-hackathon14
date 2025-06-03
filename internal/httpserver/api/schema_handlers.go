package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tallycat/tallycat/internal/repository"
	"github.com/tallycat/tallycat/internal/schema"
)

// HandleTelemetryList returns a paginated, filtered, and searched list of schemas as JSON.
func HandleTelemetryList(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := ParseListQueryParams(r)
		telemetries, total, err := schemaRepo.ListTelemetries(ctx, params)
		if err != nil {
			slog.Error("failed to list telemetry", "error", err)
			http.Error(w, "failed to list telemetry", http.StatusInternalServerError)
			return
		}

		resp := ListSchemasResponse{
			Items:    telemetries,
			Total:    total,
			Page:     params.Page,
			PageSize: params.PageSize,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func HandleGetTelemetry(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		signalKey := chi.URLParam(r, "key")

		schema, err := schemaRepo.GetTelemetry(ctx, signalKey)
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

func HandleTelemetrySchemas(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		schemaKey := chi.URLParam(r, "key")
		params := ParseListQueryParams(r)

		assignments, total, err := schemaRepo.ListTelemetrySchemas(ctx, schemaKey, params)
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

func HandleTelemetrySchemaVersionAssignment(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		assignment := schema.SchemaAssignment{}
		err := json.NewDecoder(r.Body).Decode(&assignment)
		if err != nil {
			http.Error(w, "failed to decode request body", http.StatusBadRequest)
			return
		}

		err = schemaRepo.AssignTelemetrySchemaVersion(ctx, assignment)
		if err != nil {
			http.Error(w, "failed to assign schema version", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(assignment)
	}
}

func HandleGetTelemetrySchema(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		schemaId := chi.URLParam(r, "schemaId")

		schema, err := schemaRepo.GetTelemetrySchema(ctx, schemaId)
		if err != nil {
			http.Error(w, "failed to get schema", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(schema)
	}
}
