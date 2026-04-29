import { useEffect, useRef } from 'react'
import { Badge } from '@/components/ui/Badge'
import { formatBytes } from '@/lib/utils'
import { Download, ExternalLink } from 'lucide-react'

interface ContextMenuProps {
  isOpen: boolean
  x: number
  y: number
  group: {
    baseTitle: string
    year: string
    poster_url: string
    variants: Array<{
      title: string
      infoHash: string
      magnetLink: string
      torrentURL: string
      size: number
      seeders: number
      leechers: number
      category: string
      uploadDate: string
      source?: { name?: string }
    }>
  } | null
  onClose: () => void
  onDownload: (magnetLink: string) => void
  onSelectVariant: (variant: any) => void
}

export function ContextMenu({ isOpen, x, y, group, onClose, onDownload, onSelectVariant }: ContextMenuProps) {
  const menuRef = useRef<HTMLDivElement>(null)
  
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        onClose()
      }
    }
    
    if (isOpen) {
      document.addEventListener('click', handleClickOutside)
      document.addEventListener('scroll', onClose, true)
    }
    
    return () => {
      document.removeEventListener('click', handleClickOutside)
      document.removeEventListener('scroll', onClose, true)
    }
  }, [isOpen, onClose])
  
  if (!isOpen || !group) return null
  
  // Group variants by quality
  const qualityGroups = group.variants.reduce((acc, variant) => {
    const quality = extractQuality(variant.title)
    if (!acc[quality]) {
      acc[quality] = []
    }
    acc[quality].push(variant)
    return acc
  }, {} as Record<string, typeof group.variants>)
  
  const sortedQualities = Object.keys(qualityGroups).sort((a, b) => {
    const qualityOrder = ['4K', '2160p', '1080p', '720p', '480p', 'Other']
    const idxA = qualityOrder.indexOf(a)
    const idxB = qualityOrder.indexOf(b)
    return (idxA === -1 ? 999 : idxA) - (idxB === -1 ? 999 : idxB)
  })
  
  // Adjust position to keep menu on screen
  const adjustedX = Math.min(x, window.innerWidth - 350)
  const adjustedY = Math.min(y, window.innerHeight - 400)
  
  return (
    <div
      ref={menuRef}
      className="fixed z-50 w-80 bg-popover border rounded-lg shadow-lg py-2 overflow-hidden"
      style={{ left: adjustedX, top: adjustedY }}
    >
      <div className="px-3 py-2 border-b">
        <h3 className="font-semibold text-sm truncate">{group.baseTitle}</h3>
        {group.year && <span className="text-xs text-muted-foreground">{group.year}</span>}
      </div>
      
      <div className="max-h-96 overflow-y-auto">
        {sortedQualities.map(quality => (
          <div key={quality} className="py-1">
            <div className="px-3 py-1 text-xs font-medium text-muted-foreground bg-muted/50">
              {quality} ({qualityGroups[quality].length} version{qualityGroups[quality].length > 1 ? 's' : ''})
            </div>
            {qualityGroups[quality].map((variant, idx) => (
              <div
                key={idx}
                className="px-3 py-2 hover:bg-accent cursor-pointer flex items-center justify-between group"
                onClick={() => {
                  onSelectVariant(variant)
                  onClose()
                }}
              >
                <div className="flex-1 min-w-0 mr-2">
                  <div className="text-sm truncate" title={variant.title}>
                    {variant.title}
                  </div>
                  <div className="flex items-center gap-2 mt-1">
                    <Badge variant="outline" className="text-[10px]">
                      {formatBytes(variant.size)}
                    </Badge>
                    <Badge className="bg-green-600 text-white text-[10px]">
                      {variant.seeders}S
                    </Badge>
                    {variant.source?.name && (
                      <Badge variant="secondary" className="text-[10px]">
                        {variant.source.name}
                      </Badge>
                    )}
                  </div>
                </div>
                
                <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100">
                  {variant.magnetLink && (
                    <button
                      className="p-1.5 rounded hover:bg-primary hover:text-primary-foreground transition-colors"
                      onClick={(e) => {
                        e.stopPropagation()
                        onDownload(variant.magnetLink)
                      }}
                      title="Download"
                    >
                      <Download className="h-3.5 w-3.5" />
                    </button>
                  )}
                  {variant.magnetLink && (
                    <button
                      className="p-1.5 rounded hover:bg-primary hover:text-primary-foreground transition-colors"
                      onClick={(e) => {
                        e.stopPropagation()
                        window.open(variant.magnetLink, '_blank')
                      }}
                      title="Open Magnet"
                    >
                      <ExternalLink className="h-3.5 w-3.5" />
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        ))}
      </div>
      
      <div className="border-t px-3 py-2 text-xs text-muted-foreground">
        {group.variants.length} total variants
      </div>
    </div>
  )
}

function extractQuality(title: string): string {
  const lower = title.toLowerCase()
  if (lower.includes('2160p') || lower.includes('4k')) return '4K'
  if (lower.includes('1080p')) return '1080p'
  if (lower.includes('720p')) return '720p'
  if (lower.includes('480p')) return '480p'
  return 'Other'
}
