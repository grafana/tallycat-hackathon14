import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'
import type { ClassValue } from 'clsx'

export function cn(...inputs: Array<ClassValue>) {
  return twMerge(clsx(inputs))
}

export enum DateFormat {
  short = 'short',
  long = 'long',
  time = 'time',
  datetime = 'datetime',
  shortDateTime = 'shortDateTime',
}

type DateTimeConfig = {
  [key in DateFormat]: Intl.DateTimeFormatOptions
}

const config: DateTimeConfig = {
  [DateFormat.long]: {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  },
  [DateFormat.time]: {
    hour: '2-digit',
    minute: '2-digit',
  },
  [DateFormat.datetime]: {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  },
  [DateFormat.short]: {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  },
  [DateFormat.shortDateTime]: {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  },
}

/**
 * Formats a date string into a localized date format
 * @param dateString - The date string to format
 * @param format - The format to use ('short' | 'long' | 'time' | 'datetime' | 'shortDateTime')
 * @returns Formatted date string
 *
 * @example
 * formatDate('2024-03-20') // "Mar 20, 2024"
 * formatDate('2024-03-20', DateFormat.long) // "March 20, 2024"
 * formatDate('2024-03-20', DateFormat.time) // "12:00 AM"
 * formatDate('2024-03-20', DateFormat.datetime) // "Mar 20, 2024, 12:00 AM"
 * formatDate('2024-03-20', DateFormat.shortDateTime) // "Mar 20, 12:00 AM"
 */
export const formatDate = (
  dateString: string,
  format: DateFormat = DateFormat.short,
) => {
  try {
    const date = new Date(dateString)
    if (isNaN(date.getTime())) {
      return 'Invalid date'
    }
    return new Intl.DateTimeFormat('en-US', config[format]).format(date)
  } catch (error) {
    console.error('Error formatting date:', error)
    return 'Invalid date'
  }
}

/**
 * Formats scope names for better readability in the UI
 * Trims GitHub hostname and repository owner from scope names
 * 
 * Examples:
 * - "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver" 
 *   -> "receiver/prometheusreceiver"
 * - "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
 *   -> "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp" (unchanged)
 */
export const formatScopeName = (scopeName: string): string => {
  if (!scopeName || scopeName === 'UNKNOWN') {
    return scopeName
  }

  // Handle GitHub repository paths
  const githubPattern = /^github\.com\/[^\/]+\/[^\/]+\/(.+)$/
  const githubMatch = scopeName.match(githubPattern)
  if (githubMatch) {
    return githubMatch[1]
  }

  // Return original name if no pattern matches
  return scopeName
}
