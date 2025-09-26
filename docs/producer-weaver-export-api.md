# Producer Weaver Schema Export API

This document describes the new producer-based Weaver schema export endpoint that bundles all metrics from a specific producer into a single ZIP file.

## Overview

The `/api/v1/producers/{producerName}---{producerVersion}/weaver-schema.zip` endpoint allows you to download all Weaver schema definitions for metrics produced by a specific service in a single ZIP file. `producerVersion` is not required.

## Endpoint Details

**URL Pattern**: `GET /api/v1/producers/{producerName}---{producerVersion}/weaver-schema.zip`

**URL Parameters**:
- `{producerName}---{producerVersion}`: Producer identifier in the format `name---version`
  - Example: `my-service---1.0.0`, `payment-gateway---2.1.3`

## Response Codes

| Code | Description | Response Body |
|------|-------------|---------------|
| `200 OK` | Producer found with metrics | ZIP file containing YAML |
| `204 No Content` | Producer has no metrics or not found | Empty |
| `400 Bad Request` | Invalid producer format | Error message |
| `500 Internal Server Error` | Server processing error | Error message |

## Response Headers

For successful requests (200 OK):
```
Content-Type: application/zip
Content-Disposition: attachment; filename={producerName}-{producerVersion}.zip
Content-Length: {size}
```

## Usage Examples

### 1. Download Schemas for a Service

```bash
# Download all metrics for my-service version 1.0.0
curl -X GET "http://localhost:8080/api/v1/producers/my-service---1.0.0/weaver-schema.zip" \
     -o my-service-1.0.0.zip
```

## ZIP File Contents

The downloaded ZIP file contains a two YAML files:
* One named `{producerName}---{producerVersion}.yaml` with all metrics from that producer in OpenTelemetry Weaver format.
* One named `registry_manifest.yaml` describing the producer.
