import { describe, it, expect, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ThemeProvider } from './ThemeProvider'
import { useAppStore } from '@/store/appStore'

describe('ThemeProvider', () => {
  beforeEach(() => {
    // Reset theme to dark
    useAppStore.getState().setTheme('dark')
    // Clear document classes
    document.documentElement.classList.remove('dark')
  })

  it('should render children', () => {
    render(
      <ThemeProvider>
        <div data-testid="child">Test Content</div>
      </ThemeProvider>
    )
    
    expect(screen.getByTestId('child')).toBeInTheDocument()
  })

  it('should add dark class when theme is dark', () => {
    useAppStore.getState().setTheme('dark')
    
    render(
      <ThemeProvider>
        <div>Content</div>
      </ThemeProvider>
    )
    
    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })

  it('should remove dark class when theme is light', () => {
    useAppStore.getState().setTheme('dark')
    
    const { rerender } = render(
      <ThemeProvider>
        <div>Content</div>
      </ThemeProvider>
    )
    
    expect(document.documentElement.classList.contains('dark')).toBe(true)
    
    useAppStore.getState().setTheme('light')
    
    rerender(
      <ThemeProvider>
        <div>Content</div>
      </ThemeProvider>
    )
    
    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })

  it('should respond to theme changes', () => {
    useAppStore.getState().setTheme('dark')
    
    render(
      <ThemeProvider>
        <div>Content</div>
      </ThemeProvider>
    )
    
    expect(document.documentElement.classList.contains('dark')).toBe(true)
    
    // Toggle theme
    useAppStore.getState().toggleTheme()
    
    // Re-render to trigger effect
    render(
      <ThemeProvider>
        <div>Content 2</div>
      </ThemeProvider>
    )
    
    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })
})
