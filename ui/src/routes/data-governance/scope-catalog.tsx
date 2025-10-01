import { createFileRoute } from '@tanstack/react-router'
import { ScopeCatalogTable } from '@/components/scope-catalog'

export const Route = createFileRoute('/data-governance/scope-catalog')({
  component: ScopeCatalogPage,
})

function ScopeCatalogPage() {
  return (
    <div className="flex-1 space-y-4 p-4 md:p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Scope Catalog</h2>
          <p className="text-muted-foreground">
            Explore OpenTelemetry instrumentation scopes and their associated telemetry data.
            Each scope represents an instrumentation library or component that generates telemetry.
          </p>
        </div>
      </div>
      <ScopeCatalogTable />
    </div>
  )
}
