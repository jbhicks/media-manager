import { Menu, Moon, Sun, Bell } from 'lucide-react'
import { useAppStore } from '@/store/appStore'
import { VPNStatus } from './VPNStatus'


export function Header() {
  const { theme, toggleTheme, toggleSidebar } = useAppStore()

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="flex h-14 items-center px-4 gap-4">
        <button
          onClick={toggleSidebar}
          className="inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground h-9 w-9"
        >
          <Menu className="h-5 w-5" />
        </button>
        
        <div className="flex items-center gap-2">
          <span className="text-xl font-bold">📺 Media Manager</span>
        </div>

        <div className="flex flex-1 items-center justify-end gap-2">
          <VPNStatus />
          
          <button
            onClick={toggleTheme}
            className="inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground h-9 w-9"
          >
            {theme === 'dark' ? <Sun className="h-5 w-5" /> : <Moon className="h-5 w-5" />}
          </button>
          
          <button className="inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground h-9 w-9">
            <Bell className="h-5 w-5" />
          </button>
        </div>
      </div>
    </header>
  )
}