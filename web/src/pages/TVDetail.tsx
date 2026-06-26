import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Star, Play, ArrowLeft, Tv, Heart, Download, ChevronDown } from 'lucide-react'
import { useWatchlist } from '@/contexts/WatchlistContext'

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
  const [showTrailer, setShowTrailer] = useState(false)
  const { addToWatchlist, removeFromWatchlist, isInWatchlist, watchlist } = useWatchlist()
  const [isWatchlistLoading, setIsWatchlistLoading] = useState(false)

  useEffect(() => {
    const fetchTV = async () => {
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
      } catch (error) {
        console.error('Error fetching TV show:', error)
      } finally {
        setIsLoading(false)
      }
    }

    fetchTV()
  }, [id])

  const fetchEpisodes = async (seasonNumber: number) => {
    try {
      const response = await fetch(`/api/discover/tv/${id}/season/${seasonNumber}`)
      if (!response.ok) throw new Error('Failed to fetch episodes')
      const data = await response.json()
      setEpisodes(data.episodes || [])
    } catch (error) {
      console.error('Error fetching episodes:', error)
      setEpisodes([])
    }
  }

  const handleSeasonChange = (seasonNumber: number) => {
    setSelectedSeason(seasonNumber)
    fetchEpisodes(seasonNumber)
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-[#121212] flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-[#1ed760]" />
      </div>
    )
  }

  if (!show) {
    return (
      <div className="min-h-screen bg-[#121212] flex items-center justify-center">
        <p className="text-white">TV show not found</p>
      </div>
    )
  }

  const trailer = show.videos?.results?.find(v => v.type === 'Trailer' && v.site === 'YouTube')
  const avgRuntime = show.episode_run_time?.length > 0 
    ? Math.round(show.episode_run_time.reduce((a, b) => a + b, 0) / show.episode_run_time.length)
    : 0
  const inWatchlist = isInWatchlist('tv_show', show.id)
  const watchlistItem = watchlist.find(item => item.media_type === 'tv_show' && item.tmdb_id === show.id)

  const handleWatchlistToggle = async () => {
    if (isWatchlistLoading) return
    setIsWatchlistLoading(true)
    try {
      if (inWatchlist && watchlistItem) {
        await removeFromWatchlist(watchlistItem.id)
      } else {
        await addToWatchlist({
          media_type: 'tv_show',
          tmdb_id: show.id,
          title: show.name,
          poster_url: show.poster_path
        })
      }
    } catch (error) {
      console.error('Watchlist error:', error)
    } finally {
      setIsWatchlistLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-[#121212]">
      {/* Backdrop */}
      <div className="relative h-[60vh]">
        {show.backdrop_path ? (
          <div
            className="absolute inset-0 bg-cover bg-center"
            style={{ backgroundImage: `url(https://image.tmdb.org/t/p/original${show.backdrop_path})` }}
          />
        ) : (
          <div className="absolute inset-0 bg-gradient-to-br from-[#1a1a2e] to-[#121212]" />
        )}
        <div className="absolute inset-0 bg-gradient-to-t from-[#121212] via-[#121212]/50 to-transparent" />
        
        {/* Back button */}
        <Link
          to="/discover"
          className="absolute top-4 left-4 z-10 p-2 rounded-full bg-black/50 text-white hover:bg-black/70 transition-colors"
        >
          <ArrowLeft className="w-6 h-6" />
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
                <span className="font-bold">{show.vote_average.toFixed(1)}</span>
                <span className="text-[#b3b3b3]">({show.vote_count.toLocaleString()} votes)</span>
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

            <div className="flex gap-3">
              {trailer && (
                <button
                  onClick={() => setShowTrailer(true)}
                  className="flex items-center gap-2 px-6 py-3 bg-[#1ed760] text-black rounded-full font-semibold hover:bg-[#1ed760]/90 transition-colors"
                >
                  <Play className="w-5 h-5" />
                  Watch Trailer
                </button>
              )}
              <button 
                onClick={handleWatchlistToggle}
                disabled={isWatchlistLoading}
                className={`flex items-center gap-2 px-6 py-3 rounded-full font-semibold transition-colors ${
                  inWatchlist 
                    ? 'bg-[#1ed760] text-black hover:bg-[#1ed760]/90' 
                    : 'bg-[#1f1f1f] text-white hover:bg-[#2a2a2a]'
                }`}
              >
                <Heart className={`w-5 h-5 ${inWatchlist ? 'fill-current' : ''}`} />
                {isWatchlistLoading ? 'Loading...' : inWatchlist ? 'In Watchlist' : 'Add to Watchlist'}
              </button>
              <button className="flex items-center gap-2 px-6 py-3 bg-[#1f1f1f] text-white rounded-full font-semibold hover:bg-[#2a2a2a] transition-colors">
                <Download className="w-5 h-5" />
                Download
              </button>
            </div>
          </motion.div>
        </div>
      </div>

      {/* Main Content */}
      <div className="px-8 py-8">
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

                <div className="space-y-3">
                  {episodes.map((episode) => (
                    <div
                      key={episode.id}
                      className="flex gap-4 p-4 bg-[#1f1f1f] rounded-lg hover:bg-[#2a2a2a] transition-colors cursor-pointer group"
                    >
                      <div className="flex-shrink-0 w-[160px] aspect-video rounded-lg overflow-hidden bg-[#121212]">
                        {episode.still_path ? (
                          <img
                            src={`https://image.tmdb.org/t/p/w300${episode.still_path}`}
                            alt={episode.name}
                            className="w-full h-full object-cover"
                          />
                        ) : (
                          <div className="w-full h-full flex items-center justify-center">
                            <Tv className="w-8 h-8 text-[#4d4d4d]" />
                          </div>
                        )}
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
                            <Star className="w-3 h-3 fill-current" />
                            <span>{episode.vote_average.toFixed(1)}</span>
                          </div>
                        </div>
                      </div>
                      <div className="flex-shrink-0 self-center">
                        <button className="p-2 rounded-full bg-[#1ed760] text-black opacity-0 group-hover:opacity-100 transition-opacity">
                          <Play className="w-5 h-5" />
                        </button>
                      </div>
                    </div>
                  ))}
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
                <h2 className="text-2xl font-bold text-white mb-4">Cast</h2>
                <div className="flex gap-4 overflow-x-auto pb-4">
                  {show.credits.cast.slice(0, 10).map(actor => (
                    <div key={actor.id} className="flex-shrink-0 w-[120px]">
                      <div className="aspect-[2/3] rounded-lg overflow-hidden bg-[#1f1f1f]">
                        {actor.profile_path ? (
                          <img
                            src={`https://image.tmdb.org/t/p/w200${actor.profile_path}`}
                            alt={actor.name}
                            className="w-full h-full object-cover"
                          />
                        ) : (
                          <div className="w-full h-full flex items-center justify-center">
                            <Tv className="w-8 h-8 text-[#4d4d4d]" />
                          </div>
                        )}
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
            <div className="flex gap-4 overflow-x-auto pb-4">
              {show.similar.results.slice(0, 10).map(similar => (
                <Link
                  key={similar.id}
                  to={`/tv/${similar.id}`}
                  className="flex-shrink-0 w-[160px] group"
                >
                  <div className="aspect-[2/3] rounded-lg overflow-hidden bg-[#1f1f1f]">
                    {similar.poster_path ? (
                      <img
                        src={`https://image.tmdb.org/t/p/w500${similar.poster_path}`}
                        alt={similar.name}
                        className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-110"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center">
                        <Tv className="w-8 h-8 text-[#4d4d4d]" />
                      </div>
                    )}
                  </div>
                  <p className="mt-2 text-sm font-semibold text-white truncate group-hover:text-[#1ed760]">
                    {similar.name}
                  </p>
                  <div className="flex items-center gap-1 text-yellow-400">
                    <Star className="w-3 h-3 fill-current" />
                    <span className="text-xs">{similar.vote_average.toFixed(1)}</span>
                  </div>
                </Link>
              ))}
            </div>
          </motion.div>
        )}
      </div>

      {/* Trailer Modal */}
      {showTrailer && trailer && (
        <div
          className="fixed inset-0 z-50 bg-black/90 flex items-center justify-center p-4"
          onClick={() => setShowTrailer(false)}
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
