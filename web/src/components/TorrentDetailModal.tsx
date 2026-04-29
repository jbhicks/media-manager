import { Dialog } from '@/components/ui/Dialog'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { formatBytes } from '@/lib/utils'
import { Download, Check, X, ExternalLink, Image, Loader2 } from 'lucide-react'
import { useState, useEffect } from 'react'

interface TorrentDetailModalProps {
  isOpen: boolean
  onClose: () => void
  torrent: {
    title: string
    infoHash?: string
    magnetLink?: string
    torrentURL?: string
    size: number
    seeders: number
    leechers: number
    category?: string
    uploadDate?: string
    poster_url?: string
    posterUrl?: string
    source?: { name?: string }
  } | null
  onApprove?: () => void
  onReject?: () => void
  onDownload?: () => void
  status?: string
}

export function TorrentDetailModal({
  isOpen,
  onClose,
  torrent,
  onApprove,
  onReject,
  onDownload,
  status,
}: TorrentDetailModalProps) {
  const [posterUrl, setPosterUrl] = useState('')
  const [isExtracting, setIsExtracting] = useState(false)
  const [extractError, setExtractError] = useState('')

  // Sync poster URL when torrent changes
  useEffect(() => {
    setPosterUrl(torrent?.poster_url || torrent?.posterUrl || '')
    setExtractError('')
  }, [torrent])

  if (!torrent) return null

  const handleExtractImages = async () => {
    if (!torrent.magnetLink) {
      setExtractError('No magnet link available')
      return
    }

    setIsExtracting(true)
    setExtractError('')

    try {
      const response = await fetch('/api/search/extract-images', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ magnet_link: torrent.magnetLink }),
      })

      const data = await response.json()

      if (data.images && data.images.length > 0) {
        // Use the first extracted image as poster
        const firstImage = data.images[0]
        // Convert absolute path to web URL if needed
        const imageUrl = firstImage.startsWith('/') ? firstImage : `/images/${firstImage.split('/').pop()}`
        setPosterUrl(imageUrl)
      } else if (data.error) {
        setExtractError(data.error)
      } else {
        setExtractError('No images found in torrent')
      }
    } catch (error) {
      setExtractError('Failed to extract images')
    } finally {
      setIsExtracting(false)
    }
  }

  return (
    <Dialog open={isOpen} onClose={onClose} title={torrent.title}>
      <div className="grid grid-cols-[200px_1fr] gap-6">
        {/* Poster */}
        <div className="aspect-[2/3] rounded-lg overflow-hidden bg-muted relative">
          {posterUrl ? (
            <img
              src={posterUrl}
              alt={torrent.title}
              className="w-full h-full object-cover"
            />
          ) : (
            <div className="w-full h-full flex flex-col items-center justify-center text-muted-foreground p-4">
              <Image className="h-12 w-12 mb-2 opacity-50" />
              <span className="text-sm">No Poster</span>
              {torrent.magnetLink && (
                <Button
                  variant="outline"
                  size="sm"
                  className="mt-3"
                  onClick={handleExtractImages}
                  disabled={isExtracting}
                >
                  {isExtracting ? (
                    <Loader2 className="mr-2 h-3 w-3 animate-spin" />
                  ) : (
                    <Image className="mr-2 h-3 w-3" />
                  )}
                  Extract Cover
                </Button>
              )}
              {extractError && (
                <span className="text-xs text-red-500 mt-2 text-center">{extractError}</span>
              )}
            </div>
          )}
        </div>

        {/* Details */}
        <div className="space-y-4">
          <div className="flex flex-wrap gap-2">
            {torrent.category && (
              <Badge variant="secondary">{torrent.category}</Badge>
            )}
            <Badge className="bg-green-600 text-white">
              {torrent.seeders} Seeders
            </Badge>
            <Badge className="bg-red-600 text-white">
              {torrent.leechers} Leechers
            </Badge>
            <Badge variant="outline">{formatBytes(torrent.size)}</Badge>
            {status && (
              <Badge variant="outline">{status}</Badge>
            )}
          </div>

          <div className="space-y-2 text-sm">
            {torrent.source?.name && (
              <div>
                <span className="text-muted-foreground">Source: </span>
                {torrent.source.name}
              </div>
            )}
            {torrent.infoHash && (
              <div className="font-mono text-xs bg-muted p-2 rounded">
                <span className="text-muted-foreground">Info Hash: </span>
                {torrent.infoHash}
              </div>
            )}
            {torrent.uploadDate && (
              <div>
                <span className="text-muted-foreground">Uploaded: </span>
                {new Date(torrent.uploadDate).toLocaleDateString()}
              </div>
            )}
          </div>

          <div className="flex gap-2 pt-4">
            {onApprove && status === 'pending' && (
              <Button
                className="bg-green-600 hover:bg-green-700"
                onClick={onApprove}
              >
                <Check className="mr-2 h-4 w-4" />
                Approve
              </Button>
            )}
            {onReject && status === 'pending' && (
              <Button
                variant="outline"
                className="border-red-600 text-red-600 hover:bg-red-50"
                onClick={onReject}
              >
                <X className="mr-2 h-4 w-4" />
                Reject
              </Button>
            )}
            {onDownload && (
              <Button variant="outline" onClick={onDownload}>
                <Download className="mr-2 h-4 w-4" />
                Download
              </Button>
            )}
            {torrent.magnetLink && (
              <Button
                variant="outline"
                onClick={() => window.open(torrent.magnetLink, '_blank')}
              >
                <ExternalLink className="mr-2 h-4 w-4" />
                Magnet
              </Button>
            )}
          </div>
        </div>
      </div>
    </Dialog>
  )
}
