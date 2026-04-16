import { Card, CardContent } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { useLibrary, useReprocessLibrary } from '@/hooks/useApi'
import { useAppStore } from '@/store/appStore'
import { RefreshCw } from 'lucide-react'

export function Library() {
  const { data: movies, isLoading } = useLibrary()
  const { addToast } = useAppStore()
  const reprocessLibrary = useReprocessLibrary()

  const handleReprocess = async () => {
    try {
      await reprocessLibrary.mutateAsync()
      addToast('Library reprocessed successfully', 'success')
    } catch {
      addToast('Failed to reprocess library', 'error')
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
            onClick={handleReprocess}
            disabled={reprocessLibrary.isPending}
          >
            <RefreshCw className="mr-2 h-4 w-4" />
            Clean Filenames
          </Button>
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {movies?.map((movie, index) => (
          <Card key={index} className="overflow-hidden">
            <div className="aspect-[2/3] relative bg-muted">
              <img
                src={movie.poster_url}
                alt={movie.title}
                className="object-cover w-full h-full"
                loading="lazy"
              />
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

      {movies?.length === 0 && (
        <Card>
          <CardContent className="p-6 text-center">
            <p className="text-muted-foreground">No media files found</p>
            <p className="text-sm text-muted-foreground mt-1">
              Your library will appear here once you download some media
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}