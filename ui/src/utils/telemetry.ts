import { Status, TelemetryType } from '@/types/telemetry'

export const getTelemetryTypeBgColor = (type: TelemetryType) => {
  switch (type) {
    case TelemetryType.Metric:
      return 'bg-blue-500/10'
    case TelemetryType.Log:
      return 'bg-green-500/10'
    case TelemetryType.Span:
      return 'bg-purple-500/10'
    default:
      return 'bg-gray-500/10'
  }
}

export const getStatusBadge = (status?: Status) => {
  if (!status) {
    return null
  }

  switch (status) {
    case Status.Active:
      return {
        className: 'bg-green-500/10 text-green-500 border-green-500/20',
        label: 'Active',
      }
    case Status.Experimental:
      return {
        className: 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20',
        label: 'Experimental',
      }
    case Status.Deprecated:
      return {
        className: 'bg-red-500/10 text-red-500 border-red-500/20',
        label: 'Deprecated',
      }
    case Status.Stable:
      return {
        className: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
        label: 'Stable',
      }
    default:
      return null
  }
}
