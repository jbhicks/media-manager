import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { VPNWarning } from './VPNWarning'
import * as useApi from '@/hooks/useApi'

// Mock the useVPNStatus hook
vi.mock('@/hooks/useApi', () => ({
  useVPNStatus: vi.fn(),
}))

describe('VPNWarning', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should not render when VPN is connected', () => {
    vi.mocked(useApi.useVPNStatus).mockReturnValue({
      data: { active: true },
    } as any)
    
    render(<VPNWarning />)
    
    expect(screen.queryByText('VPN Disconnected')).not.toBeInTheDocument()
  })

  it('should render when VPN is disconnected', () => {
    vi.mocked(useApi.useVPNStatus).mockReturnValue({
      data: { active: false },
    } as any)
    
    render(<VPNWarning />)
    
    expect(screen.getByText('VPN Disconnected')).toBeInTheDocument()
    expect(screen.getByText(/Downloads are disabled for your security/)).toBeInTheDocument()
  })

  it('should not render when dismissed', () => {
    vi.mocked(useApi.useVPNStatus).mockReturnValue({
      data: { active: false },
    } as any)
    
    render(<VPNWarning />)
    
    expect(screen.getByText('VPN Disconnected')).toBeInTheDocument()
    
    // Click dismiss button
    const dismissButton = screen.getByLabelText('Dismiss warning')
    fireEvent.click(dismissButton)
    
    expect(screen.queryByText('VPN Disconnected')).not.toBeInTheDocument()
  })

  it('should have red styling', () => {
    vi.mocked(useApi.useVPNStatus).mockReturnValue({
      data: { active: false },
    } as any)
    
    render(<VPNWarning />)
    
    const banner = screen.getByText('VPN Disconnected').closest('div')?.parentElement
    expect(banner).toHaveClass('bg-red-500/10')
    expect(banner).toHaveClass('border-red-500/20')
    expect(banner).toHaveClass('text-red-500')
  })

  it('should have shield alert icon', () => {
    vi.mocked(useApi.useVPNStatus).mockReturnValue({
      data: { active: false },
    } as any)
    
    render(<VPNWarning />)
    
    // Check for the icon (it renders as svg)
    const icon = screen.getByText('VPN Disconnected').parentElement?.previousElementSibling
    expect(icon?.tagName.toLowerCase()).toBe('svg')
  })

  it('should not render when VPN status is loading', () => {
    vi.mocked(useApi.useVPNStatus).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as any)
    
    render(<VPNWarning />)
    
    expect(screen.queryByText('VPN Disconnected')).not.toBeInTheDocument()
  })
})
