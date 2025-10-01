import * as React from 'react'
import {
  GalleryVerticalEnd,
  SquareTerminal,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react'

import { NavMain } from '@/components/nav-main'

import { TeamSwitcher } from '@/components/team-switcher'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
  useSidebar,
} from '@/components/ui/sidebar'

// This is sample data.
const data = {
  user: {
    name: 'shadcn',
    email: 'm@example.com',
    avatar: '/avatars/shadcn.jpg',
  },
  teams: [
    {
      name: 'Acme Inc',
      logo: GalleryVerticalEnd,
      plan: 'Enterprise',
    },
  ],
  navMain: [
    {
      title: 'Data Governance',
      url: '#',
      icon: SquareTerminal,
      isActive: true,
      items: [
        {
          title: 'Telemetry Catalog',
          url: '/data-governance/telemetry-catalog',
        },
        {
          title: 'Entity Catalog',
          url: '/data-governance/entity-catalog',
        },
        {
          title: 'Scope Catalog',
          url: '/data-governance/scope-catalog',
        },
      ],
    },
  ],
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { state, toggleSidebar } = useSidebar()
  return (
    <Sidebar collapsible="icon" {...props} className="relative h-full">
      {/* Floating sidebar toggle button */}
      <button
        type="button"
        onClick={toggleSidebar}
        className="absolute -right-4 top-1/2 -translate-y-1/2 z-50 w-7 h-7 rounded-full bg-background shadow-lg border border-border flex items-center justify-center transition-colors hover:bg-muted focus:outline-none"
        aria-label="Toggle sidebar"
        tabIndex={0}
      >
        {state === 'expanded' ? (
          <ChevronLeft className="w-4 h-4" />
        ) : (
          <ChevronRight className="w-4 h-4" />
        )}
      </button>
      <SidebarHeader>
        <TeamSwitcher teams={data.teams} />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
      </SidebarContent>
      <SidebarFooter></SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
