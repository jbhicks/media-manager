import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { useAppStore } from '@/store/appStore'

// Mock the app store
vi.mock('@/store/appStore', () => ({
  useAppStore: vi.fn(),
}))

describe('Sidebar', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render navigation items when expanded', () => {
    vi.mocked(useAppStore).mockReturnValue({ sidebarOpen: true } as any)
    
    render(
      <MemoryRouter>
        <Sidebar />
      </MemoryRouter>
    )
    
    expect(screen.getByText('Home')).toBeInTheDocument()
    expect(screen.getByText('Downloads')).toBeInTheDocument()
    expect(screen.getByText('Library')).toBeInTheDocument()
    expect(screen.getByText('Search')).toBeInTheDocument()
    expect(screen.getByText('Suggestions')).toBeInTheDocument()
    expect(screen.getByText('Settings')).toBeInTheDocument()
  })

  it('should render only icons when collapsed', () => {
    vi.mocked(useAppStore).mockReturnValue({ sidebarOpen: false } as any)
    
    render(
      <MemoryRouter>
        <Sidebar />
      </MemoryRouter>
    )
    
    // Should not show text labels
    expect(screen.queryByText('Home')).not.toBeInTheDocument()
    expect(screen.queryByText('Downloads')).not.toBeInTheDocument()
  })

  it('should have correct width when expanded', () => {
    vi.mocked(useAppStore).mockReturnValue({ sidebarOpen: true } as any)
    
    render(
      <MemoryRouter>
        <Sidebar />
      </MemoryRouter>
    )
    
    const nav = screen.getByRole('navigation')
    expect(nav).toHaveClass('w-64')
  })

  it('should have correct width when collapsed', () => {
    vi.mocked(useAppStore).mockReturnValue({ sidebarOpen: false } as any)
    
    render(
      <MemoryRouter>
        <Sidebar />
      </MemoryRouter>
    )
    
    const nav = screen.getByRole('navigation')
    expect(nav).toHaveClass('w-16')
  })

  it('should highlight active link', () => {
    vi.mocked(useAppStore).mockReturnValue({ sidebarOpen: true } as any)
    
    render(
      <MemoryRouter initialEntries={['/downloads']}>
        <Sidebar />
      </MemoryRouter>
    )
    
    const activeLink = screen.getByText('Downloads').closest('a')
    expect(activeLink).toHaveClass('bg-accent')
    expect(activeLink).toHaveClass('text-accent-foreground')
  })

  it('should have correct navigation structure', () => {
    vi.mocked(useAppStore).mockReturnValue({ sidebarOpen: true } as any)
    
    render(
      <MemoryRouter>
        <Sidebar />
      </MemoryRouter>
    )
    
    const homeLink = screen.getByText('Home').closest('a')
    expect(homeLink).toHaveAttribute('href', '/')
    
    const downloadsLink = screen.getByText('Downloads').closest('a')
    expect(downloadsLink).toHaveAttribute('href', '/downloads')
    
    const libraryLink = screen.getByText('Library').closest('a')
    expect(libraryLink).toHaveAttribute('href', '/library')
  })

  it('should have border on right side', () => {
    vi.mocked(useAppStore).mockReturnValue({ sidebarOpen: true } as any)
    
    render(
      <MemoryRouter>
        <Sidebar />
      </MemoryRouter>
    )
    
    const nav = screen.getByRole('navigation')
    expect(nav).toHaveClass('border-r')
  })
})
