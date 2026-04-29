import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { VPNStatus } from './VPNStatus'
import * as useApi from '@/hooks/useApi'

// Mock the useVPNStatus hook
vi.mock('@/hooks/useApi', () => ({
  useVPNStatus: vi.fn(),
}))

describe('VPNStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should show loading state', () => {
    vi.mocked(useApi.useVPNStatus).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as any)
    
    render(<VPNStatus />)
    
    expect(screen.getByText('Checking VPN...')).toBeInTheDocument()
  })

  it('should show connected state', () => {
    vi.mocked(useApi.useVPNStatus).mockReturnValue({
      data: {
        active: true,
        status: 'connected',
        message: 'VPN active - United States (NordVPN)',
        provider: 'NordVPN',
        location: 'United States',
      },
      isLoading: false,
    } as any)
    
    render(<VPNStatus />)
    
    expect(screen.getByText('VPN Connected')).toBeInTheDocument()
  })

  it('should show disconnected state', () => {
    vi.mocked(useApi.useVPNStatus).mockReturnValue({
      data: {
        active: false,
        status: 'disconnected',
        message: 'VPN is not active',
      },
      isLoading: false,
    } as any)
    
    render(<VPNStatus />)
    
    expect(screen.getByText('VPN Disconnected')).toBeInTheDocument()
  })

  it('should show location when connected', () => {
    vi.mocked(useApi.useVPNStatus).mockReturnValue({
      data: {
        active: true,
        status: 'connected',
        location: 'Germany',
        country: 'DE',
      },
      isLoading: false,
    } as any)
    
    render(<VPNStatus />)
    
    expect(screen.getByText((content) => content.includes('Germany'))).toBeInTheDocument()
    expect(screen.getByText((content) => content.includes('DE'))).toBeInTheDocument()
  })

  it('should show provider when connected', () => {
    vi.mocked(useApi.useVPNStatus).mockReturnValue({
      data: {
        active: true,
        status: 'connected',
        provider: 'ExpressVPN',
      },
      isLoading: false,
    } as any)
    
    render(<VPNStatus />)
    
    expect(screen.getByText('ExpressVPN')).toBeInTheDocument()
  })

  it('should show abbreviated text on mobile', () => {
    vi.mocked(useApi.useVPNStatus).mockReturnValue({
      data: {
        active: true,
        status: 'connected',
      },
      isLoading: false,
    } as any)
    
    render(<VPNStatus />)
    
    // Check that both desktop and mobile text exist
    expect(screen.getByText('VPN')).toBeInTheDocument()
  })

  it('should apply green styling when connected', () => {
    vi.mocked(useApi.useVPNStatus).mockReturnValue({
      data: {
        active: true,
        status: 'connected',
      },
      isLoading: false,
    } as any)
    
    render(<VPNStatus />)
    
    const container = screen.getByText('VPN Connected').parentElement
    expect(container).toHaveClass('bg-green-500/10')
    expect(container).toHaveClass('text-green-500')
  })

  it('should apply red styling when disconnected', () => {
    vi.mocked(useApi.useVPNStatus).mockReturnValue({
      data: {
        active: false,
        status: 'disconnected',
      },
      isLoading: false,
    } as any)
    
    render(<VPNStatus />)
    
    const container = screen.getByText('VPN Disconnected').parentElement
    expect(container).toHaveClass('bg-red-500/10')
    expect(container).toHaveClass('text-red-500')
  })

  it('should show IP when no location or provider', () => {
    vi.mocked(useApi.useVPNStatus).mockReturnValue({
      data: {
        active: true,
        status: 'connected',
        ip: '192.168.1.1',
      },
      isLoading: false,
    } as any)
    
    render(<VPNStatus />)
    
    expect(screen.getByText('IP: 192.168.1.1')).toBeInTheDocument()
  })

  it('should have title attribute for tooltip', () => {
    vi.mocked(useApi.useVPNStatus).mockReturnValue({
      data: {
        active: true,
        status: 'connected',
        message: 'Connected to VPN',
      },
      isLoading: false,
    } as any)
    
    render(<VPNStatus />)
    
    const container = screen.getByText('VPN Connected').parentElement
    expect(container).toHaveAttribute('title', 'Connected to VPN')
  })
})
