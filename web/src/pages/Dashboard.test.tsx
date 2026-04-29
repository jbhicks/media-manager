import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Dashboard } from './Dashboard'
import * as useApi from '@/hooks/useApi'

vi.mock('@/hooks/useApi', () => ({
  useStats: vi.fn(),
}))

describe('Dashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render dashboard title', () => {
    vi.mocked(useApi.useStats).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as any)

    render(<Dashboard />)
    
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText('Welcome to your media manager. Monitor your downloads and library.')).toBeInTheDocument()
  })

  it('should show loading state for stats', () => {
    vi.mocked(useApi.useStats).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as any)

    render(<Dashboard />)
    
    const loadingIndicators = screen.getAllByText('...')
    expect(loadingIndicators.length).toBe(4)
  })

  it('should display stat values', () => {
    vi.mocked(useApi.useStats).mockReturnValue({
      data: {
        pending: 5,
        downloading: 2,
        completed: 10,
        failed: 1,
      },
      isLoading: false,
    } as any)

    render(<Dashboard />)
    
    expect(screen.getByText('5')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
    expect(screen.getByText('10')).toBeInTheDocument()
    expect(screen.getByText('1')).toBeInTheDocument()
  })

  it('should render stat card labels', () => {
    vi.mocked(useApi.useStats).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as any)

    render(<Dashboard />)
    
    expect(screen.getByText('Pending')).toBeInTheDocument()
    expect(screen.getByText('Downloading')).toBeInTheDocument()
    expect(screen.getByText('Completed')).toBeInTheDocument()
    expect(screen.getByText('Failed')).toBeInTheDocument()
  })

  it('should render quick action cards', () => {
    vi.mocked(useApi.useStats).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as any)

    render(<Dashboard />)
    
    expect(screen.getByText('Downloads')).toBeInTheDocument()
    expect(screen.getByText('Library')).toBeInTheDocument()
    expect(screen.getByText('Search')).toBeInTheDocument()
  })

  it('should render quick action descriptions', () => {
    vi.mocked(useApi.useStats).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as any)

    render(<Dashboard />)
    
    expect(screen.getByText('Manage your download queue')).toBeInTheDocument()
    expect(screen.getByText('Browse your media collection')).toBeInTheDocument()
    expect(screen.getByText('Find new content')).toBeInTheDocument()
  })

  it('should default to 0 when stats are undefined', () => {
    vi.mocked(useApi.useStats).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as any)

    render(<Dashboard />)
    
    expect(screen.getAllByText('0').length).toBe(4)
  })
})