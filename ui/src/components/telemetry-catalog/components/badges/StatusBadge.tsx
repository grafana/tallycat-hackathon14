import { CheckCircle2, AlertTriangle, XCircle } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Status } from '@/types/telemetry'

interface StatusBadgeProps {
  status: Status
}

export const StatusBadge = ({ status }: StatusBadgeProps) => {
  switch (status) {
    case Status.Active:
      return (
        <Badge
          variant="outline"
          className="bg-green-500/20 text-green-400 border-green-500/30 font-medium"
        >
          <CheckCircle2 className="h-3 w-3 mr-1" />
          Active
        </Badge>
      )
    case Status.Experimental:
      return (
        <Badge
          variant="outline"
          className="bg-yellow-500/20 text-yellow-400 border-yellow-500/30 font-medium"
        >
          <AlertTriangle className="h-3 w-3 mr-1" />
          Experimental
        </Badge>
      )
    case Status.Deprecated:
      return (
        <Badge
          variant="outline"
          className="bg-red-500/20 text-red-400 border-red-500/30 font-medium"
        >
          <XCircle className="h-3 w-3 mr-1" />
          Deprecated
        </Badge>
      )
    case Status.Stable:
      return (
        <Badge
          variant="outline"
          className="bg-blue-500/20 text-blue-400 border-blue-500/30 font-medium"
        >
          <CheckCircle2 className="h-3 w-3 mr-1" />
          Stable
        </Badge>
      )
    default:
      return null
  }
}
