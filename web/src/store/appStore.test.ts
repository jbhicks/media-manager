import { describe, it, expect, beforeEach } from 'vitest'
import { useAppStore } from './appStore'

describe('App Store', () => {
  beforeEach(() => {
    // Reset store to initial state before each test
    const store = useAppStore.getState()
    store.setTheme('dark')
    store.setSidebarOpen(true)
    store.clearSelection()
    store.setSearchQuery('')
    // Clear toasts
    useAppStore.setState({ toasts: [] })
  })

  describe('Theme', () => {
    it('should default to dark theme', () => {
      const state = useAppStore.getState()
      expect(state.theme).toBe('dark')
    })

    it('should toggle theme', () => {
      const store = useAppStore.getState()
      store.toggleTheme()
      expect(useAppStore.getState().theme).toBe('light')
      
      store.toggleTheme()
      expect(useAppStore.getState().theme).toBe('dark')
    })

    it('should set theme directly', () => {
      const store = useAppStore.getState()
      store.setTheme('light')
      expect(useAppStore.getState().theme).toBe('light')
      
      store.setTheme('dark')
      expect(useAppStore.getState().theme).toBe('dark')
    })
  })

  describe('Sidebar', () => {
    it('should default to open', () => {
      const state = useAppStore.getState()
      expect(state.sidebarOpen).toBe(true)
    })

    it('should toggle sidebar', () => {
      const store = useAppStore.getState()
      store.toggleSidebar()
      expect(useAppStore.getState().sidebarOpen).toBe(false)
      
      store.toggleSidebar()
      expect(useAppStore.getState().sidebarOpen).toBe(true)
    })

    it('should set sidebar state', () => {
      const store = useAppStore.getState()
      store.setSidebarOpen(false)
      expect(useAppStore.getState().sidebarOpen).toBe(false)
      
      store.setSidebarOpen(true)
      expect(useAppStore.getState().sidebarOpen).toBe(true)
    })
  })

  describe('Toasts', () => {
    it('should add a toast', () => {
      const store = useAppStore.getState()
      store.addToast('Test message', 'success')
      
      const toasts = useAppStore.getState().toasts
      expect(toasts).toHaveLength(1)
      expect(toasts[0].message).toBe('Test message')
      expect(toasts[0].type).toBe('success')
      expect(toasts[0].id).toBeDefined()
    })

    it('should default to info type', () => {
      const store = useAppStore.getState()
      store.addToast('Info message')
      
      const toasts = useAppStore.getState().toasts
      expect(toasts[0].type).toBe('info')
    })

    it('should remove a toast', () => {
      const store = useAppStore.getState()
      store.addToast('Test message')
      
      const toasts = useAppStore.getState().toasts
      const toastId = toasts[0].id
      
      store.removeToast(toastId)
      expect(useAppStore.getState().toasts).toHaveLength(0)
    })

    it('should auto-remove toast after duration', async () => {
      const store = useAppStore.getState()
      store.addToast('Auto remove', 'success')
      
      expect(useAppStore.getState().toasts).toHaveLength(1)
      
      // Wait for auto-remove (3s for success)
      await new Promise(resolve => setTimeout(resolve, 3100))
      
      expect(useAppStore.getState().toasts).toHaveLength(0)
    })

    it('should keep error toasts longer', async () => {
      const store = useAppStore.getState()
      store.addToast('Error message', 'error')
      
      // Wait 3 seconds - error toast should still be there (8s duration)
      await new Promise(resolve => setTimeout(resolve, 3100))
      
      expect(useAppStore.getState().toasts).toHaveLength(1)
      
      // Wait remaining time
      await new Promise(resolve => setTimeout(resolve, 5100))
      
      expect(useAppStore.getState().toasts).toHaveLength(0)
    }, 10000)
  })

  describe('Selection', () => {
    it('should toggle selection', () => {
      const store = useAppStore.getState()
      store.toggleSelection(1)
      
      expect(useAppStore.getState().selectedItems.has(1)).toBe(true)
      
      store.toggleSelection(1)
      expect(useAppStore.getState().selectedItems.has(1)).toBe(false)
    })

    it('should add multiple selections', () => {
      const store = useAppStore.getState()
      store.toggleSelection(1)
      store.toggleSelection(2)
      store.toggleSelection(3)
      
      const selected = useAppStore.getState().selectedItems
      expect(selected.has(1)).toBe(true)
      expect(selected.has(2)).toBe(true)
      expect(selected.has(3)).toBe(true)
      expect(selected.size).toBe(3)
    })

    it('should set selection from array', () => {
      const store = useAppStore.getState()
      store.setSelection([1, 2, 3])
      
      const selected = useAppStore.getState().selectedItems
      expect(selected.size).toBe(3)
      expect(selected.has(1)).toBe(true)
      expect(selected.has(2)).toBe(true)
      expect(selected.has(3)).toBe(true)
    })

    it('should clear selection', () => {
      const store = useAppStore.getState()
      store.setSelection([1, 2, 3])
      store.clearSelection()
      
      expect(useAppStore.getState().selectedItems.size).toBe(0)
    })
  })

  describe('Search', () => {
    it('should set search query', () => {
      const store = useAppStore.getState()
      store.setSearchQuery('test query')
      
      expect(useAppStore.getState().searchQuery).toBe('test query')
    })

    it('should clear search query', () => {
      const store = useAppStore.getState()
      store.setSearchQuery('test')
      store.setSearchQuery('')
      
      expect(useAppStore.getState().searchQuery).toBe('')
    })
  })
})
