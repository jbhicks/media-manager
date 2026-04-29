import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ErrorBoundary } from './ErrorBoundary'

// Component that throws an error
function ThrowError({ shouldThrow }: { shouldThrow: boolean }) {
  if (shouldThrow) {
    throw new Error('Test error message')
  }
  return <div>No error</div>
}

describe('ErrorBoundary', () => {
  beforeEach(() => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  it('should render children when there is no error', () => {
    render(
      <ErrorBoundary>
        <div>Test content</div>
      </ErrorBoundary>
    )
    
    expect(screen.getByText('Test content')).toBeInTheDocument()
  })

  it('should render error fallback when child throws', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    )
    
    expect(screen.getByText('Something went wrong')).toBeInTheDocument()
    expect(screen.getByText('Test error message')).toBeInTheDocument()
  })

  it('should have reload button', () => {
    const reloadMock = vi.fn()
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { reload: reloadMock },
    })

    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    )
    
    const reloadButton = screen.getByText('Reload Application')
    expect(reloadButton).toBeInTheDocument()
    
    fireEvent.click(reloadButton)
    expect(reloadMock).toHaveBeenCalled()
  })

  it('should show error icon', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    )
    
    // AlertTriangle icon renders as svg
    const svg = document.querySelector('svg')
    expect(svg).toBeInTheDocument()
  })
})