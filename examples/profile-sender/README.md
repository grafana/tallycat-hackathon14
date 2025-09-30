# Profile Sender Example

This example demonstrates how to send OTLP profiles to TallyCat.

## How it works

1. **Profile Sender** (this program) → sends OTLP profiles via gRPC
3. **TallyCat** → receives profiles and extracts schemas

## Usage

1. **Start TallyCat server:**
   ```bash
   go run main.go server
   ```

2. **Run the profile sender:**
   ```bash
   cd examples/profile-sender
   go mod tidy
   go run main.go
   ```

## What the sender does

- Connects to the TallyCat on `localhost:4317` (gRPC)
- Sends 5 batches of sample profiles with different attributes
- Each profile includes:
  - Resource attributes (service info, batch number)
  - Scope information (profiler details)
  - Profile data with sample types (cpu_samples, cpu_time)
  - Dictionary with string table and attribute table

## Verify it worked

1. Check TallyCat's logs for profile ingestion
2. Query TallyCat's API to see the extracted schemas:
   ```bash
   curl "http://localhost:8080/api/v1/telemetry-schemas?type=profile"
   ```

## Expected Profile Schema

The sender creates profiles with these characteristics:
- **Profile Types:** `cpu_samples` (DELTA), `cpu_time` (CUMULATIVE)
- **Units:** `count`, `nanoseconds`
- **Attributes:** `cpu`, `mode`, `thread_id`
- **Resource:** service.name, service.version, deployment.environment, batch.number
