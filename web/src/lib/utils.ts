import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return '0 Bytes'

  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']

  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i]
}

export function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h`
  return `${Math.floor(seconds / 86400)}d`
}

export function formatDate(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

export function formatDateTime(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text
  return text.slice(0, maxLength) + '...'
}

export function getStatusColor(status: string): string {
  switch (status) {
    case 'completed':
      return 'text-green-500'
    case 'downloading':
      return 'text-blue-500'
    case 'pending':
      return 'text-yellow-500'
    case 'failed':
      return 'text-red-500'
    case 'cancelled':
      return 'text-gray-500'
    default:
      return 'text-gray-500'
  }
}

export function getStatusBgColor(status: string): string {
  switch (status) {
    case 'completed':
      return 'bg-green-500/10 text-green-500'
    case 'downloading':
      return 'bg-blue-500/10 text-blue-500'
    case 'pending':
      return 'bg-yellow-500/10 text-yellow-500'
    case 'failed':
      return 'bg-red-500/10 text-red-500'
    case 'cancelled':
      return 'bg-gray-500/10 text-gray-500'
    default:
      return 'bg-gray-500/10 text-gray-500'
  }
}
