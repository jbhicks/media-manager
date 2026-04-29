import { useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAppStore } from '@/store/appStore'

interface ShortcutConfig {
  key: string
  ctrl?: boolean
  alt?: boolean
  shift?: boolean
  action: () => void
  description: string
}

export function useKeyboardShortcuts() {
  const navigate = useNavigate()
  const { toggleSidebar } = useAppStore()

  const shortcuts: ShortcutConfig[] = [
    {
      key: 'd',
      action: () => navigate('/'),
      description: 'Go to Dashboard',
    },
    {
      key: 's',
      action: () => navigate('/search'),
      description: 'Go to Search',
    },
    {
      key: 'l',
      action: () => navigate('/library'),
      description: 'Go to Library',
    },
    {
      key: 't',
      action: () => navigate('/downloads'),
      description: 'Go to Downloads',
    },
    {
      key: 'g',
      action: () => navigate('/suggestions'),
      description: 'Go to Suggestions',
    },
    {
      key: ',',
      action: () => navigate('/settings'),
      description: 'Go to Settings',
    },
    {
      key: 'b',
      action: () => toggleSidebar(),
      description: 'Toggle sidebar',
    },
  ]

  const handleKeyDown = useCallback(
    (event: KeyboardEvent) => {
      // Ignore shortcuts when typing in input fields
      if (
        event.target instanceof HTMLInputElement ||
        event.target instanceof HTMLTextAreaElement ||
        event.target instanceof HTMLSelectElement
      ) {
        return
      }

      // Ignore if modifier keys are pressed (except for specific combos)
      if (event.ctrlKey || event.altKey || event.metaKey) {
        return
      }

      const shortcut = shortcuts.find((s) => s.key.toLowerCase() === event.key.toLowerCase())
      if (shortcut) {
        event.preventDefault()
        shortcut.action()
      }
    },
    [shortcuts]
  )

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])
}

export function getShortcutsList(): { key: string; description: string }[] {
  return [
    { key: 'D', description: 'Dashboard' },
    { key: 'S', description: 'Search' },
    { key: 'L', description: 'Library' },
    { key: 'T', description: 'Downloads' },
    { key: 'G', description: 'Suggestions' },
    { key: ',', description: 'Settings' },
    { key: 'B', description: 'Toggle sidebar' },
  ]
}