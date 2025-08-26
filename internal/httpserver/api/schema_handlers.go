package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tallycat/tallycat/internal/repository"
	"github.com/tallycat/tallycat/internal/schema"
	"github.com/tallycat/tallycat/internal/weaver"
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

		resp := ListResponse[schema.Telemetry]{
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

		resp := ListResponse[schema.TelemetrySchema]{
			Items:    assignments,
			Total:    total,
			Page:     params.Page,
			PageSize: params.PageSize,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func HandleTelemetrySchemaVersionAssignment(
	schemaRepo repository.TelemetrySchemaRepository,
	historyRepo repository.TelemetryHistoryRepository,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		schemaKey := chi.URLParam(r, "key")

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

		// Record history entry after successful version assignment
		history := &schema.TelemetryHistory{
			SchemaKey: schemaKey,
			Version:   assignment.Version,
			Timestamp: time.Now(),
			Author:    nil,
			Summary:   fmt.Sprintf("Assigned schema version %s to schema %s", assignment.Version, assignment.SchemaId),
			Status:    "",
			Snapshot:  nil,
		}

		if err := historyRepo.InsertTelemetryHistory(ctx, history); err != nil {
			slog.Error("failed to record telemetry history", "error", err)
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

func HandleWeaverSchemaExport(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		schemaId := chi.URLParam(r, "schemaId")

		schema, err := schemaRepo.GetTelemetrySchema(ctx, schemaId)
		if err != nil {
			http.Error(w, "failed to get schema", http.StatusInternalServerError)
			return
		}

		if schema == nil {
			http.Error(w, "schema not found", http.StatusNotFound)
			return
		}

		telemetry, err := schemaRepo.GetTelemetry(ctx, schema.SchemaId)
		if err != nil {
			http.Error(w, "failed to get telemetry", http.StatusInternalServerError)
			return
		}

		yaml, err := weaver.GenerateYAML(telemetry, schema)
		if err != nil {
			http.Error(w, "failed to generate YAML", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename="+schema.SchemaId+".yaml")
		w.Write([]byte(yaml))
	}
}

// HandleTelemetryHistory returns paginated/sorted telemetry history entries for a given telemetry_id
func HandleTelemetryHistory(historyRepo repository.TelemetryHistoryRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		telemetryID := chi.URLParam(r, "key")
		params := ParseListQueryParams(r)

		histories, total, err := historyRepo.ListTelemetryHistory(ctx, telemetryID, params.Page, params.PageSize)
		if err != nil {
			http.Error(w, "failed to list telemetry history", http.StatusInternalServerError)
			return
		}

		resp := ListResponse[schema.TelemetryHistory]{
			Items:    histories,
			Total:    total,
			Page:     params.Page,
			PageSize: params.PageSize,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
