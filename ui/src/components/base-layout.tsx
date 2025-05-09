import React, { useEffect, useRef, useState } from 'react';
import { Monitor, Moon, Sun } from 'lucide-react';
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
import { AppSidebar } from '@/components/app-sidebar'
import { Separator } from '@/components/ui/separator'
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbPage, BreadcrumbSeparator } from '@/components/ui/breadcrumb'
import { useTheme } from '@/components/theme-provider';


export const BaseLayout = ({children}: {children: React.ReactNode}) => {
    const { theme, setTheme } = useTheme();
    return (
      <>
        <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          <header className="flex h-16 shrink-0 items-center justify-between px-4 border-b border-border bg-background transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12">
            {/* Left section */}
            <div className="flex items-center gap-2">
              <Breadcrumb>
                <BreadcrumbList>
                  <BreadcrumbItem className="hidden md:block">
                    <BreadcrumbLink href="#">
                      Building Your Application
                    </BreadcrumbLink>
                  </BreadcrumbItem>
                  <BreadcrumbSeparator className="hidden md:block" />
                  <BreadcrumbItem>
                    <BreadcrumbPage>Data Fetching</BreadcrumbPage>
                  </BreadcrumbItem>
                </BreadcrumbList>
              </Breadcrumb>
            </div>
            {/* Right section */}
            <div className="flex items-center gap-4">
              {/* Notification button */}
              <button
                className="relative rounded-full p-2 hover:bg-muted focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors"
                aria-label="Notifications"
                tabIndex={0}
                onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') e.currentTarget.click(); }}
              >
                <svg width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24" aria-hidden="true" className="text-foreground">
                  <path d="M18 16v-5a6 6 0 10-12 0v5a2 2 0 01-2 2h16a2 2 0 01-2-2z" />
                  <path d="M13.73 21a2 2 0 01-3.46 0" />
                </svg>
                <span className="absolute top-1 right-1 block h-2 w-2 rounded-full bg-primary" />
              </button>
              {/* User avatar with dropdown */}
              {(() => {
                const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);
                const avatarRef = useRef<HTMLButtonElement>(null);
                const menuRef = useRef<HTMLDivElement>(null);

                // Close menu on outside click
                useEffect(() => {
                  const handleClickOutside = (event: MouseEvent) => {
                    if (
                      menuRef.current &&
                      !menuRef.current.contains(event.target as Node) &&
                      avatarRef.current &&
                      !avatarRef.current.contains(event.target as Node)
                    ) {
                      setIsUserMenuOpen(false);
                    }
                  };
                  if (isUserMenuOpen) {
                    document.addEventListener('mousedown', handleClickOutside);
                  } else {
                    document.removeEventListener('mousedown', handleClickOutside);
                  }
                  return () => {
                    document.removeEventListener('mousedown', handleClickOutside);
                  };
                }, [isUserMenuOpen]);

                // Close menu on Escape
                useEffect(() => {
                  const handleKeyDown = (event: KeyboardEvent) => {
                    if (event.key === 'Escape') setIsUserMenuOpen(false);
                  };
                  if (isUserMenuOpen) {
                    document.addEventListener('keydown', handleKeyDown);
                  } else {
                    document.removeEventListener('keydown', handleKeyDown);
                  }
                  return () => {
                    document.removeEventListener('keydown', handleKeyDown);
                  };
                }, [isUserMenuOpen]);

                return (
                  <div className="relative">
                    <button
                      ref={avatarRef}
                      className="flex items-center justify-center w-8 h-8 rounded-full bg-muted text-foreground font-semibold focus:outline-none focus:ring-2 focus:ring-primary/50"
                      aria-label="User menu"
                      tabIndex={0}
                      onClick={() => setIsUserMenuOpen((open) => !open)}
                      onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') setIsUserMenuOpen((open) => !open); }}
                      aria-haspopup="menu"
                      aria-expanded={isUserMenuOpen}
                    >
                      CN
                    </button>
                    {isUserMenuOpen && (
                      <div
                        ref={menuRef}
                        className="absolute right-0 mt-2 w-64 rounded-xl bg-popover shadow-lg border border-border z-50 p-2 animate-in fade-in"
                        role="menu"
                        tabIndex={-1}
                      >
                        <div className="flex items-center gap-3 px-3 py-2">
                          <div className="flex items-center justify-center w-10 h-10 rounded-full bg-muted text-foreground font-semibold">CN</div>
                          <div>
                            <div className="font-medium text-foreground">shadcn</div>
                            <div className="text-xs text-muted-foreground">m@example.com</div>
                          </div>
                        </div>
                        <div className="my-2 border-t border-border" />
                        <button className="w-full flex items-center gap-2 px-3 py-2 rounded-md hover:bg-muted text-sm text-foreground" role="menuitem" tabIndex={0}>
                          <span className="font-medium">Upgrade to Pro</span>
                        </button>
                        <button className="w-full flex items-center gap-2 px-3 py-2 rounded-md hover:bg-muted text-sm text-foreground" role="menuitem" tabIndex={0}>
                          <span>Account</span>
                        </button>
                        <button className="w-full flex items-center gap-2 px-3 py-2 rounded-md hover:bg-muted text-sm text-foreground" role="menuitem" tabIndex={0}>
                          <span>Billing</span>
                        </button>
                        <button className="w-full flex items-center gap-2 px-3 py-2 rounded-md hover:bg-muted text-sm text-foreground" role="menuitem" tabIndex={0}>
                          <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24" className="text-muted-foreground"><path d="M18 16v-5a6 6 0 10-12 0v5a2 2 0 01-2 2h16a2 2 0 01-2-2z" /><path d="M13.73 21a2 2 0 01-3.46 0" /></svg>
                          <span>Notifications</span>
                        </button>
                        <div className="my-2 border-t border-border" />
                        {/* Inline theme mode toggle */}
                        <div className="w-full flex items-center gap-2 px-3 py-2">
                          <span className="text-sm text-foreground">Theme</span>
                          <button
                            className={`ml-auto rounded-md p-1 ${theme === 'light' ? 'bg-muted text-primary' : 'hover:bg-muted text-muted-foreground'}`}
                            aria-label="Light mode"
                            onClick={() => setTheme('light')}
                          >
                            <Sun className="w-4 h-4" />
                          </button>
                          <button
                            className={`rounded-md p-1 ${theme === 'dark' ? 'bg-muted text-primary' : 'hover:bg-muted text-muted-foreground'}`}
                            aria-label="Dark mode"
                            onClick={() => setTheme('dark')}
                          >
                            <Moon className="w-4 h-4" />
                          </button>
                          <button
                            className={`rounded-md p-1 ${theme === 'system' ? 'bg-muted text-primary' : 'hover:bg-muted text-muted-foreground'}`}
                            aria-label="System mode"
                            onClick={() => setTheme('system')}
                          >
                            <Monitor className="w-4 h-4" />
                          </button>
                        </div>
                        <button className="w-full flex items-center gap-2 px-3 py-2 rounded-md hover:bg-destructive/10 text-sm text-destructive" role="menuitem" tabIndex={0}>
                          <span>Log out</span>
                        </button>
                      </div>
                    )}
                  </div>
                );
              })()}
            </div>
          </header>
          <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
            {children}
            {/* <div className="grid auto-rows-min gap-4 md:grid-cols-3">
              <div className="aspect-video rounded-xl bg-muted/50" />
              <div className="aspect-video rounded-xl bg-muted/50" />
              <div className="aspect-video rounded-xl bg-muted/50" />
            </div>
            <div className="min-h-[100vh] flex-1 rounded-xl bg-muted/50 md:min-h-min" /> */}
          </div>
        </SidebarInset>
      </SidebarProvider>
      </>
    )
  }