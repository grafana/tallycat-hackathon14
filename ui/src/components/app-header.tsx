import React from 'react'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { UserMenu } from '@/components/user-menu'
import { useMatches } from '@tanstack/react-router'

interface AppHeaderProps {
  breadcrumbs?: {
    items: Array<{
      label: string
      href?: string
      isCurrentPage?: boolean
    }>
  }
}

export const AppHeader = ({ breadcrumbs: propBreadcrumbs }: AppHeaderProps) => {
  const matches = useMatches()

  // Transform route matches into breadcrumb items
  const dynamicBreadcrumbs = matches
    .filter((match) => match.routeId !== '__root__') // Skip root route
    .map((match, index, array) => {
      const isLast = index === array.length - 1
      const path = match.pathname
      const label = match.routeId.split('/').pop()?.replace(/-/g, ' ') || ''

      return {
        label: label.charAt(0).toUpperCase() + label.slice(1),
        href: isLast ? undefined : path,
        isCurrentPage: isLast,
      }
    })

  const breadcrumbs = propBreadcrumbs || { items: dynamicBreadcrumbs }

  return (
    <header className="flex h-16 shrink-0 items-center justify-between px-4 border-b border-sidebar-border bg-sidebar text-sidebar-foreground transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12">
      {/* Left section */}
      <div className="flex items-center gap-2">
        <Breadcrumb>
          <BreadcrumbList>
            {breadcrumbs.items.map((item, index) => (
              <React.Fragment key={item.label}>
                {index > 0 && (
                  <BreadcrumbSeparator className="hidden md:block" />
                )}
                <BreadcrumbItem
                  className={index === 0 ? 'hidden md:block' : ''}
                >
                  {item.isCurrentPage ? (
                    <BreadcrumbPage>{item.label}</BreadcrumbPage>
                  ) : (
                    <BreadcrumbLink href={item.href}>
                      {item.label}
                    </BreadcrumbLink>
                  )}
                </BreadcrumbItem>
              </React.Fragment>
            ))}
          </BreadcrumbList>
        </Breadcrumb>
      </div>
      {/* Right section */}
      <div className="flex items-center gap-4">
        {/* <NotificationPopover /> */}
        <UserMenu />
      </div>
    </header>
  )
}
