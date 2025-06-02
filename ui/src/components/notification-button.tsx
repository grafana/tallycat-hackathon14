import React from 'react'

interface NotificationButtonProps {
  hasNotifications?: boolean
  onClick?: () => void
}

export const NotificationButton = ({
  hasNotifications = true,
  onClick,
}: NotificationButtonProps) => {
  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      onClick?.()
    }
  }

  return (
    <button
      className="relative rounded-full p-2 hover:bg-muted focus:outline-none transition-colors"
      aria-label="Notifications"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={handleKeyPress}
    >
      <svg
        width="20"
        height="20"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        viewBox="0 0 24 24"
        aria-hidden="true"
        className="text-foreground"
      >
        <path d="M18 16v-5a6 6 0 10-12 0v5a2 2 0 01-2 2h16a2 2 0 01-2-2z" />
        <path d="M13.73 21a2 2 0 01-3.46 0" />
      </svg>
      {hasNotifications && (
        <span className="absolute top-1 right-1 block h-2 w-2 rounded-full bg-primary" />
      )}
    </button>
  )
}
