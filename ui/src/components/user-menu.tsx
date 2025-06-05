import React, { useEffect, useRef, useState } from 'react'
import { Monitor, Moon, Sun } from 'lucide-react'
import { useTheme } from '@/components/theme-provider'

interface UserMenuProps {
  userInitials?: string
  userEmail?: string
  userName?: string
}

export const UserMenu = ({
  userInitials = 'TC',
  userName = 'TallyCat',
}: UserMenuProps) => {
  const { theme, setTheme } = useTheme()
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false)
  const avatarRef = useRef<HTMLButtonElement>(null)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        menuRef.current &&
        !menuRef.current.contains(event.target as Node) &&
        avatarRef.current &&
        !avatarRef.current.contains(event.target as Node)
      ) {
        setIsUserMenuOpen(false)
      }
    }

    if (isUserMenuOpen) {
      document.addEventListener('mousedown', handleClickOutside)
    }
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [isUserMenuOpen])

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setIsUserMenuOpen(false)
    }

    if (isUserMenuOpen) {
      document.addEventListener('keydown', handleKeyDown)
    }
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isUserMenuOpen])

  const handleToggle = () => setIsUserMenuOpen((open) => !open)
  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') handleToggle()
  }

  return (
    <div className="relative">
      <button
        ref={avatarRef}
        className="flex items-center gap-2 w-auto h-8 rounded-full bg-muted text-primary font-semibold border border-border hover:bg-muted focus:outline-none transition-colors px-2"
        aria-label="User menu"
        tabIndex={0}
        onClick={handleToggle}
        onKeyDown={handleKeyPress}
        aria-haspopup="menu"
        aria-expanded={isUserMenuOpen}
      >
        <span className="flex items-center justify-center w-8 h-8 rounded-full bg-muted text-primary font-semibold">
          {userInitials}
        </span>
        <span className="font-medium text-sm text-sidebar-foreground hidden sm:block">
          {userName}
        </span>
      </button>
      {isUserMenuOpen && (
        <div
          ref={menuRef}
          className="absolute right-0 mt-2 w-64 rounded-xl bg-sidebar text-sidebar-foreground shadow-lg border border-sidebar-border z-50 p-2 animate-in fade-in"
          role="menu"
          tabIndex={-1}
        >
          <div className="flex items-center gap-3 px-3 py-2">
            <div className="flex items-center justify-center w-10 h-10 rounded-full bg-muted text-foreground font-semibold">
              {userInitials}
            </div>
            <div>
              <div className="font-medium text-foreground">{userName}</div>
              {/* <div className="text-xs text-muted-foreground">{userEmail}</div> */}
            </div>
          </div>
          <div className="my-2 border-t border-border" />
          {/* <button
            className="w-full flex items-center gap-2 px-3 py-2 rounded-md hover:bg-muted text-sm text-foreground"
            role="menuitem"
            tabIndex={0}
          >
            <span>Account</span>
          </button> */}
          {/* <button
            className="w-full flex items-center gap-2 px-3 py-2 rounded-md hover:bg-muted text-sm text-foreground"
            role="menuitem"
            tabIndex={0}
          >
            <span>Billing</span>
          </button> */}
          {/* <button
            className="w-full flex items-center gap-2 px-3 py-2 rounded-md hover:bg-muted text-sm text-foreground"
            role="menuitem"
            tabIndex={0}
          >
            <svg
              width="16"
              height="16"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              viewBox="0 0 24 24"
              className="text-muted-foreground"
            >
              <path d="M18 16v-5a6 6 0 10-12 0v5a2 2 0 01-2 2h16a2 2 0 01-2-2z" />
              <path d="M13.73 21a2 2 0 01-3.46 0" />
            </svg>
            <span>Notifications</span>
          </button> */}
          {/* <div className="my-2 border-t border-border" /> */}
          <div className="w-full flex items-center gap-2 px-3 py-2">
            <span className="text-sm text-foreground">Theme</span>
            <button
              className={`ml-auto rounded-md p-1 ${theme === 'light' ? 'bg-muted text-primary' : 'hover:bg-muted text-muted-foreground'}`}
              aria-label="Light mode"
              onClick={() => setTheme('light')}
            >
              <Sun className="w-4 h-4 text-muted-foreground" />
            </button>
            <button
              className={`rounded-md p-1 ${theme === 'dark' ? 'bg-muted text-primary' : 'hover:bg-muted text-muted-foreground'}`}
              aria-label="Dark mode"
              onClick={() => setTheme('dark')}
            >
              <Moon className="w-4 h-4 text-muted-foreground" />
            </button>
            <button
              className={`rounded-md p-1 ${theme === 'system' ? 'bg-muted text-primary' : 'hover:bg-muted text-muted-foreground'}`}
              aria-label="System mode"
              onClick={() => setTheme('system')}
            >
              <Monitor className="w-4 h-4 text-muted-foreground" />
            </button>
          </div>
          {/* <button
            className="w-full flex items-center gap-2 px-3 py-2 rounded-md hover:bg-destructive/10 text-sm text-destructive"
            role="menuitem"
            tabIndex={0}
          >
            <span>Log out</span>
          </button> */}
        </div>
      )}
    </div>
  )
}
