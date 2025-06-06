import { Code, Download } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { Attribute, Telemetry } from '@/types/telemetry'
import type { TelemetrySchema } from '@/types/telemetry-schema'

interface WeaverDefinitionProps {
  telemetry: Telemetry
  schema: TelemetrySchema
}

export function WeaverDefinition({ telemetry, schema }: WeaverDefinitionProps) {
  const formatAttribute = (attribute: Attribute) => {
    const id = attribute.name || ''
    const type = attribute.type || ''
    return [
      `    - id: ${id}`,
      `      type: ${type}`,
      `      requirement_level: recommended`,
    ].join('\n')
  }

  const generateWeaverYaml = () => {
    const yamlLines = [
      'groups:',
      '  - id: metric.' + telemetry.schemaKey,
      '    type: metric',
      '    metric_name: ' + telemetry.schemaKey,
      '    brief: ' + telemetry.brief,
      '    instrument: ' + telemetry.metricType,
      '    unit: ' + telemetry.metricUnit,
      '    attributes:',
      schema.attributes.filter((attribute) => attribute.source === 'DataPoint').map(formatAttribute).join('\n')
    ]

    return yamlLines.join('\n')
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
    a.download = `${telemetry.schemaKey}.yaml`
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
          Semantic convention definition in Weaver format for OpenTelemetry instrumentation
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