import { useState } from 'react'
import { Card, CardContent } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Badge } from '@/components/ui/Badge'
import { 
  useSuggestions, 
  useApproveSuggestion, 
  useRejectSuggestion,
  useBulkApproveSuggestions,
  useBulkRejectSuggestions,
  useGenerateSuggestions 
} from '@/hooks/useApi'
import { useAppStore } from '@/store/appStore'
import { DownloadSuggestion } from '@/types'
import { getStatusBgColor, formatBytes } from '@/lib/utils'
import { Search, Check, X, Sparkles, Loader2 } from 'lucide-react'

export function SearchPage() {
  const [query, setQuery] = useState('')
  const [status, setStatus] = useState('pending')
  const { data: suggestions, isLoading } = useSuggestions({ q: query, status })
  const { addToast, selectedItems, toggleSelection, clearSelection } = useAppStore()
  
  const approveSuggestion = useApproveSuggestion()
  const rejectSuggestion = useRejectSuggestion()
  const bulkApprove = useBulkApproveSuggestions()
  const bulkReject = useBulkRejectSuggestions()
  const generateSuggestions = useGenerateSuggestions()

  const handleApprove = async (id: number) => {
    try {
      await approveSuggestion.mutateAsync(id)
      addToast('Suggestion approved', 'success')
    } catch {
      addToast('Failed to approve suggestion', 'error')
    }
  }

  const handleReject = async (id: number) => {
    try {
      await rejectSuggestion.mutateAsync(id)
      addToast('Suggestion rejected', 'success')
    } catch {
      addToast('Failed to reject suggestion', 'error')
    }
  }

  const handleBulkApprove = async () => {
    try {
      await bulkApprove.mutateAsync(Array.from(selectedItems))
      addToast(`${selectedItems.size} suggestions approved`, 'success')
      clearSelection()
    } catch {
      addToast('Failed to approve suggestions', 'error')
    }
  }

  const handleBulkReject = async () => {
    try {
      await bulkReject.mutateAsync(Array.from(selectedItems))
      addToast(`${selectedItems.size} suggestions rejected`, 'success')
      clearSelection()
    } catch {
      addToast('Failed to reject suggestions', 'error')
    }
  }

  const handleGenerate = async () => {
    try {
      await generateSuggestions.mutateAsync()
      addToast('Suggestions generated', 'success')
    } catch {
      addToast('Failed to generate suggestions', 'error')
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Search</h1>
          <p className="text-muted-foreground">
            Find and approve torrents to download
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

      {/* Search and Filters */}
      <div className="flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search torrents..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="pl-9"
          />
        </div>
        <select
          value={status}
          onChange={(e) => setStatus(e.target.value)}
          className="rounded-md border border-input bg-background px-3 py-2"
        >
          <option value="pending">Pending</option>
          <option value="approved">Approved</option>
          <option value="rejected">Rejected</option>
          <option value="downloaded">Downloaded</option>
        </select>
      </div>

      {/* Bulk Actions */}
      {selectedItems.size > 0 && (
        <div className="flex items-center gap-2 p-2 bg-accent rounded-md">
          <span className="text-sm font-medium">{selectedItems.size} selected</span>
          <Button size="sm" onClick={handleBulkApprove}>
            <Check className="mr-2 h-4 w-4" />
            Approve
          </Button>
          <Button size="sm" variant="outline" onClick={handleBulkReject}>
            <X className="mr-2 h-4 w-4" />
            Reject
          </Button>
          <Button size="sm" variant="ghost" onClick={clearSelection}>
            Clear
          </Button>
        </div>
      )}

      {/* Results Grid */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {suggestions?.data?.map((suggestion: DownloadSuggestion) => (
          <Card key={suggestion.id} className="overflow-hidden">
            <div className="aspect-[2/3] relative bg-muted">
              <img
                src={suggestion.poster_url || '/web/images/placeholder-movie.jpg'}
                alt={suggestion.title}
                className="object-cover w-full h-full"
                loading="lazy"
              />
              <div className="absolute top-2 left-2">
                <input
                  type="checkbox"
                  checked={selectedItems.has(suggestion.id)}
                  onChange={() => toggleSelection(suggestion.id)}
                  className="h-4 w-4 rounded border-gray-300"
                />
              </div>
            </div>
            <CardContent className="p-4">
              <h3 className="font-semibold text-sm truncate">{suggestion.title}</h3>
              <div className="flex items-center justify-between mt-2">
                <Badge className={getStatusBgColor(suggestion.status)}>
                  {suggestion.status}
                </Badge>
                <span className="text-xs text-muted-foreground">
                  {formatBytes(suggestion.size)}
                </span>
              </div>
              <div className="flex items-center justify-between mt-2 text-xs text-muted-foreground">
                <span>{suggestion.seeders} seeders</span>
                <span>{suggestion.leechers} leechers</span>
              </div>
              
              {suggestion.status === 'pending' && (
                <div className="flex gap-2 mt-3">
                  <Button
                    size="sm"
                    className="flex-1"
                    onClick={() => handleApprove(suggestion.id)}
                    disabled={approveSuggestion.isPending}
                  >
                    <Check className="mr-1 h-3 w-3" />
                    Approve
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    className="flex-1"
                    onClick={() => handleReject(suggestion.id)}
                    disabled={rejectSuggestion.isPending}
                  >
                    <X className="mr-1 h-3 w-3" />
                    Reject
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {suggestions?.data?.length === 0 && !isLoading && (
        <Card>
          <CardContent className="p-6 text-center">
            <p className="text-muted-foreground">No suggestions found</p>
            <p className="text-sm text-muted-foreground mt-1">
              Try generating suggestions or adjusting your search
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

export { SearchPage as Search }