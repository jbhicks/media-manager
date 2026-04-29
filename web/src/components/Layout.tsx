import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Header } from './Header'
import { Toaster } from './Toaster'
import { VPNWarning } from './VPNWarning'
import { useKeyboardShortcuts } from '@/hooks/useKeyboardShortcuts'

export function Layout() {
  useKeyboardShortcuts()

  return (
    <div className="min-h-screen bg-background">
      <Header />
      <div className="flex">
        <Sidebar />
        <main className="flex-1 p-6">
          <VPNWarning />
          <Outlet />
        </main>
      </div>
      <Toaster />
    </div>
  )
}