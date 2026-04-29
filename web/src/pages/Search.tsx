import { useState, useMemo, useEffect, useRef } from 'react'
import { Card, CardContent } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Badge } from '@/components/ui/Badge'
import {
  useSources,
  useSearchActivity,
  useAllJackettIndexers,
  useCreateSource,
  useUpdateSource,
  useDeleteSource,
} from '@/hooks/useApi'
import { useAppStore } from '@/store/appStore'
import { PosterImage } from '@/components/PosterImage'
import { SearchHistory, addToHistory } from '@/components/SearchHistory'
import { TorrentDetailModal } from '@/components/TorrentDetailModal'
import { ContextMenu } from '@/components/ContextMenu'
import { DownloadSource } from '@/types'
import { formatBytes } from '@/lib/utils'
import {
  Search,
  Loader2,
  Download,
  Activity,
  Radio,
  Server,
  Plus,
  Pencil,
  Trash2,
  ChevronDown,
  ChevronUp,
  Grid3X3,
  List,
  Filter,
  ArrowUpDown,
  X,
} from 'lucide-react'

interface SearchResult {
  id?: number
  title: string
  info_hash: string
  magnet_link: string
  torrentURL: string
  size: number
  seeders: number
  leechers: number
  category: string
  uploadDate: string
  poster_url: string
  indexer?: string
  sourceName: string
  source?: any
  status?: string
}

interface GroupedResult {
  baseTitle: string
  year: string
  poster_url: string
  variants: SearchResult[]
}

// Extract base title and year from torrent names
function extractBaseTitle(title: string): { baseTitle: string; year: string } {
  const yearMatch = title.match(/\(?(\d{4})\)?/)
  const year = yearMatch ? yearMatch[1] : ''

  let baseTitle = title
    .replace(/\(?(\d{4})\)?/g, '')
    .replace(/\d{3,4}p/gi, '')
    .replace(/\b(BluRay|WEB-DL|WEBRip|HDRip|DVDRip|BRRip|HDTV|WEB-DL|WEB\.DL|AMZN|NF|DSNP)\b/gi, '')
    .replace(/\b(x264|x265|HEVC|H\.264|H\.265|AVC|AV1)\b/gi, '')
    .replace(/\b(AC3|AAC|DTS|DDP5\.1|DDP|Atmos|TrueHD)\b/gi, '')
    .replace(/\b(5\.1|2\.0|7\.1)\b/g, '')
    .replace(/-[A-Za-z0-9]+$/g, '')
    .replace(/\.|_/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()

  return { baseTitle, year }
}

// Calculate quality score for a search result (same logic as Suggestions)
function getQualityScore(result: SearchResult): number {
  const t = result.title.toLowerCase()
  let score = 0
  // Resolution
  if (t.includes('2160p') || t.includes('4k')) score += 50
  else if (t.includes('1080p')) score += 40
  else if (t.includes('720p')) score += 30
  else if (t.includes('480p')) score += 20
  else score += 10
  // Source quality
  if (t.includes('web-dl') || t.includes('webdl')) score += 20
  else if (t.includes('bluray') || t.includes('blu-ray')) score += 18
  else if (t.includes('webrip')) score += 15
  else if (t.includes('hdrip')) score += 10
  else if (t.includes('ts') || t.includes('cam')) score -= 20
  // Seeders
  score += Math.min(result.seeders / 10, 20)
  return score
}

// Get quality label from title
function getQualityLabel(title: string): string {
  const t = title.toLowerCase()
  const parts: string[] = []
  if (t.includes('2160p') || t.includes('4k')) parts.push('4K')
  else if (t.includes('1080p')) parts.push('1080p')
  else if (t.includes('720p')) parts.push('720p')
  else if (t.includes('480p')) parts.push('480p')
  
  if (t.includes('web-dl') || t.includes('webdl')) parts.push('WEB-DL')
  else if (t.includes('bluray') || t.includes('blu-ray')) parts.push('BluRay')
  else if (t.includes('webrip')) parts.push('WEBRip')
  else if (t.includes('hdrip')) parts.push('HDRip')
  else if (t.includes('hdtv')) parts.push('HDTV')
  
  if (t.includes('x265') || t.includes('hevc')) parts.push('HEVC')
  
  return parts.join(' ') || 'Unknown'
}

function groupResults(results: SearchResult[]): GroupedResult[] {
  const groups = new Map<string, GroupedResult>()

  for (const result of results) {
    const { baseTitle, year } = extractBaseTitle(result.title)
    const key = `${baseTitle.toLowerCase()}_${year}`

    if (!groups.has(key)) {
      groups.set(key, {
        baseTitle,
        year,
        poster_url: result.poster_url || '',
        variants: [],
      })
    }

    const group = groups.get(key)!
    group.variants.push(result)

    if (!group.poster_url && result.poster_url) {
      group.poster_url = result.poster_url
    }
  }

  // Sort variants by quality score (best first) instead of just seeders
  for (const group of groups.values()) {
    group.variants.sort((a, b) => getQualityScore(b) - getQualityScore(a))
  }

  return Array.from(groups.values()).sort((a, b) => {
    const aScore = a.variants[0] ? getQualityScore(a.variants[0]) : 0
    const bScore = b.variants[0] ? getQualityScore(b.variants[0]) : 0
    return bScore - aScore
  })
}

// Filter helpers
function filterByQuality(results: SearchResult[], quality: string): SearchResult[] {
  if (!quality || quality === 'all') return results
  const qualityMap: Record<string, string[]> = {
    '1080p': ['1080p', '1920x1080'],
    '720p': ['720p', '1280x720'],
    '2160p': ['2160p', '4k', 'uhd'],
    'webdl': ['web-dl', 'webdl', 'web dl'],
    'bluray': ['bluray', 'blu-ray', 'bdrip', 'brrip'],
    'webrip': ['webrip', 'web-rip'],
  }
  const patterns = qualityMap[quality] || [quality]
  return results.filter(r => patterns.some(p => r.title.toLowerCase().includes(p)))
}

function filterBySize(results: SearchResult[], minSize: number, maxSize: number): SearchResult[] {
  return results.filter(r => {
    if (minSize > 0 && r.size < minSize) return false
    if (maxSize > 0 && r.size > maxSize) return false
    return true
  })
}

function filterBySeeders(results: SearchResult[], minSeeders: number): SearchResult[] {
  if (minSeeders <= 0) return results
  return results.filter(r => r.seeders >= minSeeders)
}

function filterByCategory(results: SearchResult[], category: string): SearchResult[] {
  if (!category || category === 'all') return results
  return results.filter(r => r.category?.toLowerCase().includes(category.toLowerCase()))
}

// Sort helpers
function sortResults(results: SearchResult[], sortBy: string, direction: 'asc' | 'desc'): SearchResult[] {
  const sorted = [...results]
  const mult = direction === 'asc' ? 1 : -1
  switch (sortBy) {
    case 'seeders':
      sorted.sort((a, b) => mult * (a.seeders - b.seeders))
      break
    case 'size':
      sorted.sort((a, b) => mult * (a.size - b.size))
      break
    case 'name':
      sorted.sort((a, b) => mult * a.title.localeCompare(b.title))
      break
    case 'date':
      sorted.sort((a, b) => mult * (new Date(a.uploadDate || 0).getTime() - new Date(b.uploadDate || 0).getTime()))
      break
    case 'leechers':
      sorted.sort((a, b) => mult * (a.leechers - b.leechers))
      break
    case 'quality':
      sorted.sort((a, b) => {
        const getQuality = (title: string) => {
          const t = title.toLowerCase()
          if (t.includes('2160p') || t.includes('4k')) return 5
          if (t.includes('1080p')) return 4
          if (t.includes('720p')) return 3
          if (t.includes('480p')) return 2
          return 1
        }
        return mult * (getQuality(a.title) - getQuality(b.title))
      })
      break
    default:
      sorted.sort((a, b) => mult * (a.seeders - b.seeders))
  }
  return sorted
}

const INDEXERS_STORAGE_KEY = 'media-manager-selected-indexers'

export function SearchPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [viewMode, setViewMode] = useState<'grid' | 'table'>('table')
  const [showActivityLog, setShowActivityLog] = useState(false)
  const [selectedIndexers, setSelectedIndexers] = useState<Set<string>>(new Set())
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set())

  // Filter state
  const [showFilters, setShowFilters] = useState(false)
  const [qualityFilter, setQualityFilter] = useState('all')
  const [minSeedersFilter, setMinSeedersFilter] = useState(0)
  const [minSizeFilter, setMinSizeFilter] = useState(0)
  const [maxSizeFilter, setMaxSizeFilter] = useState(0)
  const [categoryFilter, setCategoryFilter] = useState('all')

  // Sort state
  const [sortBy, setSortBy] = useState('seeders')
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc')

  // Detail modal state
  const [detailModalOpen, setDetailModalOpen] = useState(false)
  const [selectedTorrent, setSelectedTorrent] = useState<any>(null)

  // Context menu state
  const [contextMenuOpen, setContextMenuOpen] = useState(false)
  const [contextMenuPos, setContextMenuPos] = useState({ x: 0, y: 0 })
  const [contextMenuGroup, setContextMenuGroup] = useState<any>(null)

  // Pagination state
  const [allSearchResults, setAllSearchResults] = useState<SearchResult[]>([])
  const [currentOffset, setCurrentOffset] = useState(0)
  const [hasMoreResults, setHasMoreResults] = useState(true)
  const [isSearching, setIsSearching] = useState(false)
  const RESULTS_PER_PAGE = 100

  // Infinite scroll ref
  const loadMoreRef = useRef<HTMLDivElement>(null)
  const scrollContainerRef = useRef<HTMLDivElement>(null)

  // Async poster extraction state
  const [extractingPosters, setExtractingPosters] = useState<Set<string>>(new Set())
  const [failedExtractions, setFailedExtractions] = useState<Set<string>>(new Set())
  const MAX_CONCURRENT_EXTRACTIONS = 3
  const EXTRACTION_DELAY_MS = 500

  // Extract poster from torrent for a single result
  const extractPosterForResult = async (result: SearchResult) => {
    if (!result.magnet_link || !result.info_hash) return
    if (extractingPosters.has(result.info_hash)) return
    if (failedExtractions.has(result.info_hash)) return
    if ((result as any).poster_url || (result as any).posterUrl) return

    setExtractingPosters(prev => new Set(prev).add(result.info_hash))

    try {
      const response = await fetch('/api/search/extract-images', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ magnet_link: result.magnet_link }),
      })

      if (!response.ok) {
        // Mark as failed so we don't retry infinitely
        setFailedExtractions(prev => new Set(prev).add(result.info_hash))
        return
      }

      const data = await response.json()
      if (data.images && data.images.length > 0) {
        const firstImage = data.images[0]
        const imageUrl = firstImage.startsWith('/') ? firstImage : `/images/${firstImage.split('/').pop()}`

        // Update the result in state
        setAllSearchResults(prev =>
          prev.map(r =>
            r.info_hash === result.info_hash
              ? { ...r, poster_url: imageUrl, posterUrl: imageUrl }
              : r
          )
        )
      }
    } catch (error) {
      console.error(`Failed to extract poster for ${result.title}:`, error)
      // Mark as failed so we don't retry infinitely
      setFailedExtractions(prev => new Set(prev).add(result.info_hash))
    } finally {
      setExtractingPosters(prev => {
        const next = new Set(prev)
        next.delete(result.info_hash)
        return next
      })
    }
  }

  // Trigger async poster extraction for results without posters
  useEffect(() => {
    if (allSearchResults.length === 0) return

    const resultsWithoutPosters = allSearchResults.filter(
      r => r.magnet_link && !((r as any).poster_url || (r as any).posterUrl) && !failedExtractions.has(r.info_hash)
    )

    if (resultsWithoutPosters.length === 0) return

    // Process in batches with delay to avoid overwhelming the backend
    let index = 0
    let timeoutId: ReturnType<typeof setTimeout> | null = null
    const processBatch = () => {
      const batch = resultsWithoutPosters.slice(index, index + MAX_CONCURRENT_EXTRACTIONS)
      batch.forEach(result => extractPosterForResult(result))
      index += MAX_CONCURRENT_EXTRACTIONS

      if (index < resultsWithoutPosters.length) {
        timeoutId = setTimeout(processBatch, EXTRACTION_DELAY_MS)
      }
    }

    processBatch()

    // Cleanup: clear pending timeout when effect re-runs or unmounts
    return () => {
      if (timeoutId) {
        clearTimeout(timeoutId)
      }
    }
  }, [allSearchResults, failedExtractions])

  // Source management state
  const [showSourceForm, setShowSourceForm] = useState(false)
  const [editingSource, setEditingSource] = useState<DownloadSource | null>(null)
  const [sourceForm, setSourceForm] = useState({
    name: '',
    type: 'jackett',
    url: '',
    api_key: '',
    enabled: true,
    priority: 0,
  })

  // Logs viewer state
  const [showLogs, setShowLogs] = useState(false)
  const [logs, setLogs] = useState<string[]>([])
  const [logsLoading, setLogsLoading] = useState(false)

  // Load selected indexers from localStorage
  useEffect(() => {
    const stored = localStorage.getItem(INDEXERS_STORAGE_KEY)
    if (stored) {
      try {
        const parsed = JSON.parse(stored)
        setSelectedIndexers(new Set(parsed))
      } catch (e) {
        console.error('Failed to parse selected indexers:', e)
      }
    }
  }, [])

  // Save selected indexers to localStorage
  useEffect(() => {
    if (selectedIndexers.size > 0) {
      localStorage.setItem(INDEXERS_STORAGE_KEY, JSON.stringify(Array.from(selectedIndexers)))
    }
  }, [selectedIndexers])

  // Filter and sort search results
  const filteredResults = useMemo(() => {
    let results = [...allSearchResults]
    results = filterByQuality(results, qualityFilter)
    results = filterBySeeders(results, minSeedersFilter)
    results = filterBySize(results, minSizeFilter * 1024 * 1024 * 1024, maxSizeFilter * 1024 * 1024 * 1024)
    results = filterByCategory(results, categoryFilter)
    results = sortResults(results, sortBy, sortDirection)
    return results
  }, [allSearchResults, qualityFilter, minSeedersFilter, minSizeFilter, maxSizeFilter, categoryFilter, sortBy, sortDirection])

  // Group filtered results
  const groupedResults = useMemo(() => {
    if (!filteredResults || filteredResults.length === 0) return []
    return groupResults(filteredResults)
  }, [filteredResults])

  // Fetch search results with pagination
  const fetchSearchResults = async (query: string, offset: number, append: boolean = false) => {
    setIsSearching(true)
    try {
      const response = await fetch(`/api/search?q=${encodeURIComponent(query)}&indexers=${Array.from(selectedIndexers).join(',')}&limit=${RESULTS_PER_PAGE}&offset=${offset}`)

      if (!response.ok) {
        const errorText = await response.text()
        console.error(`Search API error (${response.status}):`, errorText)
        addToast(`Search failed: ${response.status} error`, 'error')
        return
      }

      const data = await response.json()

      if (data.results) {
        if (append) {
          setAllSearchResults(prev => [...prev, ...data.results])
        } else {
          setAllSearchResults(data.results)
        }

        const currentTotal = append ? offset + data.results.length : data.results.length
        setHasMoreResults(currentTotal < (data.total || 0))
      }
    } catch (error) {
      console.error('Search failed:', error)
      addToast('Search failed: Network error', 'error')
    } finally {
      setIsSearching(false)
    }
  }

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setCurrentOffset(0)
    setHasMoreResults(true)
    fetchSearchResults(searchQuery, 0, false)
    if (searchQuery.trim()) {
      addToHistory(searchQuery)
    }
  }

  const handleLoadMore = () => {
    const newOffset = currentOffset + RESULTS_PER_PAGE
    setCurrentOffset(newOffset)
    fetchSearchResults(searchQuery, newOffset, true)
  }

  // Infinite scroll with IntersectionObserver
  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && hasMoreResults && !isSearching && searchQuery) {
          handleLoadMore()
        }
      },
      { root: scrollContainerRef.current, rootMargin: '400px' }
    )

    if (loadMoreRef.current) {
      observer.observe(loadMoreRef.current)
    }

    return () => observer.disconnect()
  }, [hasMoreResults, isSearching, searchQuery, currentOffset])

  const { data: sources } = useSources()
  const { data: searchActivity } = useSearchActivity()
  const { data: allJackettIndexers, isLoading: indexersLoading, error: indexersError } = useAllJackettIndexers()
  const { addToast } = useAppStore()

  // Default all indexers to checked when they load
  useEffect(() => {
    if (allJackettIndexers && allJackettIndexers.sources?.length > 0) {
      const stored = localStorage.getItem(INDEXERS_STORAGE_KEY)
      if (!stored) {
        const allIndexerIds = new Set<string>()
        allJackettIndexers.sources.forEach((source: any) => {
          source.indexers?.forEach((indexer: any) => {
            allIndexerIds.add(indexer.id)
          })
        })
        setSelectedIndexers(allIndexerIds)
      }
    }
  }, [allJackettIndexers])

  const createSource = useCreateSource()
  const updateSource = useUpdateSource()
  const deleteSource = useDeleteSource()

  const getErrorMessage = (error: any): string => {
    return error?.userMessage || error?.message || 'Unknown error'
  }

  const handleDownload = (magnetLink: string) => {
    window.open(magnetLink, '_blank')
    addToast('Opening magnet link...', 'success')
  }

  const toggleGroup = (groupKey: string) => {
    setExpandedGroups(prev => {
      const next = new Set(prev)
      if (next.has(groupKey)) {
        next.delete(groupKey)
      } else {
        next.add(groupKey)
      }
      return next
    })
  }

  const openDetailModal = (torrent: any) => {
    setSelectedTorrent(torrent)
    setDetailModalOpen(true)
  }

  // Source management handlers
  const openAddSource = () => {
    setEditingSource(null)
    setSourceForm({
      name: '',
      type: 'jackett',
      url: '',
      api_key: '',
      enabled: true,
      priority: 0,
    })
    setShowSourceForm(true)
  }

  const openEditSource = (source: DownloadSource) => {
    setEditingSource(source)
    setSourceForm({
      name: source.name,
      type: source.type,
      url: source.url,
      api_key: source.api_key || '',
      enabled: source.enabled,
      priority: source.priority,
    })
    setShowSourceForm(true)
  }

  const handleSaveSource = async () => {
    try {
      if (!sourceForm.name || !sourceForm.type) {
        addToast('Name and type are required', 'error')
        return
      }

      if (editingSource) {
        await updateSource.mutateAsync({
          ...editingSource,
          ...sourceForm,
        })
        addToast('Source updated successfully', 'success')
      } else {
        await createSource.mutateAsync(sourceForm)
        addToast('Source created successfully', 'success')
      }
      setShowSourceForm(false)
    } catch (error: any) {
      addToast(`Failed to save source: ${getErrorMessage(error)}`, 'error')
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this source?')) return
    try {
      await deleteSource.mutateAsync(id)
      addToast('Source deleted successfully', 'success')
    } catch (error: any) {
      addToast(`Failed to delete source: ${getErrorMessage(error)}`, 'error')
    }
  }

  const fetchLogs = async () => {
    setLogsLoading(true)
    try {
      const response = await fetch('/api/logs?lines=30&filter=poster')
      if (response.ok) {
        const data = await response.json()
        setLogs(data.logs || [])
      }
    } catch (error) {
      console.error('Failed to fetch logs:', error)
    } finally {
      setLogsLoading(false)
    }
  }

  useEffect(() => {
    if (showLogs) {
      fetchLogs()
      const interval = setInterval(fetchLogs, 3000)
      return () => clearInterval(interval)
    }
  }, [showLogs])

  const enabledSources = sources?.filter((s: any) => s.enabled) || []

  const totalEnabledIndexers =
    allJackettIndexers?.sources?.reduce((total: number, source: any) => {
      return (
        total +
        (source.indexers?.filter((i: any) => i.status === 'enabled').length || 0)
      )
    }, 0) || 0

  // Get unique categories from results
  const categories = useMemo(() => {
    const cats = new Set<string>()
    allSearchResults.forEach(r => {
      if (r.category) cats.add(r.category)
    })
    return Array.from(cats).sort()
  }, [allSearchResults])

  return (
    <div className="flex gap-6 h-[calc(100vh-8rem)]">
      {/* Main Content */}
      <div ref={scrollContainerRef} className="flex-1 min-w-0 space-y-6 overflow-y-auto pr-2">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Search</h1>
            <p className="text-muted-foreground">
              Search torrents across all indexers
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowActivityLog(!showActivityLog)}
            >
              <Activity className="mr-2 h-4 w-4" />
              {showActivityLog ? 'Hide Activity' : 'Search Activity'}
            </Button>
          </div>
        </div>

        {/* Active Indexers */}
        {enabledSources.length > 0 && (
          <div className="flex items-center gap-2 flex-wrap">
            <Radio className="h-4 w-4 text-green-500" />
            <span className="text-sm text-muted-foreground">Active indexers:</span>
            {enabledSources.map((source: any) => (
              <Badge key={source.id} variant="outline" className="text-xs">
                {source.name}
              </Badge>
            ))}
          </div>
        )}

        {/* Search Activity Log */}
        {showActivityLog && (
          <Card>
            <CardContent className="p-4">
              <h3 className="font-semibold text-sm mb-3">Recent Search Activity</h3>
              {searchActivity && searchActivity.length > 0 ? (
                <div className="space-y-2 max-h-64 overflow-y-auto">
                  {searchActivity.map((activity: any, idx: number) => (
                    <div key={idx} className="text-sm p-2 rounded bg-muted/50">
                      <div className="flex items-center justify-between">
                        <div className="flex-1 min-w-0">
                          <span className="font-medium truncate">{activity.query}</span>
                          <span className="text-muted-foreground ml-2">
                            {activity.result_count} results
                            {activity.saved_count > 0 && ` (${activity.saved_count} saved)`}
                          </span>
                          {activity.error && (
                            <span className="text-red-500 ml-2">Error: {activity.error}</span>
                          )}
                        </div>
                        <div className="flex items-center gap-2 text-xs text-muted-foreground shrink-0">
                          <span>{activity.duration_seconds.toFixed(2)}s</span>
                          <span>{new Date(activity.timestamp).toLocaleTimeString()}</span>
                        </div>
                      </div>
                      {activity.provider_timings && activity.provider_timings.length > 0 && (
                        <div className="mt-1 flex gap-2 flex-wrap">
                          {activity.provider_timings.map((timing: any, tidx: number) => (
                            <span key={tidx} className="text-xs text-muted-foreground">
                              {timing.name}: {timing.duration_seconds.toFixed(2)}s
                              {timing.error && <span className="text-red-500"> (err)</span>}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">No recent search activity</p>
              )}
            </CardContent>
          </Card>
        )}

        {/* Logs Viewer */}
        <div className="flex justify-end">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowLogs(!showLogs)}
            className="text-xs text-muted-foreground hover:text-foreground"
          >
            {showLogs ? 'Hide Logs' : 'Show Logs'}
          </Button>
        </div>

        {showLogs && (
          <Card>
            <CardContent className="p-3">
              <div className="flex items-center justify-between mb-2">
                <h3 className="font-semibold text-xs text-muted-foreground uppercase tracking-wider">
                  Recent Logs
                </h3>
                {logsLoading && <Loader2 className="h-3 w-3 animate-spin text-muted-foreground" />}
              </div>
              <div className="bg-black/90 rounded p-2 max-h-48 overflow-y-auto font-mono text-[10px] leading-tight">
                {logs.length > 0 ? (
                  logs.map((log, idx) => (
                    <div
                      key={idx}
                      className={`${
                        log.includes('[ERROR]') || log.includes('failed')
                          ? 'text-red-400'
                          : log.includes('[SEARCH]')
                          ? 'text-green-400'
                          : log.includes('[TMDB]')
                          ? 'text-blue-400'
                          : log.includes('[TORRENT]')
                          ? 'text-yellow-400'
                          : 'text-gray-400'
                      }`}
                    >
                      {log}
                    </div>
                  ))
                ) : (
                  <span className="text-gray-600">No logs available</span>
                )}
              </div>
            </CardContent>
          </Card>
        )}

        {/* Search Form */}
        <form onSubmit={handleSearch} className="flex gap-4">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search torrents..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9"
            />
            <SearchHistory
              onSelect={(query) => {
                setSearchQuery(query)
                setCurrentOffset(0)
                setHasMoreResults(true)
                fetchSearchResults(query, 0, false)
              }}
              currentQuery={searchQuery}
            />
          </div>
          <Button type="submit" disabled={isSearching}>
            {isSearching ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Search className="mr-2 h-4 w-4" />
            )}
            {searchQuery.trim() ? 'Search' : 'Browse All'}
          </Button>
        </form>

        {/* Filters and Sort */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowFilters(!showFilters)}
            >
              <Filter className="mr-2 h-4 w-4" />
              Filters
              {(qualityFilter !== 'all' || minSeedersFilter > 0 || minSizeFilter > 0 || maxSizeFilter > 0 || categoryFilter !== 'all') && (
                <Badge variant="secondary" className="ml-1">Active</Badge>
              )}
            </Button>

            {viewMode === 'grid' && (
              <div className="flex items-center gap-1">
                <ArrowUpDown className="h-4 w-4 text-muted-foreground" />
                <select
                  value={sortBy}
                  onChange={(e) => {
                    setSortBy(e.target.value)
                    setSortDirection('desc')
                  }}
                  className="h-9 rounded-md border border-input bg-background px-3 py-1 text-sm"
                >
                  <option value="seeders">Seeders</option>
                  <option value="size">Size</option>
                  <option value="name">Name</option>
                  <option value="date">Date</option>
                  <option value="leechers">Leechers</option>
                  <option value="quality">Quality</option>
                </select>
              </div>
            )}
          </div>

          <div className="flex gap-1 p-1">
            <button
              onClick={() => setViewMode('grid')}
              className={`p-2 rounded ${viewMode === 'grid' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-muted'}`}
              title="Grid view"
            >
              <Grid3X3 className="h-4 w-4" />
            </button>
            <button
              onClick={() => setViewMode('table')}
              className={`p-2 rounded ${viewMode === 'table' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-muted'}`}
              title="Table view"
            >
              <List className="h-4 w-4" />
            </button>
          </div>
        </div>

        {/* Filter Panel */}
        {showFilters && (
          <Card>
            <CardContent className="p-4 space-y-4">
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div>
                  <label className="text-sm font-medium mb-1 block">Quality</label>
                  <select
                    value={qualityFilter}
                    onChange={(e) => setQualityFilter(e.target.value)}
                    className="w-full h-9 rounded-md border border-input bg-background px-3 py-1 text-sm"
                  >
                    <option value="all">All</option>
                    <option value="2160p">4K/2160p</option>
                    <option value="1080p">1080p</option>
                    <option value="720p">720p</option>
                    <option value="webdl">WEB-DL</option>
                    <option value="bluray">BluRay</option>
                    <option value="webrip">WEBRip</option>
                  </select>
                </div>

                <div>
                  <label className="text-sm font-medium mb-1 block">Min Seeders</label>
                  <Input
                    type="number"
                    value={minSeedersFilter || ''}
                    onChange={(e) => setMinSeedersFilter(parseInt(e.target.value) || 0)}
                    placeholder="0"
                  />
                </div>

                <div>
                  <label className="text-sm font-medium mb-1 block">Min Size (GB)</label>
                  <Input
                    type="number"
                    value={minSizeFilter || ''}
                    onChange={(e) => setMinSizeFilter(parseFloat(e.target.value) || 0)}
                    placeholder="0"
                    step="0.1"
                  />
                </div>

                <div>
                  <label className="text-sm font-medium mb-1 block">Max Size (GB)</label>
                  <Input
                    type="number"
                    value={maxSizeFilter || ''}
                    onChange={(e) => setMaxSizeFilter(parseFloat(e.target.value) || 0)}
                    placeholder="∞"
                    step="0.1"
                  />
                </div>
              </div>

              {categories.length > 0 && (
                <div>
                  <label className="text-sm font-medium mb-1 block">Category</label>
                  <select
                    value={categoryFilter}
                    onChange={(e) => setCategoryFilter(e.target.value)}
                    className="w-full h-9 rounded-md border border-input bg-background px-3 py-1 text-sm"
                  >
                    <option value="all">All Categories</option>
                    {categories.map(cat => (
                      <option key={cat} value={cat}>{cat}</option>
                    ))}
                  </select>
                </div>
              )}

              <div className="flex justify-end">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    setQualityFilter('all')
                    setMinSeedersFilter(0)
                    setMinSizeFilter(0)
                    setMaxSizeFilter(0)
                    setCategoryFilter('all')
                  }}
                >
                  Reset Filters
                </Button>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Grid View */}
        {viewMode === 'grid' && (
          <div className="grid gap-4 grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6">
            {groupedResults?.map((group: GroupedResult, groupIndex: number) => {
              const groupKey = `${group.baseTitle}_${group.year}`
              const isExpanded = expandedGroups.has(groupKey)
              const topVariant = group.variants[0]
              const variantCount = group.variants.length

              return (
                <div
                  key={groupIndex}
                  className="relative aspect-[2/3] rounded-lg overflow-hidden group cursor-pointer"
                  onClick={() => openDetailModal({
                    ...topVariant,
                    poster_url: group.poster_url,
                    source: topVariant.source,
                  })}
                  onContextMenu={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                    setContextMenuPos({ x: e.clientX, y: e.clientY })
                    setContextMenuGroup(group)
                    setContextMenuOpen(true)
                  }}
                >
                  {/* Background image with lazy loading and skeleton */}
                  <PosterImage
                    src={group.poster_url}
                    alt={group.baseTitle}
                    title={group.baseTitle}
                    className="absolute inset-0 w-full h-full transition-transform duration-300 group-hover:scale-105"
                  />

                  {/* Gradient overlay for text readability */}
                  <div className="absolute inset-0 bg-gradient-to-t from-black/90 via-black/40 to-transparent" />

                  {/* Top badges */}
                  <div className="absolute top-2 left-2 right-2 flex gap-1 flex-wrap">
                    {group.year && (
                      <Badge className="bg-yellow-600/80 text-white border-0 text-[10px] backdrop-blur-sm">
                        {group.year}
                      </Badge>
                    )}
                    <Badge className="bg-black/60 text-white border-0 text-[10px] backdrop-blur-sm">
                      {variantCount} version{variantCount > 1 ? 's' : ''}
                    </Badge>
                    {topVariant.indexer && (
                      <Badge className="bg-blue-600/80 text-white border-0 text-[10px] backdrop-blur-sm">
                        {topVariant.indexer}
                      </Badge>
                    )}
                    {variantCount > 1 && (
                      <Badge className="bg-purple-600/80 text-white border-0 text-[10px] backdrop-blur-sm">
                        {getQualityLabel(topVariant.title)}
                      </Badge>
                    )}
                  </div>

                  {/* Bottom content */}
                  <div className="absolute bottom-0 left-0 right-0 p-3">
                    <h3 className="text-white font-semibold text-sm leading-tight line-clamp-2 mb-1.5 drop-shadow-lg">
                      {group.baseTitle}
                    </h3>

                    {/* Top variant summary */}
                    <div className="flex items-center gap-1.5 flex-wrap mb-1.5">
                      <Badge className="bg-green-600/80 text-white border-0 text-[10px] backdrop-blur-sm">
                        {topVariant.seeders}S
                      </Badge>
                      <Badge className="bg-black/60 text-white border-0 text-[10px] backdrop-blur-sm">
                        {formatBytes(topVariant.size)}
                      </Badge>
                    </div>

                    {/* Expand/Collapse button */}
                    {variantCount > 1 && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          toggleGroup(groupKey)
                        }}
                        className="flex items-center gap-1 text-xs text-white/80 hover:text-white transition-colors"
                      >
                        {isExpanded ? (
                          <>
                            <ChevronUp className="h-3 w-3" />
                            Show less
                          </>
                        ) : (
                          <>
                            <ChevronDown className="h-3 w-3" />
                            Show all {variantCount} versions
                          </>
                        )}
                      </button>
                    )}
                  </div>

                  {/* Hover overlay with download button for top variant */}
                  <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                    <Button
                      size="sm"
                      className="bg-white text-black hover:bg-white/90"
                      onClick={(e) => {
                        e.stopPropagation()
                        topVariant.magnet_link && handleDownload(topVariant.magnet_link)
                      }}
                      disabled={!topVariant.magnet_link}
                    >
                      <Download className="h-4 w-4 mr-1" />
                      Download Best
                    </Button>
                  </div>

                  {/* Expanded variants list */}
                  {isExpanded && variantCount > 1 && (
                    <div className="absolute inset-x-0 bottom-0 top-1/3 bg-black/95 backdrop-blur-sm overflow-y-auto p-2 space-y-1 z-20">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-white text-xs font-semibold">All Versions</span>
                        <button
                          onClick={(e) => {
                            e.stopPropagation()
                            toggleGroup(groupKey)
                          }}
                          className="text-white/60 hover:text-white"
                        >
                          <X className="h-3 w-3" />
                        </button>
                      </div>
                      {group.variants.map((variant: SearchResult, vIndex: number) => (
                        <div
                          key={vIndex}
                          className="flex items-center justify-between p-1.5 rounded bg-white/10 hover:bg-white/20 cursor-pointer transition-colors"
                          onClick={(e) => {
                            e.stopPropagation()
                            variant.magnet_link && handleDownload(variant.magnet_link)
                          }}
                        >
                          <div className="min-w-0 flex-1 mr-2">
                            <p className="text-white text-[10px] truncate" title={variant.title}>
                              {variant.title}
                            </p>
                            <div className="flex items-center gap-1 mt-0.5 flex-wrap">
                              <span className="text-green-400 text-[9px]">{variant.seeders}S</span>
                              <span className="text-red-400 text-[9px]">{variant.leechers}L</span>
                              <span className="text-white/60 text-[9px]">{formatBytes(variant.size)}</span>
                              {vIndex === 0 && (
                                <span className="text-yellow-400 text-[9px] font-semibold">BEST</span>
                              )}
                            </div>
                          </div>
                          <div className="flex items-center gap-1 shrink-0">
                            <span className="text-[8px] text-white/40 bg-white/10 px-1 rounded">
                              {getQualityLabel(variant.title)}
                            </span>
                            <Download className="h-3 w-3 text-white/60" />
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        )}

        {/* Table View */}
        {viewMode === 'table' && (
          <div className="border rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-muted">
                <tr>
                  <th className="text-left p-3 font-medium w-14">Preview</th>
                  <th
                    className="text-left p-3 font-medium cursor-pointer hover:bg-muted/80 select-none"
                    onClick={() => {
                      if (sortBy === 'name') {
                        setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
                      } else {
                        setSortBy('name')
                        setSortDirection('asc')
                      }
                    }}
                  >
                    <div className="flex items-center gap-1">
                      Title
                      {sortBy === 'name' && (
                        <span className="text-xs">{sortDirection === 'asc' ? '↑' : '↓'}</span>
                      )}
                    </div>
                  </th>
                  <th
                    className="text-left p-3 font-medium cursor-pointer hover:bg-muted/80 select-none"
                    onClick={() => {
                      if (sortBy === 'size') {
                        setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
                      } else {
                        setSortBy('size')
                        setSortDirection('desc')
                      }
                    }}
                  >
                    <div className="flex items-center gap-1">
                      Size
                      {sortBy === 'size' && (
                        <span className="text-xs">{sortDirection === 'asc' ? '↑' : '↓'}</span>
                      )}
                    </div>
                  </th>
                  <th
                    className="text-left p-3 font-medium cursor-pointer hover:bg-muted/80 select-none"
                    onClick={() => {
                      if (sortBy === 'seeders') {
                        setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
                      } else {
                        setSortBy('seeders')
                        setSortDirection('desc')
                      }
                    }}
                  >
                    <div className="flex items-center gap-1">
                      Seeders
                      {sortBy === 'seeders' && (
                        <span className="text-xs">{sortDirection === 'asc' ? '↑' : '↓'}</span>
                      )}
                    </div>
                  </th>
                  <th
                    className="text-left p-3 font-medium cursor-pointer hover:bg-muted/80 select-none"
                    onClick={() => {
                      if (sortBy === 'leechers') {
                        setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
                      } else {
                        setSortBy('leechers')
                        setSortDirection('desc')
                      }
                    }}
                  >
                    <div className="flex items-center gap-1">
                      Leechers
                      {sortBy === 'leechers' && (
                        <span className="text-xs">{sortDirection === 'asc' ? '↑' : '↓'}</span>
                      )}
                    </div>
                  </th>
                  <th
                    className="text-left p-3 font-medium cursor-pointer hover:bg-muted/80 select-none"
                    onClick={() => {
                      if (sortBy === 'date') {
                        setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
                      } else {
                        setSortBy('date')
                        setSortDirection('desc')
                      }
                    }}
                  >
                    <div className="flex items-center gap-1">
                      Date
                      {sortBy === 'date' && (
                        <span className="text-xs">{sortDirection === 'asc' ? '↑' : '↓'}</span>
                      )}
                    </div>
                  </th>
                  <th
                    className="text-left p-3 font-medium cursor-pointer hover:bg-muted/80 select-none"
                    onClick={() => {
                      if (sortBy === 'quality') {
                        setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
                      } else {
                        setSortBy('quality')
                        setSortDirection('desc')
                      }
                    }}
                  >
                    <div className="flex items-center gap-1">
                      Quality
                      {sortBy === 'quality' && (
                        <span className="text-xs">{sortDirection === 'asc' ? '↑' : '↓'}</span>
                      )}
                    </div>
                  </th>
                  <th className="text-left p-3 font-medium">Indexer</th>
                  <th className="text-left p-3 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredResults?.length === 0 && !isSearching && (
                  <tr>
                    <td colSpan={9} className="p-8 text-center text-muted-foreground">
                      No results found. Try adjusting your search or filters.
                    </td>
                  </tr>
                )}
                {filteredResults?.map((result: SearchResult, idx: number) => (
                  <tr key={idx} className="border-t hover:bg-muted/50 cursor-pointer" onClick={() => openDetailModal(result)}>
                    <td className="p-2">
                      <div className="w-8 h-10 rounded overflow-hidden bg-muted flex items-center justify-center flex-shrink-0 relative">
                        {(result as any).poster_url || (result as any).posterUrl ? (
                          <img
                            src={(result as any).poster_url || (result as any).posterUrl}
                            alt={result.title}
                            className="w-full h-full object-cover"
                            loading="lazy"
                          />
                        ) : extractingPosters.has(result.info_hash) ? (
                          <Loader2 className="h-3 w-3 animate-spin text-muted-foreground" />
                        ) : (
                          <span className="text-[10px] font-bold text-muted-foreground">
                            {result.title.charAt(0).toUpperCase()}
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="p-2">
                      <div className="font-medium truncate max-w-[600px] lg:max-w-[500px] xl:max-w-[700px]" title={result.title}>{result.title}</div>
                    </td>
                    <td className="p-2 whitespace-nowrap">{formatBytes(result.size)}</td>
                    <td className="p-2">
                      <Badge className="bg-green-600 text-white">{result.seeders}</Badge>
                    </td>
                    <td className="p-2">{result.leechers}</td>
                    <td className="p-2 whitespace-nowrap">
                      {result.uploadDate ? new Date(result.uploadDate).toLocaleDateString() : '-'}
                    </td>
                    <td className="p-2">
                      {result.title.includes('2160p') || result.title.includes('4K') ? (
                        <Badge className="bg-purple-600 text-white">4K</Badge>
                      ) : result.title.includes('1080p') ? (
                        <Badge className="bg-blue-600 text-white">1080p</Badge>
                      ) : result.title.includes('720p') ? (
                        <Badge className="bg-yellow-600 text-white">720p</Badge>
                      ) : (
                        <span className="text-muted-foreground">-</span>
                      )}
                    </td>
                    <td className="p-2">
                      {result.indexer ? (
                        <Badge variant="outline" className="text-[10px]">{result.indexer}</Badge>
                      ) : (
                        <span className="text-muted-foreground">-</span>
                      )}
                    </td>
                    <td className="p-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={(e) => {
                          e.stopPropagation()
                          result.magnet_link && handleDownload(result.magnet_link)
                        }}
                        disabled={!result.magnet_link}
                      >
                        <Download className="h-3 w-3" />
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {filteredResults.length === 0 && !isSearching && searchQuery && (
              <div className="text-center py-8 text-muted-foreground">
                No results match your filters
              </div>
            )}
          </div>
        )}

        {/* Infinite scroll sentinel */}
        {groupedResults.length > 0 && hasMoreResults && (
          <div ref={loadMoreRef} className="flex justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        )}

        {allSearchResults.length === 0 && !isSearching && searchQuery && (
          <div className="text-center py-12">
            <p className="text-muted-foreground text-lg">No search results found</p>
            <p className="text-sm text-muted-foreground mt-1">Try a different search term</p>
          </div>
        )}
      </div>

      {/* Right Sidebar - Indexers */}
      <div className="w-80 shrink-0 border-l pl-6 overflow-y-auto">
        <div className="sticky top-0 bg-background pb-4 z-10">
          <div className="flex items-center gap-2 mb-4">
            <Server className="h-5 w-5 text-primary" />
            <h2 className="text-lg font-semibold">Indexer Sources</h2>
            {totalEnabledIndexers > 0 && (
              <Badge variant="secondary" className="ml-auto">
                {totalEnabledIndexers} active
              </Badge>
            )}
          </div>

          {/* Add Source Button */}
          <Button variant="outline" size="sm" className="w-full mb-4" onClick={openAddSource}>
            <Plus className="mr-2 h-4 w-4" />
            Add Source
          </Button>
        </div>

        {/* Configured Sources */}
        {sources && sources.length > 0 && (
          <div className="mb-6 space-y-2">
            <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
              Configured Sources
            </h3>
            {sources.map((source: DownloadSource) => (
              <div
                key={source.id}
                className="flex items-center justify-between p-2 rounded border bg-muted/20 text-sm"
              >
                <div className="flex-1 min-w-0 mr-2">
                  <div className="font-medium truncate">{source.name}</div>
                  <div className="text-xs text-muted-foreground truncate">{source.url}</div>
                </div>
                <div className="flex items-center gap-1 shrink-0">
                  <Badge
                    variant="outline"
                    className={`text-xs ${
                      source.enabled
                        ? 'bg-green-500/10 text-green-600 border-green-200'
                        : 'bg-gray-500/10 text-gray-600 border-gray-200'
                    }`}
                  >
                    {source.enabled ? 'ON' : 'OFF'}
                  </Badge>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 w-6 p-0"
                    onClick={() => openEditSource(source)}
                  >
                    <Pencil className="h-3 w-3" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 w-6 p-0 text-red-500 hover:text-red-600"
                    onClick={() => handleDelete(source.id)}
                  >
                    <Trash2 className="h-3 w-3" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Jackett Indexers */}
        {indexersLoading ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : indexersError ? (
          <div className="text-center py-8">
            <p className="text-sm text-red-500">Failed to load indexers</p>
            <p className="text-xs text-muted-foreground mt-1">
              {getErrorMessage(indexersError)}
            </p>
          </div>
        ) : allJackettIndexers && allJackettIndexers.sources?.length > 0 ? (
          <div className="space-y-6">
            {allJackettIndexers.sources.map((source: any) => (
              <div key={source.source} className="space-y-2">
                <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
                  {source.source}
                </h3>
                <div className="space-y-1">
                  {source.indexers.map((indexer: any) => (
                    <div
                      key={indexer.id}
                      className="flex items-center justify-between p-2 rounded border bg-muted/20 text-sm"
                    >
                      <div className="flex items-center gap-2 flex-1 min-w-0 mr-2">
                        <input
                          type="checkbox"
                          checked={selectedIndexers.has(indexer.id)}
                          onChange={(e) => {
                            const newSelected = new Set(selectedIndexers)
                            if (e.target.checked) {
                              newSelected.add(indexer.id)
                            } else {
                              newSelected.delete(indexer.id)
                            }
                            setSelectedIndexers(newSelected)
                          }}
                          className="h-4 w-4 rounded border-gray-300"
                        />
                        <div className="min-w-0">
                          <div className="font-medium truncate">{indexer.name}</div>
                        {indexer.error && (
                          <details className="text-xs text-red-500 mt-1">
                            <summary className="cursor-pointer hover:underline">
                              Error: {indexer.error.split('\n')[0].substring(0, 50)}
                              {indexer.error.split('\n')[0].length > 50 ? '...' : ''}
                            </summary>
                            <pre className="mt-1 p-2 bg-red-950/30 rounded text-[10px] overflow-x-auto whitespace-pre-wrap max-h-32 overflow-y-auto">
                              {indexer.error}
                            </pre>
                          </details>
                        )}
                      </div>
                      <div
                        className={`w-2.5 h-2.5 rounded-full shrink-0 ${
                          indexer.status === 'enabled'
                            ? 'bg-green-500'
                            : indexer.status === 'error'
                              ? 'bg-red-500'
                              : 'bg-gray-400'
                        }`}
                        title={indexer.status}
                      />
                    </div>
                  </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-center py-8">
            <p className="text-sm text-muted-foreground">No Jackett sources configured</p>
          </div>
        )}
      </div>

      {/* Source Form Modal */}
      {showSourceForm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <Card className="w-full max-w-md mx-4">
            <CardContent className="p-6">
              <h3 className="text-lg font-semibold mb-4">
                {editingSource ? 'Edit Source' : 'Add Source'}
              </h3>

              <div className="space-y-4">
                <div>
                  <label className="text-sm font-medium mb-1 block">Name</label>
                  <Input
                    value={sourceForm.name}
                    onChange={(e) => setSourceForm({ ...sourceForm, name: e.target.value })}
                    placeholder="My Jackett"
                  />
                </div>

                <div>
                  <label className="text-sm font-medium mb-1 block">Type</label>
                  <select
                    className="w-full h-10 rounded-md border border-input bg-background px-3 py-2 text-sm"
                    value={sourceForm.type}
                    onChange={(e) => setSourceForm({ ...sourceForm, type: e.target.value })}
                  >
                    <option value="jackett">Jackett</option>
                    <option value="rarbg">RARBG</option>
                  </select>
                </div>

                <div>
                  <label className="text-sm font-medium mb-1 block">URL</label>
                  <Input
                    value={sourceForm.url}
                    onChange={(e) => setSourceForm({ ...sourceForm, url: e.target.value })}
                    placeholder="http://localhost:9117"
                  />
                </div>

                <div>
                  <label className="text-sm font-medium mb-1 block">API Key</label>
                  <Input
                    type="password"
                    value={sourceForm.api_key}
                    onChange={(e) => setSourceForm({ ...sourceForm, api_key: e.target.value })}
                    placeholder="Optional"
                  />
                </div>

                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    id="enabled"
                    checked={sourceForm.enabled}
                    onChange={(e) => setSourceForm({ ...sourceForm, enabled: e.target.checked })}
                    className="h-4 w-4"
                  />
                  <label htmlFor="enabled" className="text-sm">
                    Enabled
                  </label>
                </div>

                <div>
                  <label className="text-sm font-medium mb-1 block">Priority</label>
                  <Input
                    type="number"
                    value={sourceForm.priority}
                    onChange={(e) =>
                      setSourceForm({ ...sourceForm, priority: parseInt(e.target.value) || 0 })
                    }
                    placeholder="0"
                  />
                </div>
              </div>

              <div className="flex gap-2 mt-6">
                <Button variant="outline" className="flex-1" onClick={() => setShowSourceForm(false)}>
                  Cancel
                </Button>
                <Button
                  className="flex-1"
                  onClick={handleSaveSource}
                  disabled={createSource.isPending || updateSource.isPending}
                >
                  {createSource.isPending || updateSource.isPending ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : null}
                  {editingSource ? 'Update' : 'Create'}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Torrent Detail Modal */}
      <TorrentDetailModal
        isOpen={detailModalOpen}
        onClose={() => setDetailModalOpen(false)}
        torrent={selectedTorrent}
        status={selectedTorrent?.status}
        onDownload={selectedTorrent?.magnet_link ? () => handleDownload(selectedTorrent.magnet_link) : undefined}
      />

      {/* Context Menu */}
      <ContextMenu
        isOpen={contextMenuOpen}
        x={contextMenuPos.x}
        y={contextMenuPos.y}
        group={contextMenuGroup}
        onClose={() => setContextMenuOpen(false)}
        onDownload={handleDownload}
        onSelectVariant={(variant) => {
          openDetailModal(variant)
        }}
      />
    </div>
  )
}

export { SearchPage as Search }
