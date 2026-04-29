import { create } from 'zustand'

interface Toast {
  id: string
  message: string
  type: 'success' | 'error' | 'info'
}

interface AppState {
  // Theme
  theme: 'light' | 'dark'
  toggleTheme: () => void
  setTheme: (theme: 'light' | 'dark') => void
  
  // Navigation
  sidebarOpen: boolean
  toggleSidebar: () => void
  setSidebarOpen: (open: boolean) => void
  
  // Toasts
  toasts: Toast[]
  addToast: (message: string, type?: 'success' | 'error' | 'info') => void
  removeToast: (id: string) => void
  
  // Selection
  selectedItems: Set<number>
  toggleSelection: (id: number) => void
  setSelection: (ids: number[]) => void
  clearSelection: () => void
  
  // Search
  searchQuery: string
  setSearchQuery: (query: string) => void
}

export const useAppStore = create<AppState>((set) => ({
  // Theme
  theme: 'dark',
  toggleTheme: () => set((state) => ({ theme: state.theme === 'light' ? 'dark' : 'light' })),
  setTheme: (theme) => set({ theme }),
  
  // Navigation
  sidebarOpen: true,
  toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
  setSidebarOpen: (open) => set({ sidebarOpen: open }),
  
  // Toasts
  toasts: [],
    addToast: (message, type = 'info') => {
      const id = Math.random().toString(36).substring(7)
      set((state) => ({
        toasts: [...state.toasts, { id, message, type }],
      }))
      // Auto-remove: 3s for success/info, 8s for errors (so user can read them)
      const duration = type === 'error' ? 8000 : 3000
      setTimeout(() => {
        set((state) => ({
          toasts: state.toasts.filter((t) => t.id !== id),
        }))
      }, duration)
    },
  removeToast: (id) =>
    set((state) => ({
      toasts: state.toasts.filter((t) => t.id !== id),
    })),
  
  // Selection
  selectedItems: new Set(),
  toggleSelection: (id) =>
    set((state) => {
      const newSet = new Set(state.selectedItems)
      if (newSet.has(id)) {
        newSet.delete(id)
      } else {
        newSet.add(id)
      }
      return { selectedItems: newSet }
    }),
  setSelection: (ids) => set({ selectedItems: new Set(ids) }),
  clearSelection: () => set({ selectedItems: new Set() }),
  
  // Search
  searchQuery: '',
  setSearchQuery: (query) => set({ searchQuery: query }),
}))