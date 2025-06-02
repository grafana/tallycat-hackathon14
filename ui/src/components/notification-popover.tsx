import { useState, useRef, useEffect } from 'react'
import { Bell } from 'lucide-react'

const notifications = [
  { id: 1, title: 'New message from Alice', time: '2m ago' },
  { id: 2, title: 'Server backup completed', time: '10m ago' },
  { id: 3, title: 'Payment received', time: '1h ago' },
]

export const NotificationPopover = () => {
  const [open, setOpen] = useState(false)
  const buttonRef = useRef<HTMLButtonElement>(null)
  const popoverRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const handleClick = (e: MouseEvent) => {
      if (
        popoverRef.current &&
        !popoverRef.current.contains(e.target as Node) &&
        buttonRef.current &&
        !buttonRef.current.contains(e.target as Node)
      ) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [open])

  return (
    <div className="relative">
      <button
        ref={buttonRef}
        className="relative w-8 h-8 flex items-center justify-center rounded-full p-2 border border-border hover:bg-muted focus:outline-none focus-visible:ring-2 focus-visible:ring-primary transition-colors"
        aria-label="Notifications"
        tabIndex={0}
        onClick={() => setOpen((v) => !v)}
        aria-haspopup="menu"
        aria-expanded={open}
      >
        <Bell className="w-5 h-5 text-primary" />
        <span className="absolute top-1 right-1 block h-2 w-2 rounded-full bg-primary" />
      </button>
      {open && (
        <div
          ref={popoverRef}
          className="absolute right-0 mt-2 w-80 rounded-xl bg-sidebar text-sidebar-foreground shadow-xl border border-sidebar-border z-50 p-2 animate-in fade-in"
          role="menu"
          tabIndex={-1}
        >
          <div className="font-semibold px-3 py-2 text-sm border-b border-border mb-2">
            Notifications
          </div>
          {notifications.length === 0 ? (
            <div className="px-3 py-6 text-center text-muted-foreground text-sm">
              No notifications
            </div>
          ) : (
            <ul className="flex flex-col gap-1 max-h-60 overflow-auto">
              {notifications.map((n) => (
                <li
                  key={n.id}
                  className="px-3 py-2 rounded-md hover:bg-sidebar-accent transition-colors cursor-pointer flex items-center justify-between"
                >
                  <span className="truncate text-sm">{n.title}</span>
                  <span className="ml-2 text-xs text-muted-foreground whitespace-nowrap">
                    {n.time}
                  </span>
                </li>
              ))}
            </ul>
          )}
          <div className="mt-2 border-t border-border pt-2 text-right">
            <a
              href="#"
              className="text-xs text-primary hover:underline font-medium"
            >
              View all
            </a>
          </div>
        </div>
      )}
    </div>
  )
}
