import { useState, useEffect, useRef, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { Play, Star, TrendingUp, Film, Tv, Clock, Settings as SettingsIcon, Search, Home, Heart, History } from 'lucide-react'
import { Link, useNavigate } from 'react-router-dom'

interface TVItem {
  id: number
  title?: string
  name?: string
  poster_path: string
  backdrop_path: string
  vote_average: number
  release_date?: string
  first_air_date?: string
  media_type: string
}

interface TVSection {
  title: string
  endpoint: string
  icon: React.ReactNode
}

const sections: TVSection[] = [
  { title: 'Trending Movies', endpoint: '/api/discover/movies/trending', icon: <TrendingUp className="w-8 h-8" /> },
  { title: 'Popular Movies', endpoint: '/api/discover/movies/popular', icon: <Star className="w-8 h-8" /> },
  { title: 'Now Playing', endpoint: '/api/discover/movies/now_playing', icon: <Play className="w-8 h-8" /> },
  { title: 'Trending TV', endpoint: '/api/discover/tv/trending', icon: <Tv className="w-8 h-8" /> },
  { title: 'Popular TV', endpoint: '/api/discover/tv/popular', icon: <Star className="w-8 h-8" /> },
  { title: 'Airing Today', endpoint: '/api/discover/tv/airing_today', icon: <Clock className="w-8 h-8" /> },
]

// Navigation items for sidebar
const navItems = [
  { id: 'home', label: 'Home', icon: <Home className="w-10 h-10" />, path: '/tv' },
  { id: 'discover', label: 'Discover', icon: <Search className="w-10 h-10" />, path: '/tv/discover' },
  { id: 'watchlist', label: 'Watchlist', icon: <Heart className="w-10 h-10" />, path: '/tv/watchlist' },
  { id: 'history', label: 'History', icon: <History className="w-10 h-10" />, path: '/tv/history' },
  { id: 'settings', label: 'Settings', icon: <SettingsIcon className="w-10 h-10" />, path: '/tv/settings' },
]

function TVRow({ section, isFocused, focusedIndex }: { section: TVSection; isFocused: boolean; focusedIndex: number }) {
  const { data, isLoading } = useQuery({
    queryKey: ['tv-discover', section.endpoint],
    queryFn: async () => {
      const response = await fetch(section.endpoint)
      if (!response.ok) throw new Error('Failed to fetch')
      return response.json()
    },
    staleTime: 1000 * 60 * 5,
  })

  const items: TVItem[] = data?.results?.slice(0, 8) || []

  if (isLoading) {
    return (
      <div className="mb-12">
        <div className="h-10 w-64 bg-white/10 rounded animate-pulse mb-6" />
        <div className="flex gap-6">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="flex-shrink-0 w-[280px]">
              <div className="aspect-[2/3] bg-white/10 rounded-xl animate-pulse" />
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="mb-12">
      <div className="flex items-center gap-4 mb-6 px-12">
        <span className="text-[#1ed760]">{section.icon}</span>
        <h2 className="text-3xl font-bold text-white">{section.title}</h2>
      </div>

      <div className="flex gap-6 overflow-x-auto pb-6 px-12 scrollbar-hide">
        {items.map((item, index) => {
          const isItemFocused = isFocused && index === focusedIndex
          return (
            <Link
              key={item.id}
              to={`/${item.media_type === 'tv' ? 'tv' : 'movie'}/${item.id}`}
              className={`flex-shrink-0 w-[280px] transition-all duration-300 ${
                isItemFocused ? 'scale-110 z-10' : 'scale-100'
              }`}
            >
              <div className={`relative aspect-[2/3] rounded-xl overflow-hidden bg-[#1f1f1f] transition-all ${
                isItemFocused ? 'ring-4 ring-[#1ed760] shadow-2xl shadow-[#1ed760]/20' : ''
              }`}>
                {item.poster_path ? (
                  <img
                    src={`https://image.tmdb.org/t/p/w500${item.poster_path}`}
                    alt={item.title || item.name}
                    className="w-full h-full object-cover"
                    loading="lazy"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center">
                    <Film className="w-20 h-20 text-[#4d4d4d]" />
                  </div>
                )}
                {isItemFocused && (
                  <div className="absolute inset-0 bg-gradient-to-t from-black/90 via-transparent to-transparent">
                    <div className="absolute bottom-4 left-4 right-4">
                      <div className="flex items-center gap-2 text-yellow-400 mb-2">
                        <Star className="w-6 h-6 fill-current" />
                        <span className="text-xl font-bold">{item.vote_average.toFixed(1)}</span>
                      </div>
                    </div>
                  </div>
                )}
              </div>
              <h3 className={`mt-3 text-xl font-semibold truncate transition-colors ${
                isItemFocused ? 'text-[#1ed760]' : 'text-white'
              }`}>
                {item.title || item.name}
              </h3>
              <p className="text-lg text-[#b3b3b3]">
                {item.release_date?.substring(0, 4) || item.first_air_date?.substring(0, 4) || 'N/A'}
              </p>
            </Link>
          )
        })}
      </div>
    </div>
  )
}

export function TVInterface() {
  const [focusedRow, setFocusedRow] = useState(0)
  const [focusedItem, setFocusedItem] = useState(0)
  const [focusedNav, setFocusedNav] = useState(0)
  const [navOpen, setNavOpen] = useState(true)
  const containerRef = useRef<HTMLDivElement>(null)
  const navigate = useNavigate()

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    // D-pad / Arrow keys navigation
    switch (e.key) {
      case 'ArrowUp':
      case 'Up':
        e.preventDefault()
        if (navOpen) {
          setFocusedNav(prev => Math.max(0, prev - 1))
        } else {
          setFocusedRow(prev => Math.max(0, prev - 1))
          setFocusedItem(0)
        }
        break
      case 'ArrowDown':
      case 'Down':
        e.preventDefault()
        if (navOpen) {
          setFocusedNav(prev => Math.min(navItems.length - 1, prev + 1))
        } else {
          setFocusedRow(prev => Math.min(sections.length - 1, prev + 1))
          setFocusedItem(0)
        }
        break
      case 'ArrowLeft':
      case 'Left':
        e.preventDefault()
        if (!navOpen) {
          setFocusedItem(prev => Math.max(0, prev - 1))
        }
        break
      case 'ArrowRight':
      case 'Right':
        e.preventDefault()
        if (navOpen) {
          setNavOpen(false)
          setFocusedItem(0)
        } else {
          setFocusedItem(prev => Math.min(7, prev + 1))
        }
        break
      case 'Enter':
      case 'Select':
        e.preventDefault()
        if (navOpen) {
          const item = navItems[focusedNav]
          if (item.path) {
            navigate(item.path)
          }
        } else {
          // Open the focused item
          // Navigate to the selected movie/show
        }
        break
      case 'Back':
      case 'Escape':
        e.preventDefault()
        if (!navOpen) {
          setNavOpen(true)
        }
        break
      case 'Home':
        e.preventDefault()
        setNavOpen(true)
        setFocusedNav(0)
        break
    }
  }, [focusedRow, focusedItem, focusedNav, navOpen, navigate])

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  // Auto-hide nav after inactivity
  useEffect(() => {
    if (!navOpen) {
      const timer = setTimeout(() => setNavOpen(true), 30000)
      return () => clearTimeout(timer)
    }
  }, [navOpen, focusedRow, focusedItem])

  return (
    <div ref={containerRef} className="min-h-screen bg-[#0a0a0a] text-white overflow-hidden">
      {/* Side Navigation */}
      <motion.div
        className="fixed left-0 top-0 h-full bg-[#121212]/95 backdrop-blur-md z-50 flex flex-col py-12"
        animate={{ width: navOpen ? 280 : 80 }}
        transition={{ duration: 0.3 }}
      >
        {/* Logo */}
        <div className="px-8 mb-12">
          <motion.div
            className="flex items-center gap-4"
            animate={{ justifyContent: navOpen ? 'flex-start' : 'center' }}
          >
            <div className="w-14 h-14 bg-[#1ed760] rounded-2xl flex items-center justify-center flex-shrink-0">
              <Play className="w-8 h-8 text-black ml-1" fill="currentColor" />
            </div>
            {navOpen && (
              <motion.span
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="text-2xl font-bold"
              >
                Media
              </motion.span>
            )}
          </motion.div>
        </div>

        {/* Nav Items */}
        <nav className="flex-1 flex flex-col gap-3 px-4">
          {navItems.map((item, index) => {
            const isFocused = navOpen && focusedNav === index
            return (
              <button
                key={item.id}
                onClick={() => {
                  setFocusedNav(index)
                  navigate(item.path)
                }}
                className={`flex items-center gap-5 px-6 py-5 rounded-2xl transition-all duration-200 ${
                  isFocused
                    ? 'bg-[#1ed760] text-black scale-105'
                    : 'text-white/70 hover:bg-white/10 hover:text-white'
                }`}
              >
                <span className="flex-shrink-0">{item.icon}</span>
                {navOpen && (
                  <motion.span
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    className="text-xl font-semibold whitespace-nowrap"
                  >
                    {item.label}
                  </motion.span>
                )}
              </button>
            )
          })}
        </nav>

        {/* User Info */}
        {navOpen && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="px-8 py-6 border-t border-white/10"
          >
            <div className="flex items-center gap-4">
              <div className="w-12 h-12 bg-[#1f1f1f] rounded-full flex items-center justify-center">
                <span className="text-xl font-bold">U</span>
              </div>
              <div>
                <p className="text-lg font-semibold">User</p>
                <p className="text-sm text-[#b3b3b3]">TV Mode</p>
              </div>
            </div>
          </motion.div>
        )}
      </motion.div>

      {/* Main Content */}
      <motion.div
        className="min-h-screen"
        animate={{ marginLeft: navOpen ? 280 : 80 }}
        transition={{ duration: 0.3 }}
      >
        {/* Hero Section */}
        <div className="relative h-[60vh] bg-gradient-to-b from-[#1a1a2e] to-[#0a0a0a]">
          <div className="absolute inset-0 bg-[url('https://image.tmdb.org/t/p/original/wwemzKWzjKYJFfCniXl0DKJ8f7x.jpg')] bg-cover bg-center opacity-40" />
          <div className="absolute inset-0 bg-gradient-to-t from-[#0a0a0a] via-transparent to-transparent" />
          <div className="relative h-full flex flex-col justify-end pb-12 px-12">
            <motion.h1
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="text-7xl font-bold text-white mb-4"
            >
              Discover
            </motion.h1>
            <motion.p
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.1 }}
              className="text-2xl text-[#b3b3b3]"
            >
              Press D-pad to navigate, Enter to select
            </motion.p>
          </div>
        </div>

        {/* Content Sections */}
        <div className="py-8">
          {sections.map((section, index) => (
            <TVRow
              key={section.endpoint}
              section={section}
              isFocused={!navOpen && focusedRow === index}
              focusedIndex={focusedItem}
            />
          ))}
        </div>
      </motion.div>

      {/* Focus Indicator */}
      {!navOpen && (
        <div className="fixed bottom-8 right-8 bg-[#1ed760] text-black px-6 py-3 rounded-full text-xl font-bold">
          Row {focusedRow + 1} / {sections.length}
        </div>
      )}
    </div>
  )
}
