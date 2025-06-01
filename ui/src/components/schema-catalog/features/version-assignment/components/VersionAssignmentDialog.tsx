"use client"

import { useState, useCallback } from "react"
import {
  CheckCircle2,
  AlertTriangle,
  XCircle,
  Edit,
  Save,
  X,
  Tag,
  Info,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog"
import type { Schema } from "@/types/schema-catalog"
import { Status } from "@/types/telemetry"

interface VersionAssignmentDialogProps {
  isOpen: boolean
  onClose: () => void
  selectedSchema: string | null
  currentSchema: Schema | null
  onVersionChange: (schemaId: string, version: string) => void
}

// Semantic versioning validation
const validateSemanticVersion = (version: string): string | null => {
  const semverRegex = /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$/
  if (!version) return "Version is required"
  if (!semverRegex.test(version)) return "Invalid semantic version format"
  return null
}

export const VersionAssignmentDialog = ({
  isOpen,
  onClose,
  selectedSchema,
  currentSchema,
  onVersionChange,
}: VersionAssignmentDialogProps) => {
  const [version, setVersion] = useState("")
  const [error, setError] = useState<string | null>(null)

  const handleVersionChange = useCallback((value: string) => {
    setVersion(value)
    setError(null)
  }, [])

  const handleSubmit = useCallback(() => {
    if (!selectedSchema) return

    const validationError = validateSemanticVersion(version)
    if (validationError) {
      setError(validationError)
      return
    }

    onVersionChange(selectedSchema, version)
    onClose()
  }, [selectedSchema, version, onVersionChange, onClose])

  const handleClose = useCallback(() => {
    setVersion("")
    setError(null)
    onClose()
  }, [onClose])

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Tag className="h-5 w-5" />
            {currentSchema?.status === Status.Experimental ? "Assign Schema Version" : "Update Schema Version"}
          </DialogTitle>
          <DialogDescription>
            {currentSchema?.status === Status.Experimental
              ? `Assign a semantic version to schema ${selectedSchema}`
              : `Update the version for schema ${selectedSchema}`}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Current Version Info */}
          {currentSchema?.version && (
            <div className="p-3 bg-muted/50 rounded-lg border">
              <div className="flex items-center gap-2 text-sm">
                <Info className="h-4 w-4 text-muted-foreground" />
                <span className="text-muted-foreground">Current version:</span>
                <Badge variant="outline" className="font-mono text-xs">
                  v{currentSchema.version}
                </Badge>
              </div>
            </div>
          )}

          {/* Version Input */}
          <div className="space-y-2">
            <Label htmlFor="version" className="text-sm font-medium">
              Version *
            </Label>
            <Input
              id="version"
              placeholder="e.g., 1.2.3, 2.0.0-beta.1"
              value={version}
              onChange={(e) => handleVersionChange(e.target.value)}
              className={error ? "border-red-500 focus:border-red-500" : ""}
            />
            {error && (
              <div className="text-xs text-red-500 flex items-center gap-1">
                <XCircle className="h-3 w-3" />
                {error}
              </div>
            )}
          </div>

          {/* Change Description */}
          <div className="space-y-2">
            <Label htmlFor="description" className="text-sm font-medium">
              Change Description *
            </Label>
            <Textarea
              id="description"
              placeholder="Describe the changes or reason for this version assignment..."
              rows={3}
            />
          </div>

          {/* Action Buttons */}
          <div className="flex items-center gap-2 pt-2">
            <Button
              onClick={handleSubmit}
              disabled={!version || !!error}
              className="flex items-center gap-2"
            >
              <Save className="h-4 w-4" />
              {currentSchema?.status === Status.Experimental ? "Assign Version" : "Update Version"}
            </Button>
            <Button variant="outline" onClick={handleClose} className="flex items-center gap-2">
              <X className="h-4 w-4" />
              Cancel
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
} 