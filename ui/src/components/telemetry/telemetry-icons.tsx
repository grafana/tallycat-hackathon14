import { BarChart, Hash, BarChart3, Layers, FileText, Activity, Code } from 'lucide-react'
import { TelemetryType } from '@/types/telemetry'

export const DataTypeIcon = ({ dataType }: { dataType: string }) => {
  if (!dataType) {
    return <Code className="h-4 w-4 text-gray-400" />
  }

  switch (dataType.toLowerCase()) {
    case "gauge":
      return <BarChart className="h-4 w-4 text-blue-400" />
    case "counter":
      return <Hash className="h-4 w-4 text-blue-400" />
    case "histogram":
      return <BarChart3 className="h-4 w-4 text-blue-400" />
    case "summary":
      return <BarChart3 className="h-4 w-4 text-blue-400" />
    case "exponentialhistogram":
      return <BarChart3 className="h-4 w-4 text-blue-400" />
    case "structured":
      return <Layers className="h-4 w-4 text-green-400" />
    case "unstructured":
      return <FileText className="h-4 w-4 text-green-400" />
    case "span":
      return <Activity className="h-4 w-4 text-purple-400" />
    default:
      return <Code className="h-4 w-4 text-gray-400" />
  }
}

export const TelemetryTypeIcon = ({ type }: { type: TelemetryType }) => {
  switch (type) {
    case TelemetryType.Metric:
      return <BarChart3 className="h-5 w-5 text-blue-500" />
    case TelemetryType.Log:
      return <FileText className="h-5 w-5 text-green-500" />
    case TelemetryType.Trace:
      return <Activity className="h-5 w-5 text-purple-500" />
    default:
      return <Code className="h-5 w-5 text-gray-500" />
  }
} 