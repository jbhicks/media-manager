import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { useSuggestions, useSuggestionStats, useGenerateSuggestions } from '@/hooks/useApi'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { DownloadSuggestion } from '@/types'
import { useAppStore } from '@/store/appStore'
import { getStatusBgColor } from '@/lib/utils'
import { Sparkles, Loader2 } from 'lucide-react'

export function Suggestions() {
  const { data: stats } = useSuggestionStats()
  const { data: suggestions } = useSuggestions({ status: 'pending' })
  const { addToast } = useAppStore()
  const generateSuggestions = useGenerateSuggestions()

  const handleGenerate = async () => {
    try {
      await generateSuggestions.mutateAsync()
      addToast('Suggestions generated successfully', 'success')
    } catch {
      addToast('Failed to generate suggestions', 'error')
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Suggestions</h1>
          <p className="text-muted-foreground">
            AI-powered download suggestions
          </p>
        </div>
        <Button onClick={handleGenerate} disabled={generateSuggestions.isPending}>
          {generateSuggestions.isPending ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <Sparkles className="mr-2 h-4 w-4" />
          )}
          Generate Suggestions
        </Button>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Pending</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.pending || 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Approved</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.approved || 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Rejected</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.rejected || 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.total || 0}</div>
          </CardContent>
        </Card>
      </div>

      {/* Recent Suggestions */}
      <div className="space-y-4">
        <h2 className="text-xl font-semibold">Recent Suggestions</h2>
        {suggestions?.data?.slice(0, 5).map((suggestion: DownloadSuggestion) => (
          <Card key={suggestion.id}>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="font-semibold">{suggestion.title}</h3>
                  <div className="flex items-center gap-2 mt-1 text-sm text-muted-foreground">
                    <Badge className={getStatusBgColor(suggestion.status)}>
                      {suggestion.status}
                    </Badge>
                    <span>{suggestion.seeders} seeders</span>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}

        {suggestions?.data?.length === 0 && (
          <Card>
            <CardContent className="p-6 text-center">
              <p className="text-muted-foreground">No pending suggestions</p>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}