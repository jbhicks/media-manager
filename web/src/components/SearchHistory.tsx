import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/Button'
import { Clock, X } from 'lucide-react'

interface SearchHistoryProps {
  onSelect: (query: string) =>
  void
  currentQuery: string
}

const STORAGE_KEY = 'media-manager-search-history'
const MAX_HISTORY = 20

export function SearchHistory({ onSelect, currentQuery }: SearchHistoryProps) {
  const [history, setHistory] = useState<string[]>([])
  const [isOpen, setIsOpen] = useState(false)

  useEffect(() => {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      try {
        setHistory(JSON.parse(stored))
      } catch (e) {
        console.error('Failed to parse search history:', e)
      }
    }
  }, [])

  const removeFromHistory = (query: string, e: React.MouseEvent) => {
    e.stopPropagation()
    setHistory(prev => {
      const updated = prev.filter(q => q !== query)
      localStorage.setItem(STORAGE_KEY, JSON.stringify(updated))
      return updated
    })
  }

  const clearHistory = () => {
    setHistory([])
    localStorage.removeItem(STORAGE_KEY)
  }

  // Filter history based on current query
  const filteredHistory = currentQuery
    ? history.filter(q => q.toLowerCase().includes(currentQuery.toLowerCase()))
    : history

  if (filteredHistory.length === 0 && !currentQuery) return null

  return (
    <div className="relative"
    >
      {isOpen && filteredHistory.length > 0 && (
        <div className="absolute top-full left-0 right-0 mt-1 bg-popover border rounded-md shadow-lg z-50 max-h-64 overflow-y-auto"
        >
          <div className="flex items-center justify-between p-2 border-b"
          >
            <span className="text-xs text-muted-foreground flex items-center gap-1"
            >
              <Clock className="h-3 w-3" />
              Recent Searches
            </span>
            <Button
              variant="ghost"
              size="sm"
              className="h-6 text-xs"
              onClick={clearHistory}
            >
              Clear
            </Button>
          </div>
          {filteredHistory.map((query, idx) => (
            <div
              key={idx}
              className="group flex items-center justify-between px-3 py-2 hover:bg-accent cursor-pointer text-sm"
              onClick={() => {
                onSelect(query)
                setIsOpen(false)
              }}
            >
              <span className="truncate">{query}</span>
              <Button
                variant="ghost"
                size="sm"
                className="h-6 w-6 p-0 opacity-0 hover:opacity-100 group-hover:opacity-100"
                onClick={(e) => removeFromHistory(query, e)}
              >
                <X className="h-3 w-3" />
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export function addToHistory(query: string) {
  if (!query.trim()) return
  const stored = localStorage.getItem(STORAGE_KEY)
  let history: string[] = []
  if (stored) {
    try {
      history = JSON.parse(stored)
    } catch (e) {
      console.error('Failed to parse search history:', e)
    }
  }
  const filtered = history.filter(q => q.toLowerCase() !== query.toLowerCase())
  const updated = [query, ...filtered].slice(0, MAX_HISTORY)
  localStorage.setItem(STORAGE_KEY, JSON.stringify(updated))
}
