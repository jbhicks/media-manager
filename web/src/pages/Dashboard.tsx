import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card'
import { useStats } from '@/hooks/useApi'
import { Download, Library, Search, CheckCircle, XCircle, Clock, Activity } from 'lucide-react'


export function Dashboard() {
  const { data: stats, isLoading } = useStats()

  const statCards = [
    { name: 'Pending', value: stats?.pending || 0, icon: Clock, color: 'text-yellow-500' },
    { name: 'Downloading', value: stats?.downloading || 0, icon: Activity, color: 'text-blue-500' },
    { name: 'Completed', value: stats?.completed || 0, icon: CheckCircle, color: 'text-green-500' },
    { name: 'Failed', value: stats?.failed || 0, icon: XCircle, color: 'text-red-500' },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">
          Welcome to your media manager. Monitor your downloads and library.
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {statCards.map((stat) => (
          <Card key={stat.name}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">{stat.name}</CardTitle>
              <stat.icon className={`h-4 w-4 ${stat.color}`} />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {isLoading ? '...' : stat.value}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Quick Actions */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card className="cursor-pointer hover:bg-accent/50 transition-colors">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Download className="h-5 w-5" />
              Downloads
            </CardTitle>
            <CardDescription>Manage your download queue</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              View active downloads, cancel or restart tasks, and monitor progress.
            </p>
          </CardContent>
        </Card>

        <Card className="cursor-pointer hover:bg-accent/50 transition-colors">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Library className="h-5 w-5" />
              Library
            </CardTitle>
            <CardDescription>Browse your media collection</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              View downloaded movies, fetch posters, and organize your library.
            </p>
          </CardContent>
        </Card>

        <Card className="cursor-pointer hover:bg-accent/50 transition-colors">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Search className="h-5 w-5" />
              Search
            </CardTitle>
            <CardDescription>Find new content</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Search across multiple indexers and approve downloads.
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}