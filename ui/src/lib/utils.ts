import {  clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'
import type {ClassValue} from 'clsx';

export function cn(...inputs: Array<ClassValue>) {
  return twMerge(clsx(inputs))
}

export enum DateFormat {
  short = 'short',
  long = 'long',
  time = 'time',
  datetime = 'datetime',
}

type DateTimeConfig = {
  [key in DateFormat]: Intl.DateTimeFormatOptions
}


const config: DateTimeConfig = {
  [DateFormat.long]: {
    year: "numeric",
    month: "long",
    day: "numeric",
  },
  [DateFormat.time]: {
    hour: "2-digit",
    minute: "2-digit",
  },
  [DateFormat.datetime]: {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  },
  [DateFormat.short]: {
    year: "numeric",
    month: "short",
    day: "numeric",
  }
}


/**
 * Formats a date string into a localized date format
 * @param dateString - The date string to format
 * @param format - The format to use ('short' | 'long' | 'time' | 'datetime')
 * @returns Formatted date string
 * 
 * @example
 * formatDate('2024-03-20') // "Mar 20, 2024"
 * formatDate('2024-03-20', DateFormat.long) // "March 20, 2024"
 * formatDate('2024-03-20', DateFormat.time) // "12:00 AM"
 * formatDate('2024-03-20', DateFormat.datetime) // "Mar 20, 2024, 12:00 AM"
 */
export const formatDate = (dateString: string, format: DateFormat = DateFormat.short) => {
  return new Intl.DateTimeFormat("en-US", config[format]).format(new Date(dateString))
}
