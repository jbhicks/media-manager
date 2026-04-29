import { describe, it, expect } from 'vitest'
import {
  cn,
  formatBytes,
  formatDuration,
  formatDate,
  formatDateTime,
  truncateText,
  getStatusColor,
  getStatusBgColor,
} from './utils'

describe('cn (className utility)', () => {
  it('merges class names correctly', () => {
    expect(cn('foo', 'bar')).toBe('foo bar')
  })

  it('handles conditional classes', () => {
    const isActive = true
    const isDisabled = false
    expect(cn('base', isActive && 'active', isDisabled && 'disabled')).toBe('base active')
  })

  it('handles arrays of classes', () => {
    expect(cn(['foo', 'bar'], 'baz')).toBe('foo bar baz')
  })

  it('handles objects with Tailwind conflicts', () => {
    expect(cn('px-2', 'px-4')).toBe('px-4')
  })

  it('filters out falsy values', () => {
    expect(cn('foo', null, undefined, false, '', 'bar')).toBe('foo bar')
  })
})

describe('formatBytes', () => {
  it('returns 0 Bytes for 0', () => {
    expect(formatBytes(0)).toBe('0 Bytes')
  })

  it('formats bytes correctly', () => {
    expect(formatBytes(1024)).toBe('1 KB')
    expect(formatBytes(1024 * 1024)).toBe('1 MB')
    expect(formatBytes(1024 * 1024 * 1024)).toBe('1 GB')
    expect(formatBytes(1024 * 1024 * 1024 * 1024)).toBe('1 TB')
  })

  it('formats with custom decimals', () => {
    expect(formatBytes(1536, 0)).toBe('2 KB')
    expect(formatBytes(1536, 2)).toBe('1.5 KB')
  })

  it('handles negative decimals', () => {
    expect(formatBytes(1536, -1)).toBe('2 KB')
  })

  it('handles large values', () => {
    expect(formatBytes(1024 ** 5)).toBe('1 PB')
    expect(formatBytes(1024 ** 6)).toBe('1 EB')
    expect(formatBytes(1024 ** 7)).toBe('1 ZB')
    expect(formatBytes(1024 ** 8)).toBe('1 YB')
  })
})

describe('formatDuration', () => {
  it('formats seconds', () => {
    expect(formatDuration(45)).toBe('45s')
  })

  it('formats minutes', () => {
    expect(formatDuration(120)).toBe('2m')
    expect(formatDuration(59)).toBe('59s')
    expect(formatDuration(60)).toBe('1m')
  })

  it('formats hours', () => {
    expect(formatDuration(3600)).toBe('1h')
    expect(formatDuration(7200)).toBe('2h')
    expect(formatDuration(3599)).toBe('59m')
  })

  it('formats days', () => {
    expect(formatDuration(86400)).toBe('1d')
    expect(formatDuration(172800)).toBe('2d')
    expect(formatDuration(86399)).toBe('23h')
  })
})

describe('formatDate', () => {
  it('formats date string correctly', () => {
    const result = formatDate('2024-03-15')
    expect(result).toContain('2024')
    expect(result).toContain('Mar')
    // Day might be 14 or 15 depending on timezone, just check it's a valid date format
    expect(result).toMatch(/Mar \d{1,2}, 2024/)
  })

  it('handles ISO date strings', () => {
    const result = formatDate('2024-12-25T10:30:00Z')
    expect(result).toContain('2024')
    expect(result).toContain('Dec')
  })
})

describe('formatDateTime', () => {
  it('formats date and time', () => {
    const result = formatDateTime('2024-03-15T14:30:00')
    expect(result).toContain('2024')
    expect(result).toContain('Mar')
    expect(result).toContain('15')
    expect(result).toContain(':')
  })
})

describe('truncateText', () => {
  it('returns original text if shorter than max', () => {
    expect(truncateText('hello', 10)).toBe('hello')
  })

  it('truncates text and adds ellipsis', () => {
    expect(truncateText('hello world', 5)).toBe('hello...')
  })

  it('handles exact length', () => {
    expect(truncateText('hello', 5)).toBe('hello')
  })

  it('handles empty string', () => {
    expect(truncateText('', 5)).toBe('')
  })
})

describe('getStatusColor', () => {
  it('returns correct colors for each status', () => {
    expect(getStatusColor('completed')).toBe('text-green-500')
    expect(getStatusColor('downloading')).toBe('text-blue-500')
    expect(getStatusColor('pending')).toBe('text-yellow-500')
    expect(getStatusColor('failed')).toBe('text-red-500')
    expect(getStatusColor('cancelled')).toBe('text-gray-500')
  })

  it('returns default for unknown status', () => {
    expect(getStatusColor('unknown')).toBe('text-gray-500')
    expect(getStatusColor('')).toBe('text-gray-500')
  })
})

describe('getStatusBgColor', () => {
  it('returns correct background colors for each status', () => {
    expect(getStatusBgColor('completed')).toBe('bg-green-500/10 text-green-500')
    expect(getStatusBgColor('downloading')).toBe('bg-blue-500/10 text-blue-500')
    expect(getStatusBgColor('pending')).toBe('bg-yellow-500/10 text-yellow-500')
    expect(getStatusBgColor('failed')).toBe('bg-red-500/10 text-red-500')
    expect(getStatusBgColor('cancelled')).toBe('bg-gray-500/10 text-gray-500')
  })

  it('returns default for unknown status', () => {
    expect(getStatusBgColor('unknown')).toBe('bg-gray-500/10 text-gray-500')
    expect(getStatusBgColor('')).toBe('bg-gray-500/10 text-gray-500')
  })
})
