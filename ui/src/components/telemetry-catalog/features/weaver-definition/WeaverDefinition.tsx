import { Code, Download } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { Attribute, Telemetry } from '@/types/telemetry'
import { TelemetryType } from '@/types/telemetry'
import type { TelemetrySchema } from '@/types/telemetry-schema'

interface WeaverDefinitionProps {
  telemetry: Telemetry
  schema: TelemetrySchema
}

export function WeaverDefinition({ telemetry, schema }: WeaverDefinitionProps) {
  const formatAttribute = (attribute: Attribute) => {
    const id = attribute.name || ''
    const type = convertAttributeType(attribute.type || '')
    return [
      `      - id: ${id}`,
      `        type: ${type}`,
      `        requirement_level: recommended`,
      `        brief: "${attribute.brief || ''}"`,
    ].join('\n')
  }

  // Convert internal attribute types to Weaver-compatible types
  const convertAttributeType = (attrType: string): string => {
    switch (attrType) {
      case 'Str':
        return 'string'
      case 'Bool':
        return 'boolean'
      case 'Int':
        return 'int'
      case 'Double':
        return 'double'
      case 'Map':
        return 'string'
      case 'Slice':
        return 'string[]'
      case 'Bytes':
        return 'string'
      default:
        return 'string'
    }
  }

  // Convert metric type to instrument
  const convertMetricTypeToInstrument = (metricType: string): string => {
    switch (metricType) {
      case 'Gauge':
        return 'gauge'
      case 'Sum':
        return 'sum'
      case 'Histogram':
        return 'histogram'
      case 'ExponentialHistogram':
        return 'histogram'
      case 'Summary':
        return 'histogram'
      default:
        return 'gauge'
    }
  }

  // Generate log-specific attributes
  const generateLogAttributes = (): string[] => {
    const logAttributes = [
      [
        '      - id: log.severity.number',
        '        type: int',
        '        requirement_level: recommended',
        '        brief: "Log severity number"',
      ].join('\n'),
      [
        '      - id: log.severity.text',
        '        type: string',
        '        requirement_level: recommended',
        '        brief: "Log severity text"',
      ].join('\n'),
    ]

    // Add log body if present
    if (telemetry.logBody && telemetry.logBody.trim() !== '') {
      logAttributes.push([
        '      - id: log.body',
        '        type: string',
        '        requirement_level: recommended',
        '        brief: "Log body content"',
      ].join('\n'))
    }

    return logAttributes
  }

  const generateMetricYaml = (): string => {
    const yamlLines = [
      'groups:',
      `  - id: metric.${telemetry.schemaKey}`,
      '    type: metric',
      `    metric_name: ${telemetry.schemaKey}`,
      `    brief: "${telemetry.brief || ''}"`,
      `    instrument: ${convertMetricTypeToInstrument(telemetry.metricType)}`,
      `    unit: "${telemetry.metricUnit || ''}"`,
    ]

    // Filter for DataPoint attributes
    const dataPointAttributes = schema.attributes.filter(
      (attribute) => attribute.source === 'DataPoint'
    )

    if (dataPointAttributes.length > 0) {
      yamlLines.push('    attributes:')
      dataPointAttributes.forEach((attribute) => {
        yamlLines.push(formatAttribute(attribute))
      })
    }

    return yamlLines.join('\n')
  }

  const generateLogEventYaml = (): string => {
    // Determine event name: use logEventName if available, otherwise use schemaKey
    const eventName = telemetry.logEventName || telemetry.schemaKey

    const yamlLines = [
      'groups:',
      `  - id: event.${eventName}`,
      '    type: event',
      `    event_name: ${eventName}`,
      `    brief: "${telemetry.brief || ''}"`,
      '    attributes:',
    ]

    // Add LogRecord attributes first
    const logRecordAttributes = schema.attributes.filter(
      (attribute) => attribute.source === 'LogRecord'
    )

    logRecordAttributes.forEach((attribute) => {
      yamlLines.push(formatAttribute(attribute))
    })

    // Add log-specific attributes
    const logAttributes = generateLogAttributes()
    logAttributes.forEach((attr) => {
      yamlLines.push(attr)
    })

    return yamlLines.join('\n')
  }

  // Convert span kind to Weaver-compatible span_kind values
  const convertSpanKindToWeaver = (spanKind: string): string => {
    switch (spanKind) {
      case 'client':
        return 'client'
      case 'server':
        return 'server'
      case 'producer':
        return 'producer'
      case 'consumer':
        return 'consumer'
      case 'internal':
        return 'internal'
      default:
        return 'internal' // Default to internal for unknown span kinds
    }
  }

  const generateSpanYaml = (): string => {
    const yamlLines = [
      'groups:',
      `  - id: span.${telemetry.schemaKey}`,
      '    type: span',
      `    brief: "${telemetry.brief || ''}"`,
      '    stability: stable',
      `    span_kind: ${convertSpanKindToWeaver(telemetry.spanKind || 'internal')}`,
    ]

    // Filter for Span attributes
    const spanAttributes = schema.attributes.filter(
      (attribute) => attribute.source === 'Span'
    )

    if (spanAttributes.length > 0) {
      yamlLines.push('    attributes:')
      spanAttributes.forEach((attribute) => {
        yamlLines.push(formatAttribute(attribute))
      })
    }

    return yamlLines.join('\n')
  }

  const generateWeaverYaml = (): string => {
    switch (telemetry.telemetryType) {
      case TelemetryType.Log:
        return generateLogEventYaml()
      case TelemetryType.Metric:
        return generateMetricYaml()
      case TelemetryType.Span:
        return generateSpanYaml()
      default:
        // Default to metric for backwards compatibility
        return generateMetricYaml()
    }
  }

  const getFilePrefix = (): string => {
    switch (telemetry.telemetryType) {
      case TelemetryType.Log:
        return telemetry.logEventName || telemetry.schemaKey
      case TelemetryType.Metric:
        return telemetry.schemaKey
      case TelemetryType.Span:
        return telemetry.schemaKey
      default:
        return telemetry.schemaKey
    }
  }

  const getTelemetryTypeLabel = (): string => {
    switch (telemetry.telemetryType) {
      case TelemetryType.Log:
        return 'log event'
      case TelemetryType.Metric:
        return 'metric'
      case TelemetryType.Span:
        return 'span'
      default:
        return 'telemetry'
    }
  }

  const handleCopyYaml = () => {
    const yamlContent = generateWeaverYaml()
    navigator.clipboard.writeText(yamlContent)
  }

  const handleDownloadYaml = () => {
    const yamlContent = generateWeaverYaml()
    const blob = new Blob([yamlContent], { type: 'text/yaml' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${getFilePrefix()}.yaml`
    a.click()
  }

  return (
    <div className="space-y-4">
      <div>
        <h3 className="text-lg font-semibold flex items-center gap-2">
          <Code className="h-5 w-5 text-primary" />
          OpenTelemetry Weaver Definition
        </h3>
        <p className="text-sm text-muted-foreground">
          Semantic convention definition in Weaver format for OpenTelemetry {getTelemetryTypeLabel()} instrumentation
        </p>
      </div>

      <div className="relative">
        <pre className="bg-muted/50 rounded-lg p-4 text-xs overflow-x-auto border max-h-96">
          <code className="language-yaml">{generateWeaverYaml()}</code>
        </pre>

        <div className="absolute top-2 right-2 flex gap-2">
          <Button
            variant="outline"
            size="sm"
            className="h-7 px-2 text-xs"
            onClick={handleCopyYaml}
          >
            Copy YAML
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="h-7 px-2 text-xs"
            onClick={handleDownloadYaml}
          >
            <Download className="h-3 w-3 mr-1" />
            Download
          </Button>
        </div>
      </div>
    </div>
  )
} 