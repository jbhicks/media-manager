import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { motion, AnimatePresence } from 'framer-motion'
import { Play, Star, TrendingUp, Calendar, Clock, ChevronRight, Film, Tv } from 'lucide-react'
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

function DiscoverRow({ section }: { section: DiscoverSection }) {
  const { data, isLoading } = useQuery({
    queryKey: ['discover', section.endpoint],
    queryFn: async () => {
      const response = await fetch(section.endpoint)
      if (!response.ok) throw new Error('Failed to fetch')
      return response.json()
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
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

  const items: DiscoverItem[] = data?.results?.slice(0, 10) || []

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

  const filteredSections = sections.filter(section => {
    if (activeTab === 'all') return true
    if (activeTab === 'movies') return section.endpoint.includes('/movies/')
    if (activeTab === 'tv') return section.endpoint.includes('/tv/')
    return true
  })

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

      {/* Tab Navigation */}
      <div className="sticky top-0 z-40 bg-[#121212]/95 backdrop-blur-md border-b border-[#1f1f1f] px-4 py-3">
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
      </div>

      {/* Content Sections */}
      <div className="py-6">
        <AnimatePresence mode="wait">
          <motion.div
            key={activeTab}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.2 }}
          >
            {filteredSections.map((section) => (
              <DiscoverRow key={section.endpoint} section={section} />
            ))}
          </motion.div>
        </AnimatePresence>
      </div>
    </div>
  )
}
