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

				for _, sType := range p.SampleType {
					sampleType := profile.SampleType().AppendEmpty()
					sampleType.SetAggregationTemporality(pprofile.AggregationTemporality(sType.AggregationTemporality))
					sampleType.SetTypeStrindex(int32(sType.TypeStrindex))
					sampleType.SetUnitStrindex(int32(sType.UnitStrindex))
				}

			}
		}
	}

	// Extract schemas from the converted profiles
	schemas := schema.ExtractFromProfiles(profiles, req.Dictionary)

	if err := s.schemaRepo.RegisterTelemetrySchemas(ctx, schemas); err != nil {
		slog.Error("failed to register schemas", "error", err)
		return nil, err
	}

	return &profilespb.ExportProfilesServiceResponse{}, nil
}
