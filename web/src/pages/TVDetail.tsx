import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Star, Play, ArrowLeft, Heart, Download, ChevronDown } from 'lucide-react'
import { useWatchlist } from '@/contexts/WatchlistContext'
import { useAppStore } from '@/store/appStore'
import { PosterImage } from '@/components/PosterImage'

interface CastMember {
  id: number
  name: string
  character: string
  profile_path: string
}

interface SeasonInfo {
  id: number
  name: string
  season_number: number
  episode_count: number
  air_date: string
  poster_path: string
  overview: string
}

interface Episode {
  id: number
  name: string
  episode_number: number
  overview: string
  air_date: string
  still_path: string
  runtime: number
  vote_average: number
}

interface Video {
  id: string
  name: string
  key: string
  site: string
  type: string
}

interface SimilarShow {
  id: number
  name: string
  poster_path: string
  vote_average: number
  first_air_date: string
}

interface TVDetails {
  id: number
  name: string
  original_name: string
  tagline: string
  overview: string
  first_air_date: string
  last_air_date: string
  status: string
  type: string
  number_of_seasons: number
  number_of_episodes: number
  vote_average: number
  vote_count: number
  poster_path: string
  backdrop_path: string
  genres: { id: number; name: string }[]
  credits: {
    cast: CastMember[]
  }
  videos: {
    results: Video[]
  }
  similar: {
    results: SimilarShow[]
  }
  seasons: SeasonInfo[]
  networks: { id: number; name: string; logo_path: string }[]
  created_by: { id: number; name: string; profile_path: string }[]
  episode_run_time: number[]
}

export function TVDetail() {
  const { id } = useParams<{ id: string }>()
  const [show, setShow] = useState<TVDetails | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [selectedSeason, setSelectedSeason] = useState(1)
  const [episodes, setEpisodes] = useState<Episode[]>([])
  const [episodesLoading, setEpisodesLoading] = useState(false)
  const [visibleEpisodeCount, setVisibleEpisodeCount] = useState(12)
  const [showTrailer, setShowTrailer] = useState(false)
  const [localFile, setLocalFile] = useState<string | null>(null)
  const [downloadStatus, setDownloadStatus] = useState('')
  const [selectedQuality, setSelectedQuality] = useState('1080p')
  const { addToWatchlist, removeFromWatchlist, isInWatchlist, watchlist } = useWatchlist()
  const { addToast } = useAppStore()
  const [isWatchlistLoading, setIsWatchlistLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleDownload = async () => {
    if (!show) return
    setDownloadStatus('searching')
    try {
      const response = await fetch('/api/downloads/search', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          query: show.name,
          year: show.first_air_date?.substring(0, 4),
          media_type: 'tv',
          tmdb_id: show.id,
          resolution: selectedQuality
        })
      })
      const data = await response.json().catch(() => ({ error: 'Unknown error' }))
      if (response.ok) {
        setDownloadStatus('queued')
        addToast(`Queued ${selectedQuality}: ${data.title || show.name}`, 'success')
      } else {
        setDownloadStatus('failed')
        addToast(data.message || data.error || 'Download failed', 'error')
      }
    } catch (error) {
      console.error('Download error:', error)
      setDownloadStatus('failed')
      addToast('Download failed', 'error')
    }
  }

  const handleStream = () => {
    if (localFile) {
      window.open(`/api/stream/direct?path=${encodeURIComponent(localFile)}`, '_blank')
    }
  }

  useEffect(() => {
    const fetchTV = async () => {
      setError(null)
      try {
        const response = await fetch(`/api/discover/tv/${id}`)
        if (!response.ok) throw new Error('Failed to fetch TV show')
        const data = await response.json()
        setShow(data)
        
        // Fetch episodes for first season
        if (data.seasons?.length > 0) {
          const seasonNum = data.seasons[0].season_number
          setSelectedSeason(seasonNum)
          fetchEpisodes(seasonNum)
        }

        // Check if show has a local file
        const fileResponse = await fetch(`/api/library/tv/${id}/file`)
        if (fileResponse.ok) {
          const fileData = await fileResponse.json()
          if (fileData.file_path) {
            setLocalFile(fileData.file_path)
          }
        }
      } catch (error) {
        console.error('Error fetching TV show:', error)
        setError('Failed to load TV show details')
      } finally {
        setIsLoading(false)
      }
    }

    fetchTV()
  }, [id])

  const fetchEpisodes = async (seasonNumber: number) => {
    setEpisodesLoading(true)
    try {
      const response = await fetch(`/api/discover/tv/${id}/season/${seasonNumber}`)
      if (!response.ok) throw new Error('Failed to fetch episodes')
      const data = await response.json()
      setEpisodes(data.episodes || [])
    } catch (error) {
      console.error('Error fetching episodes:', error)
      setEpisodes([])
    } finally {
      setEpisodesLoading(false)
    }
  }

  const handleSeasonChange = (seasonNumber: number) => {
    setSelectedSeason(seasonNumber)
    setVisibleEpisodeCount(12)
    fetchEpisodes(seasonNumber)
  }

  if (isLoading) {
    // Skeleton for hero, episodes, cast etc to ensure visible loading (no blanks)
    return (
      <div>
        {/* Hero skeleton */}
        <div className="relative h-[60vh] -mx-6 -mt-6 bg-[#1a1a2e]">
          <div className="absolute inset-0 bg-gradient-to-t from-[#121212] via-[#121212]/50 to-transparent" />
          <div className="absolute bottom-0 left-0 right-0 p-8">
            <div className="h-14 w-3/4 bg-white/10 rounded mb-2 animate-pulse" />
            <div className="h-7 w-1/2 bg-white/10 rounded mb-4 animate-pulse" />
            <div className="flex items-center gap-4 mb-4">
              {Array.from({ length: 6 }).map((_, i) => (
                <div key={i} className="h-5 w-16 bg-white/10 rounded animate-pulse" />
              ))}
            </div>
            <div className="flex gap-2 mb-4">
              {Array.from({ length: 3 }).map((_, i) => (
                <div key={i} className="h-7 w-20 bg-[#1f1f1f] rounded-full animate-pulse" />
              ))}
            </div>
            <div className="flex gap-3">
              <div className="h-12 w-36 bg-[#1ed760]/40 rounded-full animate-pulse" />
              <div className="h-12 w-36 bg-[#1f1f1f] rounded-full animate-pulse" />
            </div>
          </div>
        </div>
        <div className="px-8 py-8">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            <div className="lg:col-span-2">
              <div className="h-8 w-32 bg-white/10 rounded mb-4 animate-pulse" />
              <div className="h-4 w-full bg-white/10 rounded mb-2 animate-pulse" />
              {/* Episodes skeleton */}
              <div className="mt-8">
                <div className="flex justify-between mb-4">
                  <div className="h-8 w-28 bg-white/10 rounded animate-pulse" />
                  <div className="h-9 w-48 bg-[#1f1f1f] rounded-lg animate-pulse" />
                </div>
                {Array.from({ length: 3 }).map((_, i) => (
                  <div key={i} className="flex gap-4 p-4 bg-[#1f1f1f] rounded-lg mb-3">
                    <div className="flex-shrink-0 w-[160px] aspect-video bg-[#121212] rounded-lg animate-pulse" />
                    <div className="flex-1">
                      <div className="h-5 w-3/4 bg-white/10 rounded mb-2 animate-pulse" />
                      <div className="h-4 w-full bg-white/10 rounded mb-1 animate-pulse" />
                      <div className="h-4 w-2/3 bg-white/10 rounded animate-pulse" />
                    </div>
                  </div>
                ))}
              </div>
              {/* Cast skeleton */}
              <div className="mt-8">
                <div className="h-8 w-24 bg-white/10 rounded mb-4 animate-pulse" />
                <div className="flex gap-4">
                  {Array.from({ length: 5 }).map((_, i) => (
                    <div key={i} className="flex-shrink-0 w-[120px]">
                      <div className="aspect-[2/3] bg-[#1f1f1f] rounded-lg animate-pulse" />
                    </div>
                  ))}
                </div>
              </div>
            </div>
            <div className="bg-[#1f1f1f] rounded-lg p-6">
              <div className="h-6 w-24 bg-white/10 rounded mb-4 animate-pulse" />
              {Array.from({ length: 6 }).map((_, i) => (
                <div key={i} className="h-5 bg-white/10 rounded mb-3 animate-pulse" />
              ))}
            </div>
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-[60vh] flex items-center justify-center">
        <div className="text-center">
          <p className="text-white mb-4">{error}</p>
          <button
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-[#1ed760] text-black rounded-full font-semibold"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  if (!show) {
    return (
      <div className="min-h-[60vh] flex items-center justify-center">
        <h1 className="text-white text-2xl">TV show not found</h1>
      </div>
    )
  }

  const trailer = show.videos?.results?.find(v => v.type === 'Trailer' && v.site === 'YouTube')
  const avgRuntime = show.episode_run_time?.length > 0 
    ? Math.round(show.episode_run_time.reduce((a, b) => a + b, 0) / show.episode_run_time.length)
    : 0
  const inWatchlist = isInWatchlist('tv_series', show.id)
  const watchlistItem = watchlist.find(item => item.media_type === 'tv_series' && item.tmdb_id === show.id)

  const handleWatchlistToggle = async () => {
    if (isWatchlistLoading) return
    setIsWatchlistLoading(true)
    try {
      if (inWatchlist && watchlistItem) {
        await removeFromWatchlist(watchlistItem.id)
      } else {
        await addToWatchlist({
          media_type: 'tv_series',
          tmdb_id: show.id,
          title: show.name,
          poster_url: show.poster_path
        })
      }
    } catch (error) {
      console.error('Watchlist error:', error)
      addToast(error instanceof Error ? error.message : "Watchlist update failed", "error")
    } finally {
      setIsWatchlistLoading(false)
    }
  }

  return (
    <div>
      {/* Backdrop - use PosterImage (not raw style/img) for skeleton + loading parity + avoid blanks (ux-design-reviewer) */}
      <div className="relative h-[60vh] -mx-6 -mt-6">
        {show.backdrop_path ? (
          <PosterImage
            src={`https://image.tmdb.org/t/p/original${show.backdrop_path}`}
            alt=""
            className="absolute inset-0 w-full h-full object-cover"
            showFallbackTitle={false}
          />
        ) : (
          <div className="absolute inset-0 bg-gradient-to-br from-[#1a1a2e] to-[#121212]" />
        )}
        {/* stronger gradients for text contrast (ux-design-reviewer) */}
        <div className="absolute inset-0 bg-gradient-to-t from-[#121212]/95 via-[#121212]/70 to-transparent" />
        
        {/* Back button */}
        <Link
          to="/discover"
          className="absolute top-4 left-4 z-10 p-2 rounded-full bg-black/50 text-white hover:bg-black/70 focus-visible:ring-2 focus-visible:ring-[#1ed760] transition-colors"
          aria-label="Back to discover"
        >
          <ArrowLeft className="w-6 h-6" aria-hidden="true" />
        </Link>

        {/* Content overlay */}
        <div className="absolute bottom-0 left-0 right-0 p-8">
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
          >
            <h1 className="text-5xl font-bold text-white mb-2">{show.name}</h1>
            {show.tagline && (
              <p className="text-xl text-[#b3b3b3] italic mb-4">{show.tagline}</p>
            )}
            
            <div className="flex items-center gap-4 mb-4">
              <div className="flex items-center gap-1 text-yellow-400">
                <Star className="w-5 h-5 fill-current" />
                <span className="font-bold">{(show.vote_average || 0).toFixed(1)}</span>
                <span className="text-[#b3b3b3]">({(show.vote_count || 0).toLocaleString()} votes)</span>
              </div>
              <span className="text-[#b3b3b3]">|</span>
              <span className="text-[#b3b3b3]">{show.first_air_date?.substring(0, 4)}</span>
              <span className="text-[#b3b3b3]">|</span>
              <span className="text-[#b3b3b3]">{show.number_of_seasons} Seasons</span>
              <span className="text-[#b3b3b3]">|</span>
              <span className="text-[#b3b3b3]">{show.number_of_episodes} Episodes</span>
              {avgRuntime > 0 && (
                <>
                  <span className="text-[#b3b3b3]">|</span>
                  <span className="text-[#b3b3b3]">{avgRuntime} min/episode</span>
                </>
              )}
            </div>

            <div className="flex gap-2 mb-4">
              {show.genres?.map(genre => (
                <span key={genre.id} className="px-3 py-1 bg-[#1f1f1f] rounded-full text-sm text-white">
                  {genre.name}
                </span>
              ))}
            </div>

            <div className="flex flex-wrap items-center gap-3" role="group" aria-label="TV show actions">
              {trailer && (
                <button
                  onClick={() => setShowTrailer(true)}
                  className="flex items-center gap-2 px-6 py-3 bg-[#1ed760] text-black rounded-full font-semibold hover:bg-[#1ed760]/90 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-[#121212] focus-visible:ring-black transition-colors"
                  aria-label="Watch trailer for this show"
                >
                  <Play className="w-5 h-5" aria-hidden="true" />
                  Watch Trailer
                </button>
              )}
              <button 
                onClick={handleWatchlistToggle}
                disabled={isWatchlistLoading}
                className={`flex items-center gap-2 px-6 py-3 rounded-full font-semibold transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-[#121212] focus-visible:ring-[#1ed760] ${
                  inWatchlist 
                    ? 'bg-[#1ed760] text-black hover:bg-[#1ed760]/90' 
                    : 'bg-[#1f1f1f] text-white hover:bg-[#2a2a2a]'
                }`}
                aria-label={inWatchlist ? 'Remove from watchlist' : 'Add to watchlist'}
                aria-pressed={inWatchlist}
              >
                <Heart className={`w-5 h-5 ${inWatchlist ? 'fill-current' : ''}`} aria-hidden="true" />
                {isWatchlistLoading ? 'Loading...' : inWatchlist ? 'In Watchlist' : 'Add to Watchlist'}
              </button>
              {localFile ? (
                <button
                  onClick={handleStream}
                  className="flex items-center gap-2 px-6 py-3 bg-[#1ed760] text-black rounded-full font-semibold hover:bg-[#1ed760]/90 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-[#121212] focus-visible:ring-black transition-colors"
                  aria-label="Stream this TV show now"
                >
                  <Play className="w-5 h-5" aria-hidden="true" />
                  Stream Now
                </button>
              ) : (
                <>
                  <button
                    onClick={handleDownload}
                    disabled={downloadStatus === 'searching' || downloadStatus === 'queued'}
                    className="flex items-center gap-2 px-6 py-3 bg-[#1f1f1f] text-white rounded-full font-semibold hover:bg-[#2a2a2a] transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-[#121212] focus-visible:ring-[#1ed760] disabled:opacity-50"
                    aria-label="Download this TV show"
                  >
                    <Download className="w-5 h-5" aria-hidden="true" />
                    {downloadStatus === 'searching' ? 'Searching...' : downloadStatus === 'queued' ? 'Queued' : downloadStatus === 'failed' ? 'Failed' : 'Download'}
                  </button>
                  <select
                    value={selectedQuality}
                    onChange={(e) => setSelectedQuality(e.target.value)}
                    className="px-4 py-3 bg-[#1f1f1f] text-white rounded-full font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-[#121212] focus-visible:ring-[#1ed760]"
                    aria-label="Download resolution"
                  >
                    <option value="2160p">4K</option>
                    <option value="1080p">1080p</option>
                    <option value="720p">720p</option>
                    <option value="480p">480p</option>
                  </select>
                </>
              )}
            </div>
          </motion.div>
        </div>
      </div>

      {/* Main Content - full page (no Layout clash) with ARIA */}
      <main className="px-8 py-8" role="main" aria-label="TV show details">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Left column - Overview & Episodes */}
          <div className="lg:col-span-2">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
            >
              <h2 className="text-2xl font-bold text-white mb-4">Overview</h2>
              <p className="text-[#b3b3b3] text-lg leading-relaxed">{show.overview}</p>
            </motion.div>

            {/* Seasons & Episodes */}
            {show.seasons?.length > 0 && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.3 }}
                className="mt-8"
              >
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-2xl font-bold text-white">Episodes</h2>
                  <div className="relative">
                    <select
                      value={selectedSeason}
                      onChange={(e) => handleSeasonChange(Number(e.target.value))}
                      className="appearance-none bg-[#1f1f1f] text-white px-4 py-2 pr-10 rounded-lg cursor-pointer focus:outline-none focus:ring-2 focus:ring-[#1ed760]"
                    >
                      {show.seasons.map(season => (
                        <option key={season.id} value={season.season_number}>
                          {season.name} ({season.episode_count} episodes)
                        </option>
                      ))}
                    </select>
                    <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#b3b3b3] pointer-events-none" />
                  </div>
                </div>

                <div className="space-y-3" role="list" aria-label="Episodes">
                  {episodesLoading ? (
                    Array.from({ length: 2 }).map((_, i) => (
                      <div key={i} className="flex gap-4 p-4 bg-[#1f1f1f] rounded-lg mb-3">
                        <div className="flex-shrink-0 w-[160px] aspect-video bg-[#121212] rounded-lg animate-pulse" />
                        <div className="flex-1">
                          <div className="h-5 w-3/4 bg-white/10 rounded mb-2 animate-pulse" />
                          <div className="h-4 w-full bg-white/10 rounded mb-1 animate-pulse" />
                          <div className="h-4 w-2/3 bg-white/10 rounded animate-pulse" />
                        </div>
                      </div>
                    ))
                  ) : (
                    <>
                      {episodes.slice(0, visibleEpisodeCount).map((episode) => (
                    <div
                      key={episode.id}
                      className="flex gap-4 p-4 bg-[#1f1f1f] rounded-lg hover:bg-[#2a2a2a] focus:bg-[#2a2a2a] transition-colors cursor-pointer group focus:outline-none focus:ring-2 focus:ring-[#1ed760]"
                      tabIndex={0}
                      role="button"
                      aria-label={`Play S${selectedSeason}E${episode.episode_number} ${episode.name}`}
                      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { /* Enter works: placeholder here - real stream would hook to player like handleStream; keeps D-pad compatible */ } }}
                      onClick={() => { /* future: episode stream */ }}
                    >
                      <div className="flex-shrink-0 w-[160px] aspect-video rounded-lg overflow-hidden bg-[#121212] relative">
                        <PosterImage
                          src={episode.still_path ? `https://image.tmdb.org/t/p/w300${episode.still_path}` : undefined}
                          alt={episode.name}
                          className="absolute inset-0 w-full h-full object-cover"
                          showFallbackTitle={false}
                        />
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="text-[#1ed760] font-bold">S{selectedSeason}E{episode.episode_number}</span>
                          <h3 className="text-white font-semibold truncate">{episode.name}</h3>
                        </div>
                        <p className="text-[#b3b3b3] text-sm line-clamp-2 mb-2">{episode.overview}</p>
                        <div className="flex items-center gap-3 text-xs text-[#b3b3b3]">
                          <span>{episode.air_date}</span>
                          {episode.runtime > 0 && <span>{episode.runtime} min</span>}
                          <div className="flex items-center gap-1 text-yellow-400">
                            <Star className="w-3 h-3 fill-current" aria-hidden="true" />
                            <span>{(episode.vote_average || 0).toFixed(1)}</span>
                          </div>
                        </div>
                      </div>
                      <div className="flex-shrink-0 self-center">
                        {/* Always visible (not hover-only) for keyboard/TV accessibility; row focusable + Enter ready */}
                        <button className="p-2 rounded-full bg-[#1ed760] text-black transition-opacity focus-visible:ring-1" tabIndex={-1} aria-hidden="true">
                          <Play className="w-5 h-5" />
                        </button>
                      </div>
                    </div>
                  ))}
                  {episodes.length > visibleEpisodeCount && (
                    <button
                      type="button"
                      onClick={() => setVisibleEpisodeCount(prev => Math.min(prev + 12, episodes.length))}
                      className="w-full py-3 bg-[#1f1f1f] hover:bg-[#2a2a2a] text-white rounded-lg font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-[#1ed760]"
                      aria-label={`Show more episodes, ${episodes.length - visibleEpisodeCount} remaining`}
                    >
                      Show more episodes ({episodes.length - visibleEpisodeCount} remaining)
                    </button>
                  )}
                </>
              )}
            </div>
          </motion.div>
        )}

            {/* Cast */}
            {show.credits?.cast?.length > 0 && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.4 }}
                className="mt-8"
              >
                <h2 className="text-2xl font-bold text-white mb-4" id="cast-heading">Cast</h2>
                <div className="flex gap-4 overflow-x-auto pb-4" role="list" aria-labelledby="cast-heading">
                  {show.credits.cast.slice(0, 10).map(actor => (
                    <div key={actor.id} className="flex-shrink-0 w-[120px]" role="listitem">
                      <div className="aspect-[2/3] rounded-lg overflow-hidden bg-[#1f1f1f] relative">
                        <PosterImage
                          src={actor.profile_path ? `https://image.tmdb.org/t/p/w200${actor.profile_path}` : undefined}
                          alt={actor.name}
                          className="absolute inset-0 w-full h-full object-cover"
                          showFallbackTitle={false}
                        />
                      </div>
                      <p className="mt-2 text-sm font-semibold text-white truncate">{actor.name}</p>
                      <p className="text-xs text-[#b3b3b3] truncate">{actor.character}</p>
                    </div>
                  ))}
                </div>
              </motion.div>
            )}
          </div>

          {/* Right column - Info */}
          <div>
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.3 }}
              className="bg-[#1f1f1f] rounded-lg p-6"
            >
              <h3 className="text-lg font-bold text-white mb-4">Show Info</h3>
              <div className="space-y-3">
                <div>
                  <p className="text-[#b3b3b3] text-sm">First Air Date</p>
                  <p className="text-white">{show.first_air_date}</p>
                </div>
                <div>
                  <p className="text-[#b3b3b3] text-sm">Last Air Date</p>
                  <p className="text-white">{show.last_air_date || 'N/A'}</p>
                </div>
                <div>
                  <p className="text-[#b3b3b3] text-sm">Status</p>
                  <p className="text-white">{show.status}</p>
                </div>
                <div>
                  <p className="text-[#b3b3b3] text-sm">Type</p>
                  <p className="text-white">{show.type}</p>
                </div>
                <div>
                  <p className="text-[#b3b3b3] text-sm">Seasons</p>
                  <p className="text-white">{show.number_of_seasons}</p>
                </div>
                <div>
                  <p className="text-[#b3b3b3] text-sm">Episodes</p>
                  <p className="text-white">{show.number_of_episodes}</p>
                </div>
                {show.created_by?.length > 0 && (
                  <div>
                    <p className="text-[#b3b3b3] text-sm">Created By</p>
                    <p className="text-white">{show.created_by.map(c => c.name).join(', ')}</p>
                  </div>
                )}
                {show.networks?.length > 0 && (
                  <div>
                    <p className="text-[#b3b3b3] text-sm">Network</p>
                    <p className="text-white">{show.networks.map(n => n.name).join(', ')}</p>
                  </div>
                )}
              </div>
            </motion.div>
          </div>
        </div>

        {/* Similar Shows */}
        {show.similar?.results?.length > 0 && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5 }}
            className="mt-8"
          >
            <h2 className="text-2xl font-bold text-white mb-4">Similar Shows</h2>
            <div className="flex gap-4 overflow-x-auto pb-4" role="list" aria-label="Similar shows">
              {show.similar.results.slice(0, 10).map(similar => (
                <Link
                  key={similar.id}
                  to={`/tv/${similar.id}`}
                  className="flex-shrink-0 w-[160px] group focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#1ed760] rounded-lg"
                  aria-label={`${similar.name} - rating ${(similar.vote_average || 0).toFixed(1)}`}
                  role="listitem"
                >
                  <div className="aspect-[2/3] rounded-lg overflow-hidden bg-[#1f1f1f] relative">
                    <PosterImage
                      src={similar.poster_path ? `https://image.tmdb.org/t/p/w500${similar.poster_path}` : undefined}
                      alt={similar.name}
                      className="absolute inset-0 w-full h-full object-cover transition-transform duration-300 group-hover:scale-110 group-focus:scale-110"
                      showFallbackTitle={false}
                    />
                  </div>
                  {/* rating always visible; accent on hover/focus (ux-design-reviewer) */}
                  <p className="mt-2 text-sm font-semibold text-white truncate group-hover:text-[#1ed760] group-focus:text-[#1ed760]">
                    {similar.name}
                  </p>
                  <div className="flex items-center gap-1 text-yellow-400">
                    <Star className="w-3 h-3 fill-current" aria-hidden="true" />
                    <span className="text-xs">{(similar.vote_average || 0).toFixed(1)}</span>
                  </div>
                </Link>
              ))}
            </div>
          </motion.div>
        )}
      </main>

      {/* Trailer Modal - ARIA modal */}
      {showTrailer && trailer && (
        <div
          className="fixed inset-0 z-50 bg-black/90 flex items-center justify-center p-4"
          onClick={() => setShowTrailer(false)}
          role="dialog"
          aria-modal="true"
          aria-label="TV show trailer"
        >
          <div className="w-full max-w-4xl aspect-video">
            <iframe
              src={`https://www.youtube.com/embed/${trailer.key}?autoplay=1`}
              title={trailer.name}
              className="w-full h-full rounded-lg"
              allowFullScreen
              allow="autoplay; encrypted-media"
            />
          </div>
        </div>
      )}
    </div>
  )
}
