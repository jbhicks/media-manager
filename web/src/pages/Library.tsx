import { useState, useMemo } from 'react'
import { Card, CardContent } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useLibrary, useReprocessLibrary, useFetchAllPosters } from '@/hooks/useApi'
import { useAppStore } from '@/store/appStore'
import { RefreshCw, ImageDown, Search, X } from 'lucide-react'

export function Library() {
  const { data: movies, isLoading } = useLibrary()
  const { addToast } = useAppStore()
  const reprocessLibrary = useReprocessLibrary()
  const fetchAllPosters = useFetchAllPosters()
  const [failedPosters, setFailedPosters] = useState<Set<string>>(new Set())
  const [filterQuery, setFilterQuery] = useState('')

  const filteredMovies = useMemo(() => {
    if (!filterQuery.trim()) return movies
    const query = filterQuery.toLowerCase()
    return movies?.filter(movie =>
      movie.title.toLowerCase().includes(query)
    )
  }, [movies, filterQuery])

  const getErrorMessage = (error: any): string => {
    return error?.userMessage || error?.message || 'Unknown error'
  }

  const handleReprocess = async () => {
    try {
      await reprocessLibrary.mutateAsync()
      addToast('Library reprocessed successfully', 'success')
    } catch (error: any) {
      addToast(`Failed to reprocess: ${getErrorMessage(error)}`, 'error')
    }
  }

  const handleFetchPosters = async () => {
    try {
      const result = await fetchAllPosters.mutateAsync()
      addToast(
        `Fetched ${result.fetched} posters, ${result.cached} cached, ${result.failed} failed`,
        'success'
      )
    } catch (error: any) {
      addToast(`Failed to fetch posters: ${getErrorMessage(error)}`, 'error')
    }
  }

  if (isLoading) {
    return <div>Loading...</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Library</h1>
          <p className="text-muted-foreground">
            Browse your media collection ({movies?.length || 0} items)
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={handleFetchPosters}
            disabled={fetchAllPosters.isPending}
          >
            <ImageDown className="mr-2 h-4 w-4" />
            Fetch Posters
          </Button>
          <Button
            variant="outline"
            onClick={handleReprocess}
            disabled={reprocessLibrary.isPending}
          >
            <RefreshCw className="mr-2 h-4 w-4" />
            Clean Filenames
          </Button>
        </div>
      </div>

      <div className="relative max-w-sm">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Filter movies..."
          value={filterQuery}
          onChange={(e) => setFilterQuery(e.target.value)}
          className="pl-9 pr-9"
        />
        {filterQuery && (
          <button
            onClick={() => setFilterQuery('')}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>

      {filterQuery && filteredMovies && (
        <p className="text-sm text-muted-foreground">
          Showing {filteredMovies.length} of {movies?.length || 0} movies
        </p>
      )}

      <div className="grid gap-4 sm:grid-cols-4 lg:grid-cols-6 xl:grid-cols-8">
        {filteredMovies?.map((movie) => (
          <Card key={movie.path} className="overflow-hidden">
            <div className="aspect-[2/3] relative bg-muted">
              {movie.poster_url && !failedPosters.has(movie.poster_url) ? (
                <img
                  src={movie.poster_url}
                  alt={movie.title}
                  className="object-cover w-full h-full"
                  loading="lazy"
                  onError={() => {
                    setFailedPosters(prev => new Set(prev).add(movie.poster_url))
                  }}
                />
              ) : (
                <div className="flex items-center justify-center w-full h-full bg-gradient-to-b from-muted/50 to-muted">
                  <div className="text-center p-4">
                    <p className="text-sm font-medium line-clamp-3">{movie.title}</p>
                    {movie.year > 0 && (
                      <p className="text-xs text-muted-foreground mt-1">{movie.year}</p>
                    )}
                  </div>
                </div>
              )}
            </div>
            <CardContent className="p-4">
              <h3 className="font-semibold truncate">{movie.title}</h3>
              <div className="flex items-center justify-between mt-2 text-sm text-muted-foreground">
                <span>{movie.size}</span>
                {movie.rating && <span>{movie.rating}</span>}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {filteredMovies?.length === 0 && (
        <Card>
          <CardContent className="p-6 text-center">
            {filterQuery ? (
              <>
                <p className="text-muted-foreground">No movies match "{filterQuery}" </p>
                <Button
                  variant="ghost"
                  size="sm"
                  className="mt-2"
                  onClick={() => setFilterQuery('')}
                >
                  Clear filter
                </Button>
              </>
            ) : (
              <>
                <p className="text-muted-foreground">No media files found</p>
                <p className="text-sm text-muted-foreground mt-1">
                  Your library will appear here once you download some media
                </p>
              </>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  )
}