package main

import (
	"context"
	"fmt"
	"log"
	"time"

	profilespb "go.opentelemetry.io/proto/otlp/collector/profiles/v1development"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	profilepb "go.opentelemetry.io/proto/otlp/profiles/v1development"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()

	// Connect directly to TallyCat (bypassing collector for debugging)
	fmt.Println("Attempting to connect directly to TallyCat at localhost:4317...")
	conn, err := grpc.Dial("localhost:4317", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to collector: %v", err)
	}
	defer conn.Close()

	// Create profiles service client
	fmt.Println("Creating profiles service client...")
	client := profilespb.NewProfilesServiceClient(conn)
	fmt.Println("Profiles client created successfully")

	// Generate and send profiles of different types
	fmt.Println("Starting to send profiles directly to TallyCat...")

	profileTypes := []string{"cpu", "memory", "locks", "exceptions", "allocs"}

	for i, profileType := range profileTypes {
		fmt.Printf("Sending %s profile batch %d directly to TallyCat...\n", profileType, i+1)

		// Create a profile request for the specific type
		req := createProfileRequestByType(profileType, i)

		// Debug: Print request details
		fmt.Printf("  Request details: %d resource profiles, dictionary with %d strings, %d attributes\n",
			len(req.ResourceProfiles),
			len(req.Dictionary.StringTable),
			len(req.Dictionary.AttributeTable))

		// Send the profile
		resp, err := client.Export(ctx, req)
		if err != nil {
			log.Printf("Failed to export profiles: %v", err)
		} else {
			fmt.Printf("Successfully sent %s profile batch %d, response: %v\n", profileType, i+1, resp)
		}

		time.Sleep(2 * time.Second)
	}

	fmt.Println("Finished sending profiles directly to TallyCat. Check TallyCat logs and API!")
}

func createProfileRequestByType(profileType string, batchNum int) *profilespb.ExportProfilesServiceRequest {
	// Define profile type specific data
	var sampleType1, sampleType2, unit1, unit2 string
	var attr1Key, attr1Value, attr2Key, attr2Value string

	switch profileType {
	case "cpu":
		sampleType1, sampleType2 = "cpu_samples", "cpu_time"
		unit1, unit2 = "count", "nanoseconds"
		attr1Key, attr1Value = "cpu_core", fmt.Sprintf("core_%d", batchNum)
		attr2Key, attr2Value = "thread_state", "running"
	case "memory":
		sampleType1, sampleType2 = "memory_allocations", "memory_usage"
		unit1, unit2 = "count", "bytes"
		attr1Key, attr1Value = "allocation_type", "heap"
		attr2Key, attr2Value = "object_size", fmt.Sprintf("%d_bytes", 1024*(batchNum+1))
	case "locks":
		sampleType1, sampleType2 = "lock_acquisitions", "lock_wait_time"
		unit1, unit2 = "count", "nanoseconds"
		attr1Key, attr1Value = "lock_type", "mutex"
		attr2Key, attr2Value = "contention_level", "high"
	case "exceptions":
		sampleType1, sampleType2 = "exception_count", "exception_overhead"
		unit1, unit2 = "count", "nanoseconds"
		attr1Key, attr1Value = "exception_type", "runtime_error"
		attr2Key, attr2Value = "stack_depth", fmt.Sprintf("%d", 10+batchNum)
	case "allocs":
		sampleType1, sampleType2 = "alloc_objects", "alloc_space"
		unit1, unit2 = "count", "bytes"
		attr1Key, attr1Value = "alloc_site", fmt.Sprintf("site_%d", batchNum)
		attr2Key, attr2Value = "gc_generation", "gen_1"
	default:
		sampleType1, sampleType2 = "profile_samples", "profile_time"
		unit1, unit2 = "count", "nanoseconds"
		attr1Key, attr1Value = "profile_type", "generic"
		attr2Key, attr2Value = "batch_id", fmt.Sprintf("%d", batchNum)
	}

	return &profilespb.ExportProfilesServiceRequest{
		ResourceProfiles: []*profilepb.ResourceProfiles{
			{
				SchemaUrl: "https://opentelemetry.io/schemas/1.21.0",
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key: "service.name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: fmt.Sprintf("profile-sender-%s", profileType),
								},
							},
						},
						{
							Key: "service.version",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "1.0.0",
								},
							},
						},
						{
							Key: "deployment.environment",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "development",
								},
							},
						},
					},
				},
				ScopeProfiles: []*profilepb.ScopeProfiles{
					{
						SchemaUrl: "https://opentelemetry.io/schemas/1.21.0",
						Scope: &commonpb.InstrumentationScope{
							Name:    fmt.Sprintf("%s-profiler", profileType),
							Version: "1.0.0",
							Attributes: []*commonpb.KeyValue{
								{
									Key: "instrumentation.provider",
									Value: &commonpb.AnyValue{
										Value: &commonpb.AnyValue_StringValue{
											StringValue: "custom",
										},
									},
								},
							},
						},
						Profiles: []*profilepb.Profile{
							{
								AttributeIndices: []int32{0, 1}, // Referring to dictionary indices
								SampleType: &profilepb.ValueType{
									TypeStrindex:           1, // sampleType1 in string table
									UnitStrindex:           2, // unit1 in string table
									AggregationTemporality: profilepb.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA,
								},
							},
							{
								AttributeIndices: []int32{0, 1}, // Referring to dictionary indices
								SampleType: &profilepb.ValueType{
									TypeStrindex:           3, // sampleType2 in string table
									UnitStrindex:           4, // unit2 in string table
									AggregationTemporality: profilepb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
								},
							},
						},
					},
				},
			},
		},
		Dictionary: &profilepb.ProfilesDictionary{
			StringTable: []string{
				"",          // Index 0: empty string (required)
				sampleType1, // Index 1
				unit1,       // Index 2
				sampleType2, // Index 3
				unit2,       // Index 4
				attr1Key,    // Index 5
				attr2Key,    // Index 6
			},
			AttributeTable: []*profilepb.KeyValueAndUnit{
				{ // Index 0
					KeyStrindex: 5, // attr1Key in string table
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{
							StringValue: attr1Value,
						},
					},
					UnitStrindex: 0, // no unit
				},
				{ // Index 1
					KeyStrindex: 6, // attr2Key in string table
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{
							StringValue: attr2Value,
						},
					},
					UnitStrindex: 0, // no unit
				},
			},
		},
	}
}

func createSampleProfileRequest(batchNum int) *profilespb.ExportProfilesServiceRequest {
	return &profilespb.ExportProfilesServiceRequest{
		ResourceProfiles: []*profilepb.ResourceProfiles{
			{
				SchemaUrl: "https://opentelemetry.io/schemas/1.21.0",
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key: "service.name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "profile-sender",
								},
							},
						},
						{
							Key: "service.version",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "1.0.0",
								},
							},
						},
						{
							Key: "deployment.environment",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "development",
								},
							},
						},
					},
				},
				ScopeProfiles: []*profilepb.ScopeProfiles{
					{
						SchemaUrl: "https://opentelemetry.io/schemas/1.21.0",
						Scope: &commonpb.InstrumentationScope{
							Name:    "go-profiler",
							Version: "1.0.0",
							Attributes: []*commonpb.KeyValue{
								{
									Key: "instrumentation.provider",
									Value: &commonpb.AnyValue{
										Value: &commonpb.AnyValue_StringValue{
											StringValue: "custom",
										},
									},
								},
							},
						},
						Profiles: []*profilepb.Profile{
							{
								AttributeIndices: []int32{0, 1, 2}, // Referring to dictionary indices
								SampleType: &profilepb.ValueType{
									TypeStrindex:           1, // "cpu_samples" in string table
									UnitStrindex:           2, // "count" in string table
									AggregationTemporality: profilepb.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA,
								},
							},
							{
								AttributeIndices: []int32{0, 1, 2}, // Referring to dictionary indices
								SampleType: &profilepb.ValueType{
									TypeStrindex:           3, // "cpu_time" in string table
									UnitStrindex:           4, // "nanoseconds" in string table
									AggregationTemporality: profilepb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
								},
							},
						},
					},
				},
			},
		},
		Dictionary: &profilepb.ProfilesDictionary{
			StringTable: []string{
				"",                   // Index 0: empty string (required)
				"cpu_samples",        // Index 1
				"count",              // Index 2
				"memory_allocations", // Index 3
				"bytes",              // Index 4
				"cpu",                // Index 5
				"mode",               // Index 6
				"thread_id",          // Index 7
			},
			AttributeTable: []*profilepb.KeyValueAndUnit{
				{ // Index 0
					KeyStrindex: 5, // "cpu" in string table
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{
							StringValue: fmt.Sprintf("cpu_%d", batchNum),
						},
					},
					UnitStrindex: 0, // no unit
				},
				{ // Index 1
					KeyStrindex: 6, // "mode" in string table
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{
							StringValue: "user",
						},
					},
					UnitStrindex: 0, // no unit
				},
				{ // Index 2
					KeyStrindex: 7, // "thread_id" in string table
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_IntValue{
							IntValue: int64(1000 + batchNum),
						},
					},
					UnitStrindex: 0, // no unit
				},
			},
		},
	}
}
