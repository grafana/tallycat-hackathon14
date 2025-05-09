import React from 'react';
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbPage, BreadcrumbSeparator } from '@/components/ui/breadcrumb';
import { NotificationButton } from '@/components/notification-button';
import { UserMenu } from '@/components/user-menu';

interface AppHeaderProps {
  breadcrumbs?: {
    items: Array<{
      label: string;
      href?: string;
      isCurrentPage?: boolean;
    }>;
  };
}

export const AppHeader = ({ 
  breadcrumbs = {
    items: [
      { label: 'Building Your Application', href: '#' },
      { label: 'Data Fetching', isCurrentPage: true }
    ]
  }
}: AppHeaderProps) => {
  return (
    <header className="flex h-16 shrink-0 items-center justify-between px-4 border-b border-border bg-background transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12">
      {/* Left section */}
      <div className="flex items-center gap-2">
        <Breadcrumb>
          <BreadcrumbList>
            {breadcrumbs.items.map((item, index) => (
              <React.Fragment key={item.label}>
                {index > 0 && <BreadcrumbSeparator className="hidden md:block" />}
                <BreadcrumbItem className={index === 0 ? 'hidden md:block' : ''}>
                  {item.isCurrentPage ? (
                    <BreadcrumbPage>{item.label}</BreadcrumbPage>
                  ) : (
                    <BreadcrumbLink href={item.href}>{item.label}</BreadcrumbLink>
                  )}
                </BreadcrumbItem>
              </React.Fragment>
            ))}
          </BreadcrumbList>
        </Breadcrumb>
      </div>
      {/* Right section */}
      <div className="flex items-center gap-4">
        <NotificationButton />
        <UserMenu />
      </div>
    </header>
  );
}; 