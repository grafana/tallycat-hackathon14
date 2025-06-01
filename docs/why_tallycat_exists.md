# Why TallyCat Exists

Observability is overflowing with data — but very little structure.

Most teams emit logs, metrics, and traces without clear schemas, ownership, or lifecycle management. As a result, they face:
- Undocumented telemetry
- High cardinality surprises
- Duplicated signals across pipelines
- Confusion around what’s still used — or by whom
- Escalating costs with no governance layer

## Inspired by OpenAPI, dbt, and OpenMetadata

In the analytics and API ecosystems, we've seen the rise of:
- **OpenAPI** for API schema-first design
- **dbt** for versioned models and lineage in data pipelines
- **OpenMetadata** and **DataHub** for structured metadata governance

But observability still lags behind — especially in environments built on OpenTelemetry.

## TallyCat’s Vision

TallyCat helps teams bring structure, visibility, and control to observability pipelines.

It builds a **live catalog of telemetry schemas**, inferred from the data you already emit — and tracks how they evolve over time. This enables:
- Schema discovery
- Field typing and provenance
- Schema versioning
- Ownership attribution
- Cost governance and signal hygiene

## Schema-First Without Starting From Scratch

The OpenTelemetry community is moving toward **schema-first observability** — with tools like [Weaver](https://github.com/open-telemetry/opentelemetry-weaver) and native schema support in `schema_url`.

But most teams already have huge volumes of telemetry in production.

TallyCat fills this gap by reverse-engineering schemas from your current telemetry — making schema-first adoption possible even if you didn’t start that way.

## Not a Backend. A Control Plane.

TallyCat doesn’t store telemetry. It doesn’t replace your backend.

Instead, it’s the **metadata control plane** for your telemetry layer:
- Works alongside Prometheus, Perses, Jaeger, etc.
- Focuses on telemetry metadata, not raw data
- Acts like `dbt` or `OpenMetadata` — but for OpenTelemetry

---

TallyCat exists to help you **understand, govern, and evolve** the telemetry you already produce — so you can build a cleaner, more cost-effective observability practice.