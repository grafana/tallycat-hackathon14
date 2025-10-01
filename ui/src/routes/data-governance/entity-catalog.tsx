import { createFileRoute } from '@tanstack/react-router'
import { EntityCatalogTable } from '@/components/entity-catalog'

export const EntityCatalog = () => {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-medium">Entity Catalog</h1>
        <p className="text-muted-foreground">
          Browse and explore entity types and their associated telemetry signals
        </p>
      </div>

      <div className="flex flex-col gap-6">
        <EntityCatalogTable />
      </div>
    </div>
  )
}

export const Route = createFileRoute('/data-governance/entity-catalog')({
  component: EntityCatalog,
})
