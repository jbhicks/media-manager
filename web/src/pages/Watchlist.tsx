import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Heart, Trash2, Film, Tv } from 'lucide-react'
import { useWatchlist } from '@/contexts/WatchlistContext'

export function Watchlist() {
  const { watchlist, isLoading, error, removeFromWatchlist, refreshWatchlist } = useWatchlist()

  useEffect(() => {
    refreshWatchlist()
  }, [refreshWatchlist])

  if (isLoading) {
    return (
      <div className="min-h-screen bg-[#121212] flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-[#1ed760]" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen bg-[#121212] flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-400 mb-4">{error}</p>
          <button 
            onClick={refreshWatchlist}
            className="px-4 py-2 bg-[#1ed760] text-black rounded-full font-semibold hover:bg-[#1ed760]/90"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-[#121212]">
      {/* Header */}
      <div className="relative h-[30vh] bg-gradient-to-b from-[#1a1a2e] to-[#121212]">
        <div className="absolute inset-0 bg-gradient-to-t from-[#121212] via-transparent to-transparent" />
        <div className="relative h-full flex flex-col justify-end pb-8 px-8">
          <motion.h1
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="text-4xl font-bold text-white mb-2"
          >
            My Watchlist
          </motion.h1>
          <motion.p
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
            className="text-[#b3b3b3] text-lg"
          >
            {watchlist.length} {watchlist.length === 1 ? 'item' : 'items'} saved
          </motion.p>
        </div>
      </div>

      {/* Content */}
      <div className="px-8 py-6">
        {watchlist.length === 0 ? (
          <div className="text-center py-20">
            <Heart className="w-16 h-16 text-[#4d4d4d] mx-auto mb-4" />
            <h2 className="text-2xl font-bold text-white mb-2">Your watchlist is empty</h2>
            <p className="text-[#b3b3b3] mb-6">Start adding movies and TV shows you want to watch!</p>
            <Link 
              to="/discover"
              className="inline-flex items-center gap-2 px-6 py-3 bg-[#1ed760] text-black rounded-full font-semibold hover:bg-[#1ed760]/90 transition-colors"
            >
              <Film className="w-5 h-5" />
              Discover Content
            </Link>
          </div>
        ) : (
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6">
            {watchlist.map((item, index) => (
              <motion.div
                key={item.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.05 }}
                className="group relative"
              >
                <Link to={`/${item.media_type === 'movie' ? 'movie' : 'tv'}/${item.tmdb_id}`}>
                  <div className="aspect-[2/3] rounded-lg overflow-hidden bg-[#1f1f1f] relative">
                    {item.poster_url ? (
                      <img
                        src={`https://image.tmdb.org/t/p/w500${item.poster_url}`}
                        alt={item.title}
                        className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center">
                        <Film className="w-12 h-12 text-[#4d4d4d]" />
                      </div>
                    )}
                    {/* Media type badge */}
                    <div className="absolute top-2 right-2 px-2 py-1 bg-black/70 rounded-full text-xs text-white">
                      {item.media_type === 'movie' ? (
                        <Film className="w-3 h-3" />
                      ) : (
                        <Tv className="w-3 h-3" />
                      )}
                    </div>
                  </div>
                  <h3 className="mt-2 text-sm font-semibold text-white truncate group-hover:text-[#1ed760] transition-colors">
                    {item.title}
                  </h3>
                  <p className="text-xs text-[#b3b3b3]">
                    Added {new Date(item.added_at).toLocaleDateString()}
                  </p>
                </Link>
                {/* Remove button */}
                <button
                  onClick={() => removeFromWatchlist(item.id)}
                  className="absolute top-2 left-2 p-2 rounded-full bg-black/70 text-white opacity-0 group-hover:opacity-100 transition-opacity hover:bg-red-600"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </motion.div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
