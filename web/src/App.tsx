import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Layout } from '@/components/Layout'
import { Dashboard } from '@/pages/Dashboard'
import { Downloads } from '@/pages/Downloads'
import { Library } from '@/pages/Library'
import { Search } from '@/pages/Search'
import { Suggestions } from '@/pages/Suggestions'
import { Settings } from '@/pages/Settings'
import { ErrorBoundary } from '@/components/ErrorBoundary'

function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<Dashboard />} />
            <Route path="downloads" element={<Downloads />} />
            <Route path="library" element={<Library />} />
            <Route path="search" element={<Search />} />
            <Route path="suggestions" element={<Suggestions />} />
            <Route path="settings" element={<Settings />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </ErrorBoundary>
  )
}

export default App