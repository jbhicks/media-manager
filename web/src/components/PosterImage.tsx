import { useState, useEffect, useRef, useCallback } from 'react'
import { Loader2, Film } from 'lucide-react'

interface PosterImageProps {
  src?: string
  alt: string
  title?: string
  className?: string
  onLoad?: () => void
}

export function PosterImage({ src, alt, title, className = '', onLoad }: PosterImageProps) {
  const [isVisible, setIsVisible] = useState(false)
  const [isLoaded, setIsLoaded] = useState(false)
  const [imageSrc, setImageSrc] = useState<string | undefined>(src)
  const [isFetching, setIsFetching] = useState(false)
  const [fetchFailed, setFetchFailed] = useState(false)
  const imgRef = useRef<HTMLDivElement>(null)
  const hasAttemptedFetch = useRef(false)

  // Intersection Observer for lazy loading
  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true)
          observer.disconnect()
        }
      },
      {
        rootMargin: '200px', // Start loading 200px before visible
        threshold: 0.01,
      }
    )

    if (imgRef.current) {
      observer.observe(imgRef.current)
    }

    return () => observer.disconnect()
  }, [])

  // Fetch poster from backend if not provided
  const fetchPoster = useCallback(async () => {
    if (!title || hasAttemptedFetch.current) return
    
    hasAttemptedFetch.current = true
    setIsFetching(true)
    
    try {
      const response = await fetch(`/api/library/poster-by-title?title=${encodeURIComponent(title)}`)
      if (response.ok) {
        const data = await response.json()
        if (data.poster_url) {
          setImageSrc(data.poster_url)
          setFetchFailed(false)
        } else {
          setFetchFailed(true)
        }
      } else {
        setFetchFailed(true)
      }
    } catch (error) {
      console.error('Failed to fetch poster:', error)
      setFetchFailed(true)
    } finally {
      setIsFetching(false)
    }
  }, [title])

  // Trigger fetch when visible
  useEffect(() => {
    if (isVisible && !imageSrc && title && !hasAttemptedFetch.current) {
      fetchPoster()
    }
  }, [isVisible, imageSrc, title, fetchPoster])

  const handleImageLoad = () => {
    setIsLoaded(true)
    onLoad?.()
  }

  const handleImageError = () => {
    setFetchFailed(true)
    setIsLoaded(false)
  }

  // Update src if prop changes
  useEffect(() => {
    if (src && src !== imageSrc) {
      setImageSrc(src)
      setIsLoaded(false)
      setFetchFailed(false)
    }
  }, [src])

  const showSkeleton = !isLoaded || (!imageSrc && !fetchFailed)
  const showFallback = fetchFailed || (!imageSrc && !isFetching)

  return (
    <div ref={imgRef} className={`relative ${className}`}>
      {/* Skeleton placeholder */}
      {showSkeleton && (
        <div className="absolute inset-0 bg-gradient-to-br from-slate-700 to-slate-900 animate-pulse">
          {isFetching && (
            <div className="absolute inset-0 flex items-center justify-center">
              <Loader2 className="h-6 w-6 text-slate-500 animate-spin" />
            </div>
          )}
        </div>
      )}
      
      {/* Text fallback when no poster available */}
      {showFallback && (
        <div className="absolute inset-0 flex flex-col items-center justify-center bg-gradient-to-br from-slate-800 to-slate-900 text-white p-4">
          <Film className="h-12 w-12 text-slate-600 mb-2" />
          <p className="text-xs text-center text-slate-400 line-clamp-3 leading-relaxed">
            {alt || title}
          </p>
        </div>
      )}
      
      {/* Actual image */}
      {isVisible && imageSrc && !fetchFailed && (
        <img
          src={imageSrc}
          alt={alt}
          className={`absolute inset-0 w-full h-full object-cover transition-opacity duration-500 ${
            isLoaded ? 'opacity-100' : 'opacity-0'
          }`}
          onLoad={handleImageLoad}
          onError={handleImageError}
          loading="lazy"
        />
      )}
    </div>
  )
}
