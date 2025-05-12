import type { TelemetryType, Status } from '@/types/telemetry'

export const getTelemetryTypeBgColor = (type: TelemetryType) => {
  switch (type) {
    case "metric":
      return "bg-blue-500/10"
    case "log":
      return "bg-green-500/10"
    case "trace":
      return "bg-purple-500/10"
    default:
      return "bg-gray-500/10"
  }
}

export const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  return new Intl.DateTimeFormat("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  }).format(date)
}

export const getStatusBadge = (status: Status) => {
  switch (status.toLowerCase()) {
    case "active":
      return {
        className: "bg-green-500/10 text-green-500 border-green-500/20",
        label: "Active"
      }
    case "draft":
      return {
        className: "bg-yellow-500/10 text-yellow-500 border-yellow-500/20",
        label: "Draft"
      }
    case "deprecated":
      return {
        className: "bg-red-500/10 text-red-500 border-red-500/20",
        label: "Deprecated"
      }
    default:
      return null
  }
} 