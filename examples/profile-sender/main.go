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

	// Generate and send profiles
	fmt.Println("Starting to send profiles directly to TallyCat...")
	for i := 0; i < 5; i++ {
		fmt.Printf("Sending profile batch %d directly to TallyCat...\n", i+1)

		// Create a profile request
		req := createSampleProfileRequest(i)

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
			fmt.Printf("Successfully sent profile batch %d, response: %v\n", i+1, resp)
		}

		time.Sleep(2 * time.Second)
	}

	fmt.Println("Finished sending profiles directly to TallyCat. Check TallyCat logs and API!")
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
				"",            // Index 0: empty string (required)
				"cpu_samples", // Index 1
				"count",       // Index 2
				"cpu_time",    // Index 3
				"nanoseconds", // Index 4
				"cpu",         // Index 5
				"mode",        // Index 6
				"thread_id",   // Index 7
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
