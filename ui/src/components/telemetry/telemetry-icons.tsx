import {
  BarChart,
  Hash,
  BarChart3,
  Layers,
  FileText,
  Activity,
  Code,
  Server,
  Smartphone,
  ArrowUp,
  ArrowDown,
  Zap,
  Logs,
  Cpu,
  MemoryStick,
  Lock,
  AlertTriangle,
  Database,
  Target,
} from 'lucide-react'
import { TelemetryType, type Telemetry } from '@/types/telemetry'

// Helper function to get the profile type from schema key
const getProfileType = (schemaKey: string): string => {
  const lowerKey = schemaKey.toLowerCase()
  
  // Check for common profile types in the schema key
  if (lowerKey.includes('cpu')) return 'cpu'
  if (lowerKey.includes('memory') || lowerKey.includes('heap')) return 'memory'
  if (lowerKey.includes('alloc')) return 'allocs'
  if (lowerKey.includes('lock')) return 'locks'
  if (lowerKey.includes('exception') || lowerKey.includes('error')) return 'exceptions'
  
  return 'profile' // default profile type
}

// Helper function to get the correct data type based on telemetry type
export const getDataType = (telemetry: Telemetry): string => {
  switch (telemetry.telemetryType) {
    case TelemetryType.Metric:
      return telemetry.metricType || ''
    case TelemetryType.Span:
      return telemetry.spanKind || ''
    case TelemetryType.Log:
      return 'log'
    case TelemetryType.Profile:
      return getProfileType(telemetry.schemaKey)
    default:
      return ''
  }
}

export const DataTypeIcon = ({ dataType }: { dataType: string }) => {
  if (!dataType) {
    return <Code className="h-4 w-4 text-gray-400" />
  }

  switch (dataType.toLowerCase()) {
    case 'gauge':
      return <BarChart className="h-4 w-4 text-blue-400" />
    case 'counter':
      return <Hash className="h-4 w-4 text-blue-400" />
    case 'histogram':
      return <BarChart3 className="h-4 w-4 text-blue-400" />
    case 'summary':
      return <BarChart3 className="h-4 w-4 text-blue-400" />
    case 'exponentialhistogram':
      return <BarChart3 className="h-4 w-4 text-blue-400" />
    case 'structured':
      return <Layers className="h-4 w-4 text-green-400" />
    case 'unstructured':
      return <FileText className="h-4 w-4 text-green-400" />
    case 'span':
      return <Activity className="h-4 w-4 text-purple-400" />
    // Span kinds for traces
    case 'server':
      return <Server className="h-4 w-4 text-purple-400" />
    case 'client':
      return <Smartphone className="h-4 w-4 text-purple-400" />
    case 'producer':
      return <ArrowUp className="h-4 w-4 text-purple-400" />
    case 'consumer':
      return <ArrowDown className="h-4 w-4 text-purple-400" />
    case 'internal':
      return <Zap className="h-4 w-4 text-purple-400" />
    case 'log':
      return <Logs className="h-4 w-4 text-green-400" />
    // Profile types
    case 'cpu':
      return <Cpu className="h-4 w-4 text-orange-400" />
    case 'memory':
      return <MemoryStick className="h-4 w-4 text-orange-400" />
    case 'locks':
      return <Lock className="h-4 w-4 text-orange-400" />
    case 'exceptions':
      return <AlertTriangle className="h-4 w-4 text-orange-400" />
    case 'allocs':
      return <Database className="h-4 w-4 text-orange-400" />
    case 'profile':
      return <Target className="h-4 w-4 text-orange-400" />
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
    case TelemetryType.Span:
      return <Activity className="h-5 w-5 text-purple-500" />
    case TelemetryType.Profile:
      return <Target className="h-5 w-5 text-orange-500" />
    default:
      return <Code className="h-5 w-5 text-gray-500" />
  }
}
