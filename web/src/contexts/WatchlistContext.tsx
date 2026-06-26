import { createContext, useContext, useState, useCallback, type ReactNode } from 'react'

interface WatchlistItem {
  id: number
  user_id: number
  media_type: string
  tmdb_id: number
  title: string
  poster_url: string
  added_at: string
}

interface WatchlistContextType {
  watchlist: WatchlistItem[]
  isLoading: boolean
  error: string | null
  addToWatchlist: (item: { media_type: string; tmdb_id: number; title: string; poster_url?: string }) => Promise<void>
  removeFromWatchlist: (id: number) => Promise<void>
  isInWatchlist: (mediaType: string, tmdbId: number) => boolean
  refreshWatchlist: () => Promise<void>
}

const WatchlistContext = createContext<WatchlistContextType | undefined>(undefined)

export function WatchlistProvider({ children }: { children: ReactNode }) {
  const [watchlist, setWatchlist] = useState<WatchlistItem[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const getToken = () => localStorage.getItem('media_manager_token')

  const refreshWatchlist = useCallback(async () => {
    const token = getToken()
    if (!token) return

    setIsLoading(true)
    setError(null)
    try {
      const response = await fetch('/api/watchlist', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      if (!response.ok) throw new Error('Failed to fetch watchlist')
      const data = await response.json()
      setWatchlist(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setIsLoading(false)
    }
  }, [])

  const addToWatchlist = useCallback(async (item: { media_type: string; tmdb_id: number; title: string; poster_url?: string }) => {
    const token = getToken()
    if (!token) {
      setError('Please login to add to watchlist')
      return
    }

    setError(null)
    try {
      const response = await fetch('/api/watchlist', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify(item)
      })
      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.message || 'Failed to add to watchlist')
      }
      await refreshWatchlist()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    }
  }, [refreshWatchlist])

  const removeFromWatchlist = useCallback(async (id: number) => {
    const token = getToken()
    if (!token) return

    setError(null)
    try {
      const response = await fetch(`/api/watchlist/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      if (!response.ok) throw new Error('Failed to remove from watchlist')
      await refreshWatchlist()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    }
  }, [refreshWatchlist])

  const isInWatchlist = useCallback((mediaType: string, tmdbId: number) => {
    return watchlist.some(item => item.media_type === mediaType && item.tmdb_id === tmdbId)
  }, [watchlist])

  return (
    <WatchlistContext.Provider value={{
      watchlist,
      isLoading,
      error,
      addToWatchlist,
      removeFromWatchlist,
      isInWatchlist,
      refreshWatchlist
    }}>
      {children}
    </WatchlistContext.Provider>
  )
}

export function useWatchlist() {
  const context = useContext(WatchlistContext)
  if (context === undefined) {
    throw new Error('useWatchlist must be used within a WatchlistProvider')
  }
  return context
}
