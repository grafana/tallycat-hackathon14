export interface FacetOption {
  id: string
  name: string
}

export interface TechnicalFacet {
  id: string
  name: string
  options: FacetOption[]
}

export const technicalFacets: TechnicalFacet[] = [
  {
    id: 'type',
    name: 'Telemetry Type',
    options: [
      { id: 'metric', name: 'Metric' },
      { id: 'log', name: 'Log' },
      { id: 'trace', name: 'Trace' },
    ],
  },
  {
    id: 'dataType',
    name: 'Data Type',
    options: [
      { id: 'gauge', name: 'Gauge' },
      { id: 'counter', name: 'Counter' },
      { id: 'histogram', name: 'Histogram' },
      { id: 'structured', name: 'Structured' },
      { id: 'unstructured', name: 'Unstructured' },
      { id: 'span', name: 'Span' },
    ],
  },
  {
    id: 'format',
    name: 'Format',
    options: [
      { id: 'OTLP', name: 'OTLP' },
      { id: 'Prometheus', name: 'Prometheus' },
      { id: 'JSON', name: 'JSON' },
      { id: 'Zipkin', name: 'Zipkin' },
    ],
  },
  {
    id: 'source',
    name: 'Source',
    options: [
      { id: 'OpenTelemetry Collector', name: 'OpenTelemetry Collector' },
      { id: 'Prometheus', name: 'Prometheus' },
      { id: 'Fluentd', name: 'Fluentd' },
      { id: 'Jaeger', name: 'Jaeger' },
    ],
  },
  {
    id: 'cardinality',
    name: 'Cardinality',
    options: [
      { id: 'low', name: 'Low' },
      { id: 'medium', name: 'Medium' },
      { id: 'high', name: 'High' },
    ],
  },
  {
    id: 'status',
    name: 'Status',
    options: [
      { id: 'active', name: 'Active' },
      { id: 'draft', name: 'Draft' },
      { id: 'deprecated', name: 'Deprecated' },
    ],
  },
]
