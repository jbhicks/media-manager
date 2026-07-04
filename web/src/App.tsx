import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '@/contexts/AuthContext'
import { WatchlistProvider } from '@/contexts/WatchlistContext'
import { WatchHistoryProvider } from '@/contexts/WatchHistoryContext'
import { Layout } from '@/components/Layout'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import { Dashboard } from '@/pages/Dashboard'
import { Downloads } from '@/pages/Downloads'
import { Library } from '@/pages/Library'
import { Search } from '@/pages/Search'
import { Suggestions } from '@/pages/Suggestions'
import { Settings } from '@/pages/Settings'
import { Discover } from '@/pages/Discover'
import { DiscoverSection } from '@/pages/DiscoverSection'
import { MovieDetail } from '@/pages/MovieDetail'
import { TVDetail } from '@/pages/TVDetail'
import { Login } from '@/pages/Login'
import { Watchlist } from '@/pages/Watchlist'
import { TVInterface } from '@/pages/TVInterface'
import { TVGuide } from '@/pages/TVGuide'
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
                <Route
                  path="/movie/:id"
                  element={
                    <ProtectedRoute>
                      <MovieDetail />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/tv/:id"
                  element={
                    <ProtectedRoute>
                      <TVDetail />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/tv"
                  element={
                    <ProtectedRoute>
                      <TVInterface />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/"
                  element={
                    <ProtectedRoute>
                      <Layout />
                    </ProtectedRoute>
                  }
                >
                  <Route index element={<Dashboard />} />
                  <Route path="downloads" element={<Downloads />} />
                  <Route path="library" element={<Library />} />
                  <Route path="search" element={<Search />} />
                  <Route path="suggestions" element={<Suggestions />} />
                  <Route path="discover" element={<Discover />} />
                  <Route path="discover/:type/:category" element={<DiscoverSection />} />
                  <Route path="tv-guide" element={<TVGuide />} />
                  <Route path="watchlist" element={<Watchlist />} />
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