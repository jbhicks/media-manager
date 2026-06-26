import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '@/contexts/AuthContext'
import { WatchlistProvider } from '@/contexts/WatchlistContext'
import { WatchHistoryProvider } from '@/contexts/WatchHistoryContext'
import { Layout } from '@/components/Layout'
import { Dashboard } from '@/pages/Dashboard'
import { Downloads } from '@/pages/Downloads'
import { Library } from '@/pages/Library'
import { Search } from '@/pages/Search'
import { Suggestions } from '@/pages/Suggestions'
import { Settings } from '@/pages/Settings'
import { Discover } from '@/pages/Discover'
import { MovieDetail } from '@/pages/MovieDetail'
import { TVDetail } from '@/pages/TVDetail'
import { Login } from '@/pages/Login'
import { Watchlist } from '@/pages/Watchlist'
import { TVInterface } from '@/pages/TVInterface'
import { ErrorBoundary } from '@/components/ErrorBoundary'

function App() {
  return (
    <ErrorBoundary>
      <AuthProvider>
        <WatchlistProvider>
          <WatchHistoryProvider>
            <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
              <Routes>
                <Route path="/login" element={<Login />} />
                <Route path="/" element={<Layout />}>
                  <Route index element={<Dashboard />} />
                  <Route path="downloads" element={<Downloads />} />
                  <Route path="library" element={<Library />} />
                  <Route path="search" element={<Search />} />
                  <Route path="suggestions" element={<Suggestions />} />
                  <Route path="discover" element={<Discover />} />
                  <Route path="watchlist" element={<Watchlist />} />
                  <Route path="movie/:id" element={<MovieDetail />} />
                  <Route path="tv/:id" element={<TVDetail />} />
                  <Route path="tv" element={<TVInterface />} />
                  <Route path="settings" element={<Settings />} />
                </Route>
              </Routes>
            </BrowserRouter>
          </WatchHistoryProvider>
        </WatchlistProvider>
      </AuthProvider>
    </ErrorBoundary>
  )
}

export default App