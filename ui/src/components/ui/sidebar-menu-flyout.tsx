import React, { useRef, useState, useLayoutEffect } from 'react'
import { createPortal } from 'react-dom'
import { useSidebar } from './sidebar'

interface SidebarMenuFlyoutProps {
  trigger: React.ReactNode
  children: React.ReactNode
  icon?: React.ReactNode
  label?: string
}

export const SidebarMenuFlyout = ({
  trigger,
  children,
  icon,
  label,
}: SidebarMenuFlyoutProps) => {
  const { state } = useSidebar()
  const [open, setOpen] = useState(false)
  const triggerRef = useRef<HTMLDivElement>(null)
  const [flyoutStyle, setFlyoutStyle] = useState<React.CSSProperties>({})

  // Only show flyout when sidebar is collapsed
  if (state !== 'collapsed') {
    return (
      <>
        {trigger}
        {open && children}
      </>
    )
  }

  const handleOpen = () => setOpen(true)
  const handleClose = () => setOpen(false)

  useLayoutEffect(() => {
    if (open && triggerRef.current) {
      const rect = triggerRef.current.getBoundingClientRect()
      setFlyoutStyle({
        position: 'fixed',
        top: rect.top,
        left: rect.right + 8, // 8px gap
        zIndex: 9999,
        minWidth: '15rem',
      })
    }
  }, [open])

  return (
    <div
      className="relative"
      onMouseEnter={handleOpen}
      onMouseLeave={handleClose}
      onFocus={handleOpen}
      onBlur={handleClose}
      tabIndex={0}
      aria-haspopup="menu"
      aria-expanded={open}
      ref={triggerRef}
    >
      {trigger}
      {open &&
        typeof window !== 'undefined' &&
        createPortal(
          <div
            style={flyoutStyle}
            className="rounded-2xl bg-sidebar text-sidebar-foreground shadow-xl border border-border px-4 py-3 animate-in fade-in"
            role="menu"
            tabIndex={-1}
          >
            {(icon || label) && (
              <div className="flex items-center gap-2 mb-2 pb-2 border-b border-border">
                {icon && (
                  <span className="w-5 h-5 text-xl text-muted-foreground flex items-center">
                    {icon}
                  </span>
                )}
                {label && (
                  <span className="font-medium text-sm text-muted-foreground">
                    {label}
                  </span>
                )}
              </div>
            )}
            <div className="flex flex-col gap-0.5">{children}</div>
          </div>,
          document.body,
        )}
    </div>
  )
}
