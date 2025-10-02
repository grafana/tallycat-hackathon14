package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

func HandleEntityWeaverSchemaExport(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		entityType := chi.URLParam(r, "entityType")

		// Get all telemetries for this entity type
		telemetries, err := schemaRepo.ListTelemetriesByEntity(ctx, entityType)
		if err != nil {
			slog.Error("failed to get telemetries for entity", "entityType", entityType, "error", err)
			http.Error(w, "failed to get telemetries for entity", http.StatusInternalServerError)
			return
		}

		// Check if producer exists but has no telemetries
		if len(telemetries) == 0 {
			// According to our specification: return 204 for producers with no telemetries
			// We can't distinguish between "producer doesn't exist" and "producer has no telemetries"
			// from the current repository implementation, so we treat empty results as "no telemetries"
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Generate separate YAML files for metrics, logs, and spans
		metricsYAML, err := weaver.GenerateMultiMetricYAML(telemetries, nil)
		if err != nil {
			slog.Error("failed to generate metrics YAML", "entityType", entityType, "error", err)
			http.Error(w, "failed to generate metrics YAML", http.StatusInternalServerError)
			return
		}

		logsYAML, err := weaver.GenerateMultiLogYAML(telemetries, nil)
		if err != nil {
			slog.Error("failed to generate logs YAML", "entityType", entityType, "error", err)
			http.Error(w, "failed to generate logs YAML", http.StatusInternalServerError)
			return
		}

		spansYAML, err := weaver.GenerateMultiSpanYAML(telemetries, nil)
		if err != nil {
			slog.Error("failed to generate spans YAML", "entityType", entityType, "error", err)
			http.Error(w, "failed to generate spans YAML", http.StatusInternalServerError)
			return
		}

		// Check if we have any content to include in the ZIP
		hasMetrics := metricsYAML != ""
		hasLogs := logsYAML != ""
		hasSpans := spansYAML != ""

		if !hasMetrics && !hasLogs && !hasSpans {
			// No telemetries of any type found
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Generate registry manifest
		manifest := weaver.GenerateRegistryManifest(entityType, "")

		// Create a ZIP file containing the YAML schemas and registry manifest
		var buf bytes.Buffer
		zipWriter := zip.NewWriter(&buf)

		// Create the metrics YAML file inside the ZIP (only if we have metrics)
		if hasMetrics {
			metricsFileName := entityType + "-metrics.yaml"
			metricsFile, err := zipWriter.Create(metricsFileName)
			if err != nil {
				http.Error(w, "failed to create metrics file in zip", http.StatusInternalServerError)
				return
			}

			// Write the metrics YAML content to the file inside the ZIP
			_, err = metricsFile.Write([]byte(metricsYAML))
			if err != nil {
				http.Error(w, "failed to write metrics yaml to zip", http.StatusInternalServerError)
				return
			}
		}

		// Create the logs/events YAML file inside the ZIP (only if we have logs)
		if hasLogs {
			eventsFileName := entityType + "-events.yaml"
			eventsFile, err := zipWriter.Create(eventsFileName)
			if err != nil {
				http.Error(w, "failed to create events file in zip", http.StatusInternalServerError)
				return
			}

			// Write the logs YAML content to the file inside the ZIP
			_, err = eventsFile.Write([]byte(logsYAML))
			if err != nil {
				http.Error(w, "failed to write events yaml to zip", http.StatusInternalServerError)
				return
			}
		}

		// Create the spans YAML file inside the ZIP (only if we have spans)
		if hasSpans {
			spansFileName := entityType + "-spans.yaml"
			spansFile, err := zipWriter.Create(spansFileName)
			if err != nil {
				http.Error(w, "failed to create spans file in zip", http.StatusInternalServerError)
				return
			}

			// Write the spans YAML content to the file inside the ZIP
			_, err = spansFile.Write([]byte(spansYAML))
			if err != nil {
				http.Error(w, "failed to write spans yaml to zip", http.StatusInternalServerError)
				return
			}
		}

		// Create the registry manifest file inside the ZIP
		manifestFile, err := zipWriter.Create("registry_manifest.yaml")
		if err != nil {
			http.Error(w, "failed to create manifest file in zip", http.StatusInternalServerError)
			return
		}

		// Write the manifest content to the manifest file inside the ZIP
		_, err = manifestFile.Write([]byte(manifest))
		if err != nil {
			http.Error(w, "failed to write manifest to zip", http.StatusInternalServerError)
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
		w.Header().Set("Content-Disposition", "attachment; filename="+entityType+".zip")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))

		// Write the ZIP file content to the response
		w.Write(buf.Bytes())
	}
}

func HandleEntitySchemaExport(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		entityType := chi.URLParam(r, "entityType")
		telemetryType := chi.URLParam(r, "type")

		// Get all telemetries for this entity type
		telemetries, err := schemaRepo.ListTelemetriesByEntity(ctx, entityType)
		if err != nil {
			slog.Error("failed to get telemetries for entity", "entityType", entityType, "error", err)
			http.Error(w, "failed to get telemetries for entity", http.StatusInternalServerError)
			return
		}

		// Check if producer exists but has no telemetries
		if len(telemetries) == 0 {
			// According to our specification: return 204 for producers with no telemetries
			// We can't distinguish between "producer doesn't exist" and "producer has no telemetries"
			// from the current repository implementation, so we treat empty results as "no telemetries"
			w.WriteHeader(http.StatusNoContent)
			return
		}

		outBuff := &bytes.Buffer{}
		var genYaml func(telemetries []schema.Telemetry, schemas map[string]*schema.TelemetrySchema) (string, error)
		switch telemetryType {
		case "metrics":
			genYaml = weaver.GenerateMultiMetricYAML
		case "logs":
			genYaml = weaver.GenerateMultiLogYAML
		case "spans":
			genYaml = weaver.GenerateMultiSpanYAML
		}
		metricsYAML, err := genYaml(telemetries, nil)
		if err != nil {
			slog.Error("failed to generate schema YAML", "entityType", entityType, "error", err)
			http.Error(w, "failed to generate metrics YAML", http.StatusInternalServerError)
			return
		}
		outBuff.WriteString(metricsYAML)
		io.Copy(w, outBuff)
	}
}

func HandleEntityDashboardExport(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		entityType := chi.URLParam(r, "entityType")

		// Get all telemetries for this entity type
		telemetries, err := schemaRepo.ListTelemetriesByEntity(ctx, entityType)
		if err != nil {
			slog.Error("failed to get telemetries for entity", "entityType", entityType, "error", err)
			http.Error(w, "failed to get telemetries for entity", http.StatusInternalServerError)
			return
		}

		// Check if producer exists but has no telemetries
		if len(telemetries) == 0 {
			// According to our specification: return 204 for producers with no telemetries
			// We can't distinguish between "producer doesn't exist" and "producer has no telemetries"
			// from the current repository implementation, so we treat empty results as "no telemetries"
			w.WriteHeader(http.StatusNoContent)
			return
		}

		metricsSchema, err := weaver.GenerateMultiMetricYAML(telemetries, nil)
		if err != nil {
			slog.Error("failed to generate schema YAML", "entityType", entityType, "error", err)
			http.Error(w, "failed to generate metrics YAML", http.StatusInternalServerError)
			return
		}
		// logSchema, err := weaver.GenerateMultiLogYAML(telemetries, nil)
		// if err != nil {
		// 	slog.Error("failed to generate schema YAML", "entityType", entityType, "error", err)
		// 	http.Error(w, "failed to generate log YAML", http.StatusInternalServerError)
		// 	return
		// }
		// spanSchema, err := weaver.GenerateMultiSpanYAML(telemetries, nil)
		// if err != nil {
		// 	slog.Error("failed to generate schema YAML", "entityType", entityType, "error", err)
		// 	http.Error(w, "failed to generate span YAML", http.StatusInternalServerError)
		// 	return
		// }

		dashboards, err := weaver.RenderDashboards(ctx, entityType, []string{metricsSchema})
		if err != nil {
			slog.Error("failed to generate dashboards", "entityType", entityType, "error", err)
			http.Error(w, "failed to generate dashboards", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(dashboards)
	}
}

func HandleScopeWeaverSchemaExport(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		scopeName := chi.URLParam(r, "scope")

		// Get all telemetries for this scope
		telemetries, err := schemaRepo.ListTelemetriesByScope(ctx, scopeName)
		if err != nil {
			slog.Error("failed to get telemetries for scope", "scopeName", scopeName, "error", err)
			http.Error(w, "failed to get telemetries for scope", http.StatusInternalServerError)
			return
		}

		// Check if scope exists but has no telemetries
		if len(telemetries) == 0 {
			// According to our specification: return 204 for scopes with no telemetries
			// We can't distinguish between "scope doesn't exist" and "scope has no telemetries"
			// from the current repository implementation, so we treat empty results as "no telemetries"
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Generate separate YAML files for metrics, logs, and spans
		metricsYAML, err := weaver.GenerateMultiMetricYAML(telemetries, nil)
		if err != nil {
			slog.Error("failed to generate metrics YAML", "scopeName", scopeName, "error", err)
			http.Error(w, "failed to generate metrics YAML", http.StatusInternalServerError)
			return
		}

		logsYAML, err := weaver.GenerateMultiLogYAML(telemetries, nil)
		if err != nil {
			slog.Error("failed to generate logs YAML", "scopeName", scopeName, "error", err)
			http.Error(w, "failed to generate logs YAML", http.StatusInternalServerError)
			return
		}

		spansYAML, err := weaver.GenerateMultiSpanYAML(telemetries, nil)
		if err != nil {
			slog.Error("failed to generate spans YAML", "scopeName", scopeName, "error", err)
			http.Error(w, "failed to generate spans YAML", http.StatusInternalServerError)
			return
		}

		// Check if we have any content to include in the ZIP
		hasMetrics := metricsYAML != ""
		hasLogs := logsYAML != ""
		hasSpans := spansYAML != ""

		if !hasMetrics && !hasLogs && !hasSpans {
			// No telemetries of any type found
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Generate registry manifest
		manifest := weaver.GenerateRegistryManifest(scopeName, "")

		// Create a ZIP file containing the YAML schemas and registry manifest
		var buf bytes.Buffer
		zipWriter := zip.NewWriter(&buf)

		// Create the metrics YAML file inside the ZIP (only if we have metrics)
		if hasMetrics {
			metricsFileName := scopeName + "-metrics.yaml"
			metricsFile, err := zipWriter.Create(metricsFileName)
			if err != nil {
				http.Error(w, "failed to create metrics file in zip", http.StatusInternalServerError)
				return
			}

			// Write the metrics YAML content to the file inside the ZIP
			_, err = metricsFile.Write([]byte(metricsYAML))
			if err != nil {
				http.Error(w, "failed to write metrics yaml to zip", http.StatusInternalServerError)
				return
			}
		}

		// Create the logs/events YAML file inside the ZIP (only if we have logs)
		if hasLogs {
			eventsFileName := scopeName + "-events.yaml"
			eventsFile, err := zipWriter.Create(eventsFileName)
			if err != nil {
				http.Error(w, "failed to create events file in zip", http.StatusInternalServerError)
				return
			}

			// Write the logs YAML content to the file inside the ZIP
			_, err = eventsFile.Write([]byte(logsYAML))
			if err != nil {
				http.Error(w, "failed to write events yaml to zip", http.StatusInternalServerError)
				return
			}
		}

		// Create the spans YAML file inside the ZIP (only if we have spans)
		if hasSpans {
			spansFileName := scopeName + "-spans.yaml"
			spansFile, err := zipWriter.Create(spansFileName)
			if err != nil {
				http.Error(w, "failed to create spans file in zip", http.StatusInternalServerError)
				return
			}

			// Write the spans YAML content to the file inside the ZIP
			_, err = spansFile.Write([]byte(spansYAML))
			if err != nil {
				http.Error(w, "failed to write spans yaml to zip", http.StatusInternalServerError)
				return
			}
		}

		// Create the registry manifest file inside the ZIP
		manifestFile, err := zipWriter.Create("registry_manifest.yaml")
		if err != nil {
			http.Error(w, "failed to create manifest file in zip", http.StatusInternalServerError)
			return
		}

		// Write the manifest content to the manifest file inside the ZIP
		_, err = manifestFile.Write([]byte(manifest))
		if err != nil {
			http.Error(w, "failed to write manifest to zip", http.StatusInternalServerError)
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
		w.Header().Set("Content-Disposition", "attachment; filename="+scopeName+".zip")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))

		// Write the ZIP file content to the response
		w.Write(buf.Bytes())
	}
}

func HandleScopeDashboardExport(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		scopeName := chi.URLParam(r, "scope")

		// Get all telemetries for this scope
		telemetries, err := schemaRepo.ListTelemetriesByScope(ctx, scopeName)
		if err != nil {
			slog.Error("failed to get telemetries for scope", "scopeName", scopeName, "error", err)
			http.Error(w, "failed to get telemetries for scope", http.StatusInternalServerError)
			return
		}

		// Check if scope exists but has no telemetries
		if len(telemetries) == 0 {
			// According to our specification: return 204 for scopes with no telemetries
			// We can't distinguish between "scope doesn't exist" and "scope has no telemetries"
			// from the current repository implementation, so we treat empty results as "no telemetries"
			w.WriteHeader(http.StatusNoContent)
			return
		}

		metricsSchema, err := weaver.GenerateMultiMetricYAML(telemetries, nil)
		if err != nil {
			slog.Error("failed to generate schema YAML", "scopeName", scopeName, "error", err)
			http.Error(w, "failed to generate metrics YAML", http.StatusInternalServerError)
			return
		}
		// logSchema, err := weaver.GenerateMultiLogYAML(telemetries, nil)
		// if err != nil {
		// 	slog.Error("failed to generate schema YAML", "scopeName", scopeName, "error", err)
		// 	http.Error(w, "failed to generate log YAML", http.StatusInternalServerError)
		// 	return
		// }
		// spanSchema, err := weaver.GenerateMultiSpanYAML(telemetries, nil)
		// if err != nil {
		// 	slog.Error("failed to generate schema YAML", "scopeName", scopeName, "error", err)
		// 	http.Error(w, "failed to generate span YAML", http.StatusInternalServerError)
		// 	return
		// }

		dashboards, err := weaver.RenderDashboards(ctx, scopeName, []string{metricsSchema})
		if err != nil {
			slog.Error("failed to generate dashboards", "scopeName", scopeName, "error", err)
			http.Error(w, "failed to generate dashboards", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(dashboards)
	}
}

func HandleScopeSchemaExport(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		scopeName := chi.URLParam(r, "scope")
		telemetryType := chi.URLParam(r, "type")

		// Get all telemetries for this scope
		telemetries, err := schemaRepo.ListTelemetriesByScope(ctx, scopeName)
		if err != nil {
			slog.Error("failed to get telemetries for scope", "scopeName", scopeName, "error", err)
			http.Error(w, "failed to get telemetries for scope", http.StatusInternalServerError)
			return
		}

		// Check if scope exists but has no telemetries
		if len(telemetries) == 0 {
			// According to our specification: return 204 for scopes with no telemetries
			// We can't distinguish between "scope doesn't exist" and "scope has no telemetries"
			// from the current repository implementation, so we treat empty results as "no telemetries"
			w.WriteHeader(http.StatusNoContent)
			return
		}

		outBuff := &bytes.Buffer{}
		var genYaml func(telemetries []schema.Telemetry, schemas map[string]*schema.TelemetrySchema) (string, error)
		switch telemetryType {
		case "metrics":
			genYaml = weaver.GenerateMultiMetricYAML
		case "logs":
			genYaml = weaver.GenerateMultiLogYAML
		case "spans":
			genYaml = weaver.GenerateMultiSpanYAML
		}
		schemaYAML, err := genYaml(telemetries, nil)
		if err != nil {
			slog.Error("failed to generate schema YAML", "scopeName", scopeName, "error", err)
			http.Error(w, "failed to generate schema YAML", http.StatusInternalServerError)
			return
		}
		outBuff.WriteString(schemaYAML)
		io.Copy(w, outBuff)
	}
}

// parseProducerNameVersion parses the producer name---version format
// Supports empty versions (e.g., "node-exporter---" for producers without version)
func parseProducerNameVersion(nameVersion string) (string, string, error) {
	// Find the last --- to separate name and version
	separator := "---"
	lastSep := strings.LastIndex(nameVersion, separator)
	if lastSep == -1 || lastSep == 0 {
		return "", "", fmt.Errorf("invalid producer format, expected name---version or name---")
	}

	// Allow empty version (URL ending with ---)
	if lastSep == len(nameVersion)-len(separator) {
		// URL ends with ---
		name := nameVersion[:lastSep]
		return name, "", nil
	}

	name := nameVersion[:lastSep]
	version := nameVersion[lastSep+len(separator):]

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

// HandleEntityList returns a paginated, filtered, and searched list of entities as JSON.
func HandleEntityList(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := ParseListQueryParams(r)
		entities, total, err := schemaRepo.ListEntities(ctx, params)
		if err != nil {
			slog.Error("failed to list entities", "error", err)
			http.Error(w, "failed to list entities", http.StatusInternalServerError)
			return
		}

		resp := ListResponse[schema.Entity]{
			Items:    entities,
			Total:    total,
			Page:     params.Page,
			PageSize: params.PageSize,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// HandleTelemetryEntityList returns all entities for a specific telemetry as JSON.
func HandleTelemetryEntityList(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		telemetryKey := chi.URLParam(r, "key")

		entities, err := schemaRepo.ListEntitiesByTelemetry(ctx, telemetryKey)
		if err != nil {
			slog.Error("failed to list entities for telemetry", "error", err, "telemetry_key", telemetryKey)
			http.Error(w, "failed to list entities for telemetry", http.StatusInternalServerError)
			return
		}

		resp := ListResponse[schema.Entity]{
			Items:    entities,
			Total:    len(entities),
			Page:     1,
			PageSize: len(entities),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// HandleScopeList returns a paginated, filtered, and searched list of scopes as JSON.
func HandleScopeList(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := ParseListQueryParams(r)
		scopes, total, err := schemaRepo.ListScopes(ctx, params)
		if err != nil {
			slog.Error("failed to list scopes", "error", err)
			http.Error(w, "failed to list scopes", http.StatusInternalServerError)
			return
		}

		resp := ListResponse[schema.Scope]{
			Items:    scopes,
			Total:    total,
			Page:     params.Page,
			PageSize: params.PageSize,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// HandleTelemetryScopeList returns all scopes for a specific telemetry as JSON.
func HandleTelemetryScopeList(schemaRepo repository.TelemetrySchemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		telemetryKey := chi.URLParam(r, "key")

		scopes, err := schemaRepo.ListScopesByTelemetry(ctx, telemetryKey)
		if err != nil {
			slog.Error("failed to list scopes for telemetry", "error", err, "telemetry_key", telemetryKey)
			http.Error(w, "failed to list scopes for telemetry", http.StatusInternalServerError)
			return
		}

		resp := ListResponse[schema.Scope]{
			Items:    scopes,
			Total:    len(scopes),
			Page:     1,
			PageSize: len(scopes),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
