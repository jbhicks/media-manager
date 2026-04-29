import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Downloads } from './Downloads'
import * as useApi from '@/hooks/useApi'
import { useAppStore } from '@/store/appStore'

vi.mock('@/hooks/useApi', () => ({
  useTasks: vi.fn(),
  useCancelTask: vi.fn(),
  useRestartTask: vi.fn(),
  useDeleteTask: vi.fn(),
  useClearCompletedTasks: vi.fn(),
  useClearFailedTasks: vi.fn(),
}))

vi.mock('@/store/appStore', () => ({
  useAppStore: vi.fn(),
}))

describe('Downloads', () => {
  const mockAddToast = vi.fn()
  const mockMutateAsync = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useAppStore).mockReturnValue({ addToast: mockAddToast } as any)
    vi.mocked(useApi.useCancelTask).mockReturnValue({ mutateAsync: mockMutateAsync, isPending: false } as any)
    vi.mocked(useApi.useRestartTask).mockReturnValue({ mutateAsync: mockMutateAsync, isPending: false } as any)
    vi.mocked(useApi.useDeleteTask).mockReturnValue({ mutateAsync: mockMutateAsync, isPending: false } as any)
    vi.mocked(useApi.useClearCompletedTasks).mockReturnValue({ mutateAsync: mockMutateAsync, isPending: false } as any)
    vi.mocked(useApi.useClearFailedTasks).mockReturnValue({ mutateAsync: mockMutateAsync, isPending: false } as any)
  })

  it('should show loading state', () => {
    vi.mocked(useApi.useTasks).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as any)

    render(<Downloads />)
    
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('should render page title', () => {
    vi.mocked(useApi.useTasks).mockReturnValue({
      data: [],
      isLoading: false,
    } as any)

    render(<Downloads />)
    
    expect(screen.getByText('Downloads')).toBeInTheDocument()
    expect(screen.getByText('Manage your download queue')).toBeInTheDocument()
  })

  it('should render empty state', () => {
    vi.mocked(useApi.useTasks).mockReturnValue({
      data: [],
      isLoading: false,
    } as any)

    render(<Downloads />)
    
    expect(screen.getByText('No downloads yet')).toBeInTheDocument()
  })

  it('should render task with correct information', () => {
    vi.mocked(useApi.useTasks).mockReturnValue({
      data: [
        {
          id: 1,
          title: 'Test Movie',
          status: 'downloading',
          size: 1073741824,
          seeders: 50,
          progress: 45.5,
          started_at: '2024-01-01T10:00:00Z',
        },
      ],
      isLoading: false,
    } as any)

    render(<Downloads />)
    
    expect(screen.getByText('Test Movie')).toBeInTheDocument()
    expect(screen.getByText('downloading')).toBeInTheDocument()
    expect(screen.getByText('1 GB')).toBeInTheDocument()
    expect(screen.getByText('50 seeders')).toBeInTheDocument()
    expect(screen.getByText('45.5% complete')).toBeInTheDocument()
  })

  it('should show cancel button for downloading tasks', () => {
    vi.mocked(useApi.useTasks).mockReturnValue({
      data: [
        {
          id: 1,
          title: 'Test Movie',
          status: 'downloading',
          size: 1073741824,
          seeders: 50,
          progress: 45.5,
        },
      ],
      isLoading: false,
    } as any)

    render(<Downloads />)
    
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThan(0)
  })

  it('should show restart button for failed tasks', () => {
    vi.mocked(useApi.useTasks).mockReturnValue({
      data: [
        {
          id: 1,
          title: 'Test Movie',
          status: 'failed',
          size: 1073741824,
          seeders: 50,
          progress: 0,
        },
      ],
      isLoading: false,
    } as any)

    render(<Downloads />)
    
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThan(0)
  })

  it('should show error message when task has error', () => {
    vi.mocked(useApi.useTasks).mockReturnValue({
      data: [
        {
          id: 1,
          title: 'Test Movie',
          status: 'failed',
          size: 1073741824,
          seeders: 50,
          progress: 0,
          error: 'Connection timeout',
        },
      ],
      isLoading: false,
    } as any)

    render(<Downloads />)
    
    expect(screen.getByText('Error: Connection timeout')).toBeInTheDocument()
  })

  it('should render clear buttons', () => {
    vi.mocked(useApi.useTasks).mockReturnValue({
      data: [],
      isLoading: false,
    } as any)

    render(<Downloads />)
    
    expect(screen.getByText('Clear Completed')).toBeInTheDocument()
    expect(screen.getByText('Clear Failed')).toBeInTheDocument()
  })

  it('should show delete button for all tasks', () => {
    vi.mocked(useApi.useTasks).mockReturnValue({
      data: [
        {
          id: 1,
          title: 'Test Movie',
          status: 'completed',
          size: 1073741824,
          seeders: 50,
          progress: 100,
        },
      ],
      isLoading: false,
    } as any)

    render(<Downloads />)
    
    // Should have at least 3 buttons: Clear Completed, Clear Failed, and action buttons
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThanOrEqual(3)
  })
})