# 🐾 TallyCat

**Telemetry Governance for OpenTelemetry**

TallyCat is an open-source telemetry governance platform for teams using OpenTelemetry. It helps you understand what telemetry is being emitted, where it comes from, and how it changes over time.

TallyCat builds a structured view of your telemetry by extracting and tracking metadata—schemas, field types, source context, and usage patterns. This metadata powers a governance layer that helps teams improve observability hygiene, reduce duplication, track ownership, and control costs.

> 🐱 **Why the name "TallyCat"?**  
> The name combines "tally" — to count and classify — with "cat" as in catalog.  
> TallyCat catalogs the structure of your telemetry so you can govern and optimize it.

---

## ❓ Why observability needs data governance

Modern observability pipelines produce huge amounts of telemetry. But most teams can't answer basic questions like:

- What telemetry are we emitting?
- Which fields are actually used?
- How has our schema changed?
- Where is all this cardinality coming from?

Most teams today operate in a "telemetry last" world — emitting data without structure or oversight. This leads to hidden costs, inconsistent fields, and poor documentation. TallyCat brings visibility and structure, enabling teams to reason about their telemetry just like they would with APIs or database schemas.

🛠️ Tools like OpenMetadata and DataHub solve this for analytics.  
🔎 TallyCat brings the same governance principles to observability — where they're still missing.

With TallyCat, you get visibility and control over the structure and cost of the data you send to your observability backend.

---

## ✅ What TallyCat helps with

- 📦 Discover all emitted signals, grouped by schema
- 🔬 Track field names, types, and sources (`resource`, `scope`, `data`)
- 📈 Detect schema changes and version them over time
- 🧹 Reduce duplicate or high-cardinality signals
- 👥 Map signals to owners, teams, or workloads
- 🛡️ Prepare for policy enforcement and budget limits
- 🚀 Move toward schema-first observability, even with legacy or live telemetry

---

## ✨ Core features

- 🧠 Real-time schema inference from OTLP logs, metrics, and spans
- 🧾 Field-level typing and source attribution
- 🔁 Schema versioning and seen counts

---

## 🧭 How it fits

TallyCat does not replace your observability backend.

It works alongside tools like Prometheus, Jaeger, OpenSearch, focusing on telemetry metadata, not telemetry storage.  
Think of it like **dbt** or **OpenMetadata**, but for **OpenTelemetry signals**.

---

## 🚧 Project status

TallyCat is under active development. The MVP includes:

- ✅ Schema inference for OTLP data
- ✅ Telemetry Schema Version Assignment
- ✅ Telemetry History

Coming next:

TBD

---

## 🤝 Contributing

We welcome contributions!  
Check out [CONTRIBUTING.md](CONTRIBUTING.md) to get started.