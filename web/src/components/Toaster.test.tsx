import { describe, it, expect, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Toaster } from './Toaster'
import { useAppStore } from '@/store/appStore'

describe('Toaster', () => {
  beforeEach(() => {
    // Clear all toasts before each test
    useAppStore.setState({ toasts: [] })
  })

  it('should not render when there are no toasts', () => {
    render(<Toaster />)
    
    const container = screen.queryByRole('status')
    expect(container).not.toBeInTheDocument()
  })

  it('should render a toast message', () => {
    useAppStore.getState().addToast('Test message', 'success')
    
    render(<Toaster />)
    
    expect(screen.getByText('Test message')).toBeInTheDocument()
  })

  it('should render multiple toasts', () => {
    const store = useAppStore.getState()
    store.addToast('First message', 'success')
    store.addToast('Second message', 'error')
    store.addToast('Third message', 'info')
    
    render(<Toaster />)
    
    expect(screen.getByText('First message')).toBeInTheDocument()
    expect(screen.getByText('Second message')).toBeInTheDocument()
    expect(screen.getByText('Third message')).toBeInTheDocument()
  })

  it('should render success toast with correct styling', () => {
    useAppStore.getState().addToast('Success!', 'success')
    
    render(<Toaster />)
    
    const toast = screen.getByText('Success!').closest('div')
    expect(toast).toHaveClass('border-green-500/20')
    expect(toast).toHaveClass('bg-green-500/10')
    expect(toast).toHaveClass('text-green-500')
  })

  it('should render error toast with correct styling', () => {
    useAppStore.getState().addToast('Error!', 'error')
    
    render(<Toaster />)
    
    const toast = screen.getByText('Error!').closest('div')
    expect(toast).toHaveClass('border-red-500/20')
    expect(toast).toHaveClass('bg-red-500/10')
    expect(toast).toHaveClass('text-red-500')
  })

  it('should render info toast with correct styling', () => {
    useAppStore.getState().addToast('Info!', 'info')
    
    render(<Toaster />)
    
    const toast = screen.getByText('Info!').closest('div')
    expect(toast).toHaveClass('border-blue-500/20')
    expect(toast).toHaveClass('bg-blue-500/10')
    expect(toast).toHaveClass('text-blue-500')
  })

  it('should remove toast when close button is clicked', () => {
    useAppStore.getState().addToast('Removable', 'info')
    
    render(<Toaster />)
    
    expect(screen.getByText('Removable')).toBeInTheDocument()
    
    const closeButton = screen.getByRole('button')
    fireEvent.click(closeButton)
    
    expect(screen.queryByText('Removable')).not.toBeInTheDocument()
  })

  it('should have fixed positioning', () => {
    useAppStore.getState().addToast('Position test', 'info')
    
    render(<Toaster />)
    
    const container = screen.getByText('Position test').parentElement?.parentElement
    expect(container).toHaveClass('fixed')
    expect(container).toHaveClass('bottom-4')
    expect(container).toHaveClass('right-4')
    expect(container).toHaveClass('z-50')
  })

  it('should stack toasts vertically', () => {
    useAppStore.getState().addToast('Toast 1', 'info')
    useAppStore.getState().addToast('Toast 2', 'info')
    
    render(<Toaster />)
    
    const container = screen.getByText('Toast 1').parentElement?.parentElement
    expect(container).toHaveClass('flex-col')
  })
})
