package testutil

import (
	"go.opentelemetry.io/collector/pdata/pprofile"
	profilespb "go.opentelemetry.io/proto/otlp/collector/profiles/v1development"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	profilepb "go.opentelemetry.io/proto/otlp/profiles/v1development"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
)

// ConvertProfileToRequest converts pprofile.Profiles to ExportProfilesServiceRequest
func ConvertProfileToRequest(pd pprofile.Profiles) *profilespb.ExportProfilesServiceRequest {
	request := &profilespb.ExportProfilesServiceRequest{
		ResourceProfiles: make([]*profilepb.ResourceProfiles, 0, pd.ResourceProfiles().Len()),
		Dictionary:       nil, // Will be set by the golden package if present in the YAML
	}

	for i := 0; i < pd.ResourceProfiles().Len(); i++ {
		rp := pd.ResourceProfiles().At(i)
		resourceProfiles := &profilepb.ResourceProfiles{
			Resource: &resourcepb.Resource{
				Attributes: convertAttributes(rp.Resource().Attributes()),
			},
			ScopeProfiles: make([]*profilepb.ScopeProfiles, 0, rp.ScopeProfiles().Len()),
			SchemaUrl:     rp.SchemaUrl(),
		}

		for j := 0; j < rp.ScopeProfiles().Len(); j++ {
			sp := rp.ScopeProfiles().At(j)
			scopeProfiles := &profilepb.ScopeProfiles{
				Scope: &commonpb.InstrumentationScope{
					Name:       sp.Scope().Name(),
					Version:    sp.Scope().Version(),
					Attributes: convertAttributes(sp.Scope().Attributes()),
				},
				Profiles:  make([]*profilepb.Profile, 0, sp.Profiles().Len()),
				SchemaUrl: sp.SchemaUrl(),
			}

			for k := 0; k < sp.Profiles().Len(); k++ {
				p := sp.Profiles().At(k)

				for l := 0; l < p.SampleType().Len(); l++ {
					st := p.SampleType().At(l)

					profile := &profilepb.Profile{
						AttributeIndices: make([]int32, 0, p.AttributeIndices().Len()),
						SampleType: &profilepb.ValueType{
							TypeStrindex:           int32(st.TypeStrindex()),
							UnitStrindex:           int32(st.UnitStrindex()),
							AggregationTemporality: profilepb.AggregationTemporality(st.AggregationTemporality()),
						},
					}

					// Convert attribute indices
					for m := 0; m < p.AttributeIndices().Len(); m++ {
						profile.AttributeIndices = append(profile.AttributeIndices, int32(p.AttributeIndices().At(m)))
					}

					scopeProfiles.Profiles = append(scopeProfiles.Profiles, profile)
				}
			}

			resourceProfiles.ScopeProfiles = append(resourceProfiles.ScopeProfiles, scopeProfiles)
		}

		request.ResourceProfiles = append(request.ResourceProfiles, resourceProfiles)
	}

	return request
}
