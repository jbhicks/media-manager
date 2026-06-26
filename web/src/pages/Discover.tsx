import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { motion, AnimatePresence } from 'framer-motion'
import { Play, Star, TrendingUp, Calendar, Clock, ChevronRight, Film, Tv, X, SlidersHorizontal } from 'lucide-react'
import { Link } from 'react-router-dom'

interface DiscoverItem {
  id: number
  title?: string
  name?: string
  poster_path: string
  backdrop_path: string
  overview: string
  release_date?: string
  first_air_date?: string
  vote_average: number
  vote_count: number
  media_type: string
  genre_ids?: number[]
}

interface DiscoverSection {
  title: string
  endpoint: string
  icon: React.ReactNode
}

const sections: DiscoverSection[] = [
  { title: 'Trending Movies', endpoint: '/api/discover/movies/trending', icon: <TrendingUp className="w-5 h-5" /> },
  { title: 'Popular Movies', endpoint: '/api/discover/movies/popular', icon: <Star className="w-5 h-5" /> },
  { title: 'Now Playing', endpoint: '/api/discover/movies/now_playing', icon: <Play className="w-5 h-5" /> },
  { title: 'Upcoming Movies', endpoint: '/api/discover/movies/upcoming', icon: <Calendar className="w-5 h-5" /> },
  { title: 'Top Rated Movies', endpoint: '/api/discover/movies/top_rated', icon: <Star className="w-5 h-5" /> },
  { title: 'Trending TV Shows', endpoint: '/api/discover/tv/trending', icon: <TrendingUp className="w-5 h-5" /> },
  { title: 'Popular TV Shows', endpoint: '/api/discover/tv/popular', icon: <Tv className="w-5 h-5" /> },
  { title: 'Airing Today', endpoint: '/api/discover/tv/airing_today', icon: <Clock className="w-5 h-5" /> },
  { title: 'Top Rated TV', endpoint: '/api/discover/tv/top_rated', icon: <Star className="w-5 h-5" /> },
]

// Genre definitions
const movieGenres = [
  { id: 28, name: 'Action' },
  { id: 12, name: 'Adventure' },
  { id: 16, name: 'Animation' },
  { id: 35, name: 'Comedy' },
  { id: 80, name: 'Crime' },
  { id: 99, name: 'Documentary' },
  { id: 18, name: 'Drama' },
  { id: 10751, name: 'Family' },
  { id: 14, name: 'Fantasy' },
  { id: 36, name: 'History' },
  { id: 27, name: 'Horror' },
  { id: 10402, name: 'Music' },
  { id: 9648, name: 'Mystery' },
  { id: 10749, name: 'Romance' },
  { id: 878, name: 'Science Fiction' },
  { id: 10770, name: 'TV Movie' },
  { id: 53, name: 'Thriller' },
  { id: 10752, name: 'War' },
  { id: 37, name: 'Western' },
]

const tvGenres = [
  { id: 10759, name: 'Action & Adventure' },
  { id: 16, name: 'Animation' },
  { id: 35, name: 'Comedy' },
  { id: 80, name: 'Crime' },
  { id: 99, name: 'Documentary' },
  { id: 18, name: 'Drama' },
  { id: 10751, name: 'Family' },
  { id: 10762, name: 'Kids' },
  { id: 9648, name: 'Mystery' },
  { id: 10763, name: 'News' },
  { id: 10764, name: 'Reality' },
  { id: 10765, name: 'Sci-Fi & Fantasy' },
  { id: 10766, name: 'Soap' },
  { id: 10767, name: 'Talk' },
  { id: 10768, name: 'War & Politics' },
  { id: 37, name: 'Western' },
]

interface Filters {
  genre: number | null
  minYear: number | null
  maxYear: number | null
  minRating: number | null
}

function DiscoverRow({ section, filters }: { section: DiscoverSection; filters: Filters }) {
  const { data, isLoading } = useQuery({
    queryKey: ['discover', section.endpoint, filters],
    queryFn: async () => {
      const response = await fetch(section.endpoint)
      if (!response.ok) throw new Error('Failed to fetch')
      return response.json()
    },
    staleTime: 1000 * 60 * 5,
  })

  if (isLoading) {
    return (
      <div className="mb-8">
        <div className="flex items-center gap-2 mb-4">
          <div className="w-5 h-5 bg-[#1f1f1f] rounded animate-pulse" />
          <div className="h-6 w-48 bg-[#1f1f1f] rounded animate-pulse" />
        </div>
        <div className="flex gap-4 overflow-hidden">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="flex-shrink-0 w-[180px]">
              <div className="aspect-[2/3] bg-[#1f1f1f] rounded-lg animate-pulse" />
              <div className="h-4 w-3/4 bg-[#1f1f1f] rounded mt-2 animate-pulse" />
            </div>
          ))}
        </div>
      </div>
    )
  }

  let items: DiscoverItem[] = data?.results?.slice(0, 10) || []

  // Apply filters
  if (filters.genre) {
    items = items.filter(item => item.genre_ids?.includes(filters.genre!))
  }
  if (filters.minYear) {
    items = items.filter(item => {
      const year = parseInt((item.release_date || item.first_air_date || '').substring(0, 4))
      return !isNaN(year) && year >= filters.minYear!
    })
  }
  if (filters.maxYear) {
    items = items.filter(item => {
      const year = parseInt((item.release_date || item.first_air_date || '').substring(0, 4))
      return !isNaN(year) && year <= filters.maxYear!
    })
  }
  if (filters.minRating) {
    items = items.filter(item => item.vote_average >= filters.minRating!)
  }

  if (items.length === 0) {
    return (
      <div className="mb-8">
        <div className="flex items-center gap-2 mb-4 px-4">
          <span className="text-[#1ed760]">{section.icon}</span>
          <h2 className="text-xl font-bold text-white">{section.title}</h2>
        </div>
        <p className="text-[#b3b3b3] px-4">No items match the current filters</p>
      </div>
    )
  }

  return (
    <div className="mb-8">
      <div className="flex items-center gap-2 mb-4 px-4">
        <span className="text-[#1ed760]">{section.icon}</span>
        <h2 className="text-xl font-bold text-white">{section.title}</h2>
        <ChevronRight className="w-5 h-5 text-[#b3b3b3] ml-auto" />
      </div>
      
      <div className="flex gap-4 overflow-x-auto pb-4 px-4 scrollbar-hide">
        {items.map((item) => (
          <Link
            key={item.id}
            to={`/${item.media_type === 'tv' ? 'tv' : 'movie'}/${item.id}`}
            className="flex-shrink-0 w-[180px] group cursor-pointer"
          >
            <div className="relative aspect-[2/3] rounded-lg overflow-hidden bg-[#1f1f1f]">
              {item.poster_path ? (
                <img
                  src={`https://image.tmdb.org/t/p/w500${item.poster_path}`}
                  alt={item.title || item.name}
                  className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-110"
                  loading="lazy"
                />
              ) : (
                <div className="w-full h-full flex items-center justify-center">
                  <Film className="w-12 h-12 text-[#4d4d4d]" />
                </div>
              )}
              <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                <div className="absolute bottom-2 left-2 right-2">
                  <div className="flex items-center gap-1 text-yellow-400">
                    <Star className="w-4 h-4 fill-current" />
                    <span className="text-sm font-bold">{item.vote_average.toFixed(1)}</span>
                  </div>
                </div>
              </div>
            </div>
            <h3 className="mt-2 text-sm font-semibold text-white truncate group-hover:text-[#1ed760] transition-colors">
              {item.title || item.name}
            </h3>
            <p className="text-xs text-[#b3b3b3]">
              {item.release_date?.substring(0, 4) || item.first_air_date?.substring(0, 4) || 'N/A'}
            </p>
          </Link>
        ))}
      </div>
    </div>
  )
}

export function Discover() {
  const [activeTab, setActiveTab] = useState<'all' | 'movies' | 'tv'>('all')
  const [showFilters, setShowFilters] = useState(false)
  const [filters, setFilters] = useState<Filters>({
    genre: null,
    minYear: null,
    maxYear: null,
    minRating: null,
  })

  const currentYear = new Date().getFullYear()
  const years = Array.from({ length: 50 }, (_, i) => currentYear - i)
  const ratings = [1, 2, 3, 4, 5, 6, 7, 8, 9]

  const activeGenres = activeTab === 'tv' ? tvGenres : activeTab === 'movies' ? movieGenres : [...movieGenres, ...tvGenres]

  const filteredSections = sections.filter(section => {
    if (activeTab === 'all') return true
    if (activeTab === 'movies') return section.endpoint.includes('/movies/')
    if (activeTab === 'tv') return section.endpoint.includes('/tv/')
    return true
  })

  const hasActiveFilters = filters.genre || filters.minYear || filters.maxYear || filters.minRating

  const clearFilters = () => {
    setFilters({ genre: null, minYear: null, maxYear: null, minRating: null })
  }

  return (
    <div className="min-h-screen bg-[#121212]">
      {/* Hero Section */}
      <div className="relative h-[50vh] bg-gradient-to-b from-[#1a1a2e] to-[#121212]">
        <div className="absolute inset-0 bg-[url('https://image.tmdb.org/t/p/original/wwemzKWzjKYJFfCniXl0DKJ8f7x.jpg')] bg-cover bg-center opacity-30" />
        <div className="absolute inset-0 bg-gradient-to-t from-[#121212] via-transparent to-transparent" />
        <div className="relative h-full flex flex-col justify-end pb-8 px-8">
          <motion.h1
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="text-5xl font-bold text-white mb-2"
          >
            Discover
          </motion.h1>
          <motion.p
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
            className="text-[#b3b3b3] text-lg"
          >
            Explore trending movies and TV shows
          </motion.p>
        </div>
      </div>

      {/* Tab Navigation + Filters */}
      <div className="sticky top-0 z-40 bg-[#121212]/95 backdrop-blur-md border-b border-[#1f1f1f] px-4 py-3">
        <div className="flex items-center justify-between">
          <div className="flex gap-2">
            {(['all', 'movies', 'tv'] as const).map((tab) => (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                className={`px-6 py-2 rounded-full text-sm font-semibold uppercase tracking-wider transition-all duration-300 ${
                  activeTab === tab
                    ? 'bg-white text-black'
                    : 'bg-[#1f1f1f] text-white hover:bg-[#2a2a2a]'
                }`}
              >
                {tab === 'all' ? 'All' : tab === 'movies' ? 'Movies' : 'TV Shows'}
              </button>
            ))}
          </div>
          <div className="flex items-center gap-2">
            {hasActiveFilters && (
              <button
                onClick={clearFilters}
                className="flex items-center gap-1 px-3 py-2 bg-red-500/20 text-red-400 rounded-full text-sm hover:bg-red-500/30 transition-colors"
              >
                <X className="w-4 h-4" />
                Clear
              </button>
            )}
            <button
              onClick={() => setShowFilters(!showFilters)}
              className={`flex items-center gap-2 px-4 py-2 rounded-full text-sm font-semibold transition-colors ${
                showFilters || hasActiveFilters
                  ? 'bg-[#1ed760] text-black'
                  : 'bg-[#1f1f1f] text-white hover:bg-[#2a2a2a]'
              }`}
            >
              <SlidersHorizontal className="w-4 h-4" />
              Filters
              {hasActiveFilters && (
                <span className="w-5 h-5 bg-black text-[#1ed760] rounded-full text-xs flex items-center justify-center font-bold">
                  {[filters.genre, filters.minYear, filters.maxYear, filters.minRating].filter(Boolean).length}
                </span>
              )}
            </button>
          </div>
        </div>

        {/* Filter Panel */}
        <AnimatePresence>
          {showFilters && (
            <motion.div
              initial={{ height: 0, opacity: 0 }}
              animate={{ height: 'auto', opacity: 1 }}
              exit={{ height: 0, opacity: 0 }}
              transition={{ duration: 0.2 }}
              className="overflow-hidden"
            >
              <div className="pt-4 pb-2 grid grid-cols-1 md:grid-cols-4 gap-4">
                {/* Genre Filter */}
                <div>
                  <label className="text-[#b3b3b3] text-sm mb-2 block">Genre</label>
                  <select
                    value={filters.genre || ''}
                    onChange={(e) => setFilters({ ...filters, genre: e.target.value ? parseInt(e.target.value) : null })}
                    className="w-full bg-[#1f1f1f] text-white rounded-lg px-3 py-2 border border-[#2a2a2a] focus:border-[#1ed760] focus:outline-none"
                  >
                    <option value="">All Genres</option>
                    {activeGenres.map(genre => (
                      <option key={genre.id} value={genre.id}>{genre.name}</option>
                    ))}
                  </select>
                </div>

                {/* Min Year */}
                <div>
                  <label className="text-[#b3b3b3] text-sm mb-2 block">From Year</label>
                  <select
                    value={filters.minYear || ''}
                    onChange={(e) => setFilters({ ...filters, minYear: e.target.value ? parseInt(e.target.value) : null })}
                    className="w-full bg-[#1f1f1f] text-white rounded-lg px-3 py-2 border border-[#2a2a2a] focus:border-[#1ed760] focus:outline-none"
                  >
                    <option value="">Any</option>
                    {years.map(year => (
                      <option key={year} value={year}>{year}</option>
                    ))}
                  </select>
                </div>

                {/* Max Year */}
                <div>
                  <label className="text-[#b3b3b3] text-sm mb-2 block">To Year</label>
                  <select
                    value={filters.maxYear || ''}
                    onChange={(e) => setFilters({ ...filters, maxYear: e.target.value ? parseInt(e.target.value) : null })}
                    className="w-full bg-[#1f1f1f] text-white rounded-lg px-3 py-2 border border-[#2a2a2a] focus:border-[#1ed760] focus:outline-none"
                  >
                    <option value="">Any</option>
                    {years.map(year => (
                      <option key={year} value={year}>{year}</option>
                    ))}
                  </select>
                </div>

                {/* Min Rating */}
                <div>
                  <label className="text-[#b3b3b3] text-sm mb-2 block">Min Rating</label>
                  <select
                    value={filters.minRating || ''}
                    onChange={(e) => setFilters({ ...filters, minRating: e.target.value ? parseInt(e.target.value) : null })}
                    className="w-full bg-[#1f1f1f] text-white rounded-lg px-3 py-2 border border-[#2a2a2a] focus:border-[#1ed760] focus:outline-none"
                  >
                    <option value="">Any</option>
                    {ratings.map(rating => (
                      <option key={rating} value={rating}>{rating}+ Stars</option>
                    ))}
                  </select>
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* Content Sections */}
      <div className="py-6">
        <AnimatePresence mode="wait">
          <motion.div
            key={activeTab + JSON.stringify(filters)}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.2 }}
          >
            {filteredSections.map((section) => (
              <DiscoverRow key={section.endpoint} section={section} filters={filters} />
            ))}
          </motion.div>
        </AnimatePresence>
      </div>
    </div>
  )
}
