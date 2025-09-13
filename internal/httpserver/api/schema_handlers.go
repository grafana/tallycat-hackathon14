package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
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
		metricKey := chi.URLParam(r, "key")

		schema, err := schemaRepo.GetTelemetrySchema(ctx, schemaId)
		if err != nil {
			http.Error(w, "failed to get schema", http.StatusInternalServerError)
			return
		}

		if schema == nil {
			http.Error(w, "schema not found", http.StatusNotFound)
			return
		}

		telemetry, err := schemaRepo.GetTelemetry(ctx, metricKey)
		if err != nil {
			http.Error(w, "failed to get telemetry", http.StatusInternalServerError)
			return
		}

		yaml, err := weaver.GenerateYAML(telemetry, schema)
		if err != nil {
			http.Error(w, "failed to generate YAML", http.StatusInternalServerError)
			return
		}

		// Create a ZIP file containing the YAML content
		var buf bytes.Buffer
		zipWriter := zip.NewWriter(&buf)

		// Create a file inside the ZIP with the YAML content
		yamlFileName := schema.SchemaId + ".yaml"
		yamlFile, err := zipWriter.Create(yamlFileName)
		if err != nil {
			http.Error(w, "failed to create zip file", http.StatusInternalServerError)
			return
		}

		// Write the YAML content to the file inside the ZIP
		_, err = yamlFile.Write([]byte(yaml))
		if err != nil {
			http.Error(w, "failed to write yaml to zip", http.StatusInternalServerError)
			return
		}

		// Close the ZIP writer to finalize the archive
		err = zipWriter.Close()
		if err != nil {
			http.Error(w, "failed to close zip file", http.StatusInternalServerError)
			return
		}

		// Set the appropriate headers for ZIP file download
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename="+schema.SchemaId+".zip")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))

		// Write the ZIP file content to the response
		w.Write(buf.Bytes())
	}
}

func HandleProducerWeaverSchemaExport(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		producerNameVersion := chi.URLParam(r, "producerNameVersion")

		// Parse producer name and version from the URL parameter
		producerName, producerVersion, err := parseProducerNameVersion(producerNameVersion)
		if err != nil {
			http.Error(w, "invalid producer format, expected name-version", http.StatusBadRequest)
			return
		}

		// Get all telemetries for this producer
		telemetries, err := schemaRepo.ListTelemetriesByProducer(ctx, producerName, producerVersion)
		if err != nil {
			slog.Error("failed to get telemetries for producer", "producer", producerNameVersion, "error", err)
			http.Error(w, "failed to get telemetries for producer", http.StatusInternalServerError)
			return
		}

		// Check if producer exists but has no metrics
		if len(telemetries) == 0 {
			// According to our specification: return 204 for producers with no metrics
			// We can't distinguish between "producer doesn't exist" and "producer has no metrics"
			// from the current repository implementation, so we treat empty results as "no metrics"
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Generate multi-metric YAML
		yaml, err := weaver.GenerateMultiMetricYAML(telemetries, nil)
		if err != nil {
			slog.Error("failed to generate multi-metric YAML", "producer", producerNameVersion, "error", err)
			http.Error(w, "failed to generate YAML", http.StatusInternalServerError)
			return
		}

		// Create a ZIP file containing the YAML content
		var buf bytes.Buffer
		zipWriter := zip.NewWriter(&buf)

		// Create a file inside the ZIP with the YAML content
		yamlFileName := producerNameVersion + ".yaml"
		yamlFile, err := zipWriter.Create(yamlFileName)
		if err != nil {
			http.Error(w, "failed to create zip file", http.StatusInternalServerError)
			return
		}

		// Write the YAML content to the file inside the ZIP
		_, err = yamlFile.Write([]byte(yaml))
		if err != nil {
			http.Error(w, "failed to write yaml to zip", http.StatusInternalServerError)
			return
		}

		// Close the ZIP writer to finalize the archive
		err = zipWriter.Close()
		if err != nil {
			http.Error(w, "failed to close zip file", http.StatusInternalServerError)
			return
		}

		// Set the appropriate headers for ZIP file download
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename="+producerNameVersion+".zip")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))

		// Write the ZIP file content to the response
		w.Write(buf.Bytes())
	}
}

// parseProducerNameVersion parses the producer name@version format
func parseProducerNameVersion(nameVersion string) (string, string, error) {
	// Find the last @ to separate name and version
	lastAt := strings.LastIndex(nameVersion, "@")
	if lastAt == -1 || lastAt == 0 || lastAt == len(nameVersion)-1 {
		return "", "", fmt.Errorf("invalid producer format, expected name@version")
	}

	name := nameVersion[:lastAt]
	version := nameVersion[lastAt+1:]

	return name, version, nil
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
