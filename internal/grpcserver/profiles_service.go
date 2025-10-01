package grpcserver

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/collector/pdata/pprofile"
	profilespb "go.opentelemetry.io/proto/otlp/collector/profiles/v1development"

	"github.com/tallycat/tallycat/internal/repository"
	"github.com/tallycat/tallycat/internal/schema"
)

type ProfilesServiceServer struct {
	profilespb.UnimplementedProfilesServiceServer
	schemaRepo repository.TelemetrySchemaRepository
}

func NewProfilesServiceServer(schemaRepo repository.TelemetrySchemaRepository) *ProfilesServiceServer {
	return &ProfilesServiceServer{
		schemaRepo: schemaRepo,
	}
}

func (s *ProfilesServiceServer) Export(ctx context.Context, req *profilespb.ExportProfilesServiceRequest) (*profilespb.ExportProfilesServiceResponse, error) {
	slog.Info("Received profiles export request",
		"resource_profiles_count", len(req.ResourceProfiles),
		"has_dictionary", req.Dictionary != nil)

	if req.Dictionary != nil {
		slog.Info("Dictionary details",
			"string_table_size", len(req.Dictionary.StringTable),
			"attribute_table_size", len(req.Dictionary.AttributeTable))
	}

	profiles := pprofile.NewProfiles()
	rps := profiles.ResourceProfiles()
	rps.EnsureCapacity(len(req.ResourceProfiles))

	for _, rp := range req.ResourceProfiles {
		resourceProfile := rps.AppendEmpty()
		resourceProfile.SetSchemaUrl(rp.SchemaUrl)

		// Convert resource attributes
		if rp.Resource != nil {
			for _, attr := range rp.Resource.Attributes {
				resourceProfile.Resource().Attributes().PutStr(attr.Key, attr.Value.GetStringValue())
			}
		}

		// Convert scope profiles
		sps := resourceProfile.ScopeProfiles()
		sps.EnsureCapacity(len(rp.ScopeProfiles))

		for _, sp := range rp.ScopeProfiles {
			scopeProfile := sps.AppendEmpty()
			scopeProfile.SetSchemaUrl(sp.SchemaUrl)

			// Convert scope
			if sp.Scope != nil {
				scopeProfile.Scope().SetName(sp.Scope.Name)
				scopeProfile.Scope().SetVersion(sp.Scope.Version)
				scopeProfile.SetSchemaUrl(sp.SchemaUrl)
				for _, attr := range sp.Scope.Attributes {
					scopeProfile.Scope().Attributes().PutStr(attr.Key, attr.Value.GetStringValue())
				}
			}

			// Convert profiles
			ps := scopeProfile.Profiles()
			ps.EnsureCapacity(len(sp.Profiles))

			for _, p := range sp.Profiles {
				profile := ps.AppendEmpty()
				profile.AttributeIndices().Append(p.AttributeIndices...)

				if p.SampleType != nil {
					sampleType := profile.SampleType().AppendEmpty()
					sampleType.SetAggregationTemporality(pprofile.AggregationTemporality(p.SampleType.AggregationTemporality))
					sampleType.SetTypeStrindex(int32(p.SampleType.TypeStrindex))
					sampleType.SetUnitStrindex(int32(p.SampleType.UnitStrindex))
				}

			}
		}
	}

	// Extract schemas from the converted profiles
	slog.Info("Extracting schemas from profiles")
	schemas := schema.ExtractFromProfiles(profiles, req.Dictionary)
	slog.Info("Schema extraction completed", "schemas_count", len(schemas))

	for i, schema := range schemas {
		slog.Info("Extracted schema",
			"index", i,
			"schema_key", schema.SchemaKey,
			"telemetry_type", schema.TelemetryType,
			"attributes_count", len(schema.Attributes))
	}

	if err := s.schemaRepo.RegisterTelemetrySchemas(ctx, schemas); err != nil {
		slog.Error("failed to register schemas", "error", err, "signal", "profiles")
		return nil, err
	}

	slog.Info("Successfully registered telemetry schemas", "count", len(schemas))
	return &profilespb.ExportProfilesServiceResponse{}, nil
}
