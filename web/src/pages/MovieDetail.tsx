import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Star, Play, ArrowLeft, Film, Heart, Download } from 'lucide-react'
import { useWatchlist } from '@/contexts/WatchlistContext'
import { VideoPlayer } from '@/components/VideoPlayer'

interface CastMember {
  id: number
  name: string
  character: string
  profile_path: string
}

interface CrewMember {
  id: number
  name: string
  job: string
}

interface Video {
  id: string
  name: string
  key: string
  site: string
  type: string
}

interface SimilarMovie {
  id: number
  title: string
  poster_path: string
  vote_average: number
  release_date: string
}

interface MovieDetails {
  id: number
  title: string
  tagline: string
  overview: string
  release_date: string
  runtime: number
  vote_average: number
  vote_count: number
  poster_path: string
  backdrop_path: string
  genres: { id: number; name: string }[]
  credits: {
    cast: CastMember[]
    crew: CrewMember[]
  }
  videos: {
    results: Video[]
  }
  similar: {
    results: SimilarMovie[]
  }
  status: string
  budget: number
  revenue: number
}

export function MovieDetail() {
  const { id } = useParams<{ id: string }>()
  const [movie, setMovie] = useState<MovieDetails | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [showTrailer, setShowTrailer] = useState(false)
  const [showPlayer, setShowPlayer] = useState(false)
  const [localFile, setLocalFile] = useState<string | null>(null)
  const [downloadStatus, setDownloadStatus] = useState<string>('')
  const { addToWatchlist, removeFromWatchlist, isInWatchlist, watchlist } = useWatchlist()
  const [isWatchlistLoading, setIsWatchlistLoading] = useState(false)

  useEffect(() => {
    const fetchMovie = async () => {
      try {
        const response = await fetch(`/api/discover/movie/${id}`)
        if (!response.ok) throw new Error('Failed to fetch movie')
        const data = await response.json()
        setMovie(data)

        // Check if movie has a local file
        const fileResponse = await fetch(`/api/library/movie/${id}/file`)
        if (fileResponse.ok) {
          const fileData = await fileResponse.json()
          if (fileData.file_path) {
            setLocalFile(fileData.file_path)
          }
        }
      } catch (error) {
        console.error('Error fetching movie:', error)
      } finally {
        setIsLoading(false)
      }
    }

    fetchMovie()
  }, [id])

  const handleDownload = async () => {
    if (!movie) return
    setDownloadStatus('searching')
    try {
      const response = await fetch('/api/downloads/search', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          query: movie.title,
          year: movie.release_date?.substring(0, 4),
          media_type: 'movie',
          tmdb_id: movie.id
        })
      })
      if (response.ok) {
        setDownloadStatus('queued')
      } else {
        setDownloadStatus('failed')
      }
    } catch (error) {
      console.error('Download error:', error)
      setDownloadStatus('failed')
    }
  }

  const handleStream = () => {
    if (localFile) {
      setShowPlayer(true)
    }
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-[#121212] flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-[#1ed760]" />
      </div>
    )
  }

  if (!movie) {
    return (
      <div className="min-h-screen bg-[#121212] flex items-center justify-center">
        <p className="text-white">Movie not found</p>
      </div>
    )
  }

  const trailer = movie.videos?.results?.find(v => v.type === 'Trailer' && v.site === 'YouTube')
  const directors = movie.credits?.crew?.filter(c => c.job === 'Director') || []
  const writers = movie.credits?.crew?.filter(c => c.job === 'Writer' || c.job === 'Screenplay') || []
  const inWatchlist = isInWatchlist('movie', movie.id)
  const watchlistItem = watchlist.find(item => item.media_type === 'movie' && item.tmdb_id === movie.id)

  const handleWatchlistToggle = async () => {
    if (isWatchlistLoading) return
    setIsWatchlistLoading(true)
    try {
      if (inWatchlist && watchlistItem) {
        await removeFromWatchlist(watchlistItem.id)
      } else {
        await addToWatchlist({
          media_type: 'movie',
          tmdb_id: movie.id,
          title: movie.title,
          poster_url: movie.poster_path
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
        {movie.backdrop_path ? (
          <div
            className="absolute inset-0 bg-cover bg-center"
            style={{ backgroundImage: `url(https://image.tmdb.org/t/p/original${movie.backdrop_path})` }}
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
            <h1 className="text-5xl font-bold text-white mb-2">{movie.title}</h1>
            {movie.tagline && (
              <p className="text-xl text-[#b3b3b3] italic mb-4">{movie.tagline}</p>
            )}
            
            <div className="flex items-center gap-4 mb-4">
              <div className="flex items-center gap-1 text-yellow-400">
                <Star className="w-5 h-5 fill-current" />
                <span className="font-bold">{movie.vote_average.toFixed(1)}</span>
                <span className="text-[#b3b3b3]">({movie.vote_count.toLocaleString()} votes)</span>
              </div>
              <span className="text-[#b3b3b3]">|</span>
              <span className="text-[#b3b3b3]">{movie.release_date?.substring(0, 4)}</span>
              <span className="text-[#b3b3b3]">|</span>
              <span className="text-[#b3b3b3]">{movie.runtime} min</span>
              <span className="text-[#b3b3b3]">|</span>
              <span className="text-[#b3b3b3]">{movie.status}</span>
            </div>

            <div className="flex gap-2 mb-4">
              {movie.genres?.map(genre => (
                <span key={genre.id} className="px-3 py-1 bg-[#1f1f1f] rounded-full text-sm text-white">
                  {genre.name}
                </span>
              ))}
            </div>

            <div className="flex gap-3">
              {localFile ? (
                <button
                  onClick={handleStream}
                  className="flex items-center gap-2 px-6 py-3 bg-[#1ed760] text-black rounded-full font-semibold hover:bg-[#1ed760]/90 transition-colors"
                >
                  <Play className="w-5 h-5" />
                  Stream Now
                </button>
              ) : (
                <button
                  onClick={handleDownload}
                  disabled={downloadStatus === 'searching' || downloadStatus === 'queued'}
                  className="flex items-center gap-2 px-6 py-3 bg-[#1ed760] text-black rounded-full font-semibold hover:bg-[#1ed760]/90 transition-colors disabled:opacity-50"
                >
                  <Download className="w-5 h-5" />
                  {downloadStatus === 'searching' ? 'Searching...' :
                   downloadStatus === 'queued' ? 'Queued' :
                   downloadStatus === 'failed' ? 'Try Again' :
                   'Download'}
                </button>
              )}
              {trailer && (
                <button
                  onClick={() => setShowTrailer(true)}
                  className="flex items-center gap-2 px-6 py-3 bg-[#1f1f1f] text-white rounded-full font-semibold hover:bg-[#2a2a2a] transition-colors"
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
            </div>
          </motion.div>
        </div>
      </div>

      {/* Main Content */}
      <div className="px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Left column - Overview */}
          <div className="lg:col-span-2">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
            >
              <h2 className="text-2xl font-bold text-white mb-4">Overview</h2>
              <p className="text-[#b3b3b3] text-lg leading-relaxed">{movie.overview}</p>
            </motion.div>

            {/* Cast */}
            {movie.credits?.cast?.length > 0 && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.3 }}
                className="mt-8"
              >
                <h2 className="text-2xl font-bold text-white mb-4">Cast</h2>
                <div className="flex gap-4 overflow-x-auto pb-4">
                  {movie.credits.cast.slice(0, 10).map(actor => (
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
                            <Film className="w-8 h-8 text-[#4d4d4d]" />
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

            {/* Crew */}
            {directors.length > 0 && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.4 }}
                className="mt-8"
              >
                <h2 className="text-2xl font-bold text-white mb-4">Crew</h2>
                <div className="grid grid-cols-2 gap-4">
                  {directors.length > 0 && (
                    <div>
                      <p className="text-[#b3b3b3] text-sm">Director</p>
                      <p className="text-white font-semibold">{directors.map(d => d.name).join(', ')}</p>
                    </div>
                  )}
                  {writers.length > 0 && (
                    <div>
                      <p className="text-[#b3b3b3] text-sm">Writers</p>
                      <p className="text-white font-semibold">{writers.map(w => w.name).join(', ')}</p>
                    </div>
                  )}
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
              <h3 className="text-lg font-bold text-white mb-4">Movie Info</h3>
              <div className="space-y-3">
                <div>
                  <p className="text-[#b3b3b3] text-sm">Release Date</p>
                  <p className="text-white">{movie.release_date}</p>
                </div>
                <div>
                  <p className="text-[#b3b3b3] text-sm">Runtime</p>
                  <p className="text-white">{Math.floor(movie.runtime / 60)}h {movie.runtime % 60}m</p>
                </div>
                <div>
                  <p className="text-[#b3b3b3] text-sm">Budget</p>
                  <p className="text-white">${(movie.budget / 1000000).toFixed(1)}M</p>
                </div>
                <div>
                  <p className="text-[#b3b3b3] text-sm">Revenue</p>
                  <p className="text-white">${(movie.revenue / 1000000).toFixed(1)}M</p>
                </div>
                <div>
                  <p className="text-[#b3b3b3] text-sm">Status</p>
                  <p className="text-white">{movie.status}</p>
                </div>
              </div>
            </motion.div>
          </div>
        </div>

        {/* Similar Movies */}
        {movie.similar?.results?.length > 0 && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5 }}
            className="mt-8"
          >
            <h2 className="text-2xl font-bold text-white mb-4">Similar Movies</h2>
            <div className="flex gap-4 overflow-x-auto pb-4">
              {movie.similar.results.slice(0, 10).map(similar => (
                <Link
                  key={similar.id}
                  to={`/movie/${similar.id}`}
                  className="flex-shrink-0 w-[160px] group"
                >
                  <div className="aspect-[2/3] rounded-lg overflow-hidden bg-[#1f1f1f]">
                    {similar.poster_path ? (
                      <img
                        src={`https://image.tmdb.org/t/p/w500${similar.poster_path}`}
                        alt={similar.title}
                        className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-110"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center">
                        <Film className="w-8 h-8 text-[#4d4d4d]" />
                      </div>
                    )}
                  </div>
                  <p className="mt-2 text-sm font-semibold text-white truncate group-hover:text-[#1ed760]">
                    {similar.title}
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

      {/* Video Player Modal */}
      {showPlayer && localFile && (
        <div className="fixed inset-0 z-50 bg-black">
          <VideoPlayer
            src={`/api/stream/direct?path=${encodeURIComponent(localFile)}`}
            poster={movie.backdrop_path ? `https://image.tmdb.org/t/p/original${movie.backdrop_path}` : undefined}
            title={movie.title}
            onClose={() => setShowPlayer(false)}
            autoPlay
            mediaType="movie"
            mediaId={movie.id}
            subtitlePath={localFile}
          />
        </div>
      )}
    </div>
  )
}
