import { NavLink } from 'react-router-dom'
import { 
  Home, 
  Download, 
  Library, 
  Search, 
  Lightbulb, 
  Settings 
} from 'lucide-react'
import { useAppStore } from '@/store/appStore'
import { cn } from '@/lib/utils'

const navigation = [
  { name: 'Home', href: '/', icon: Home },
  { name: 'Downloads', href: '/downloads', icon: Download },
  { name: 'Library', href: '/library', icon: Library },
  { name: 'Search', href: '/search', icon: Search },
  { name: 'Suggestions', href: '/suggestions', icon: Lightbulb },
  { name: 'Settings', href: '/settings', icon: Settings },
]

export function Sidebar() {
  const { sidebarOpen } = useAppStore()

  if (!sidebarOpen) {
    return (
      <nav className="w-16 border-r bg-background min-h-[calc(100vh-3.5rem)]">
        <div className="flex flex-col gap-2 p-2">
          {navigation.map((item) => (
            <NavLink
              key={item.name}
              to={item.href}
              className={({ isActive }) =>
                cn(
                  'inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground h-9 w-9',
                  isActive && 'bg-accent text-accent-foreground'
                )
              }
            >
              <item.icon className="h-5 w-5" />
            </NavLink>
          ))}
        </div>
      </nav>
    )
  }

  return (
    <nav className="w-64 border-r bg-background min-h-[calc(100vh-3.5rem)]">
      <div className="flex flex-col gap-1 p-4">
        {navigation.map((item) => (
          <NavLink
            key={item.name}
            to={item.href}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground',
                isActive && 'bg-accent text-accent-foreground'
              )
            }
          >
            <item.icon className="h-5 w-5" />
            {item.name}
          </NavLink>
        ))}
      </div>
    </nav>
  )
}