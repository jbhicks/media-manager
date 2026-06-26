import { createContext, useContext, useState, useCallback, useEffect } from 'react'

interface WatchProgress {
  media_type: string
  media_id: number
  position: number
  duration: number
  progress: number
}

interface WatchHistoryContextType {
  updateProgress: (data: WatchProgress) => Promise<void>
  markComplete: (mediaType: string, mediaId: number) => Promise<void>
  getResumePoints: () => Promise<WatchProgress[]>
  resumePoints: WatchProgress[]
  isLoading: boolean
}

const WatchHistoryContext = createContext<WatchHistoryContextType | undefined>(undefined)

export function WatchHistoryProvider({ children }: { children: React.ReactNode }) {
  const [resumePoints, setResumePoints] = useState<WatchProgress[]>([])
  const [isLoading, setIsLoading] = useState(false)

  const getToken = () => localStorage.getItem('media_manager_token')

  const updateProgress = useCallback(async (data: WatchProgress) => {
    const token = getToken()
    if (!token) return

    try {
      const response = await fetch('/api/history/progress', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(data),
      })

      if (!response.ok) {
        console.error('Failed to update progress:', response.status)
      }
    } catch (error) {
      console.error('Error updating progress:', error)
    }
  }, [])

  const markComplete = useCallback(async (mediaType: string, mediaId: number) => {
    const token = getToken()
    if (!token) return

    try {
      const response = await fetch('/api/history/complete', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ media_type: mediaType, media_id: mediaId }),
      })

      if (!response.ok) {
        console.error('Failed to mark complete:', response.status)
      }
    } catch (error) {
      console.error('Error marking complete:', error)
    }
  }, [])

  const getResumePoints = useCallback(async () => {
    const token = getToken()
    if (!token) return []

    setIsLoading(true)
    try {
      const response = await fetch('/api/history/resume', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      })

      if (response.ok) {
        const data = await response.json()
        const points = data.resume_points || []
        setResumePoints(points)
        return points
      }
    } catch (error) {
      console.error('Error fetching resume points:', error)
    } finally {
      setIsLoading(false)
    }
    return []
  }, [])

  // Fetch resume points on mount
  useEffect(() => {
    getResumePoints()
  }, [getResumePoints])

  return (
    <WatchHistoryContext.Provider
      value={{
        updateProgress,
        markComplete,
        getResumePoints,
        resumePoints,
        isLoading,
      }}
    >
      {children}
    </WatchHistoryContext.Provider>
  )
}

export function useWatchHistory() {
  const context = useContext(WatchHistoryContext)
  if (context === undefined) {
    throw new Error('useWatchHistory must be used within a WatchHistoryProvider')
  }
  return context
}
