import { useState, useMemo, Fragment, useEffect, useRef } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import {
  useInfiniteGroupedSuggestions,
  useSuggestionStats,
  useGenerateSuggestions,
  useApproveSuggestion,
  useRejectSuggestion,
  useBulkApproveSuggestions,
  useBulkRejectSuggestions,
} from '@/hooks/useApi'
import { useAppStore } from '@/store/appStore'
import { PosterImage } from '@/components/PosterImage'
import { DownloadSuggestion, SuggestionGroup } from '@/types'
import { formatBytes } from '@/lib/utils'
import {
  Sparkles,
  Loader2,
  Check,
  X,
  Grid3X3,
  List,
  ArrowUpDown,
  Filter,
  ChevronDown,
  ChevronUp,
} from 'lucide-react'

type ViewMode = 'grid' | 'table'
type StatusFilter = 'all' | 'pending' | 'approved' | 'rejected'

export function Suggestions() {
  const [viewMode, setViewMode] = useState<ViewMode>('table')
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('pending')
  const [categoryFilter, setCategoryFilter] = useState('all')
  const [sortBy, setSortBy] = useState('seeders')
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc')
  const [selectedSuggestions, setSelectedSuggestions] = useState<Set<number>>(new Set())
  const [selectAll, setSelectAll] = useState(false)
  const [showFilters, setShowFilters] = useState(false)
  const [autoStartEnabled, setAutoStartEnabled] = useState(false)
  const [expandedGroups, setExpandedGroups] = useState<Set<number>>(new Set())

  const loadMoreRef = useRef<HTMLDivElement>(null)

  const { data: stats } = useSuggestionStats()
  const {
    data: infiniteData,
    isLoading: suggestionsLoading,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteGroupedSuggestions({
    status: statusFilter === 'all' ? undefined : statusFilter,
  })
  const { addToast } = useAppStore()

  const generateSuggestions = useGenerateSuggestions()
  const approveSuggestion = useApproveSuggestion()
  const rejectSuggestion = useRejectSuggestion()
  const bulkApprove = useBulkApproveSuggestions()
  const bulkReject = useBulkRejectSuggestions()

  const getErrorMessage = (error: any): string => {
    return error?.userMessage || error?.message || 'Unknown error'
  }

  // Flatten all pages into a single array
  const allGroups = useMemo(() => {
    if (!infiniteData?.pages) return []
    return infiniteData.pages.flatMap((page) => page.data)
  }, [infiniteData])

  // Intersection observer for infinite scroll
  useEffect(() => {
    if (!loadMoreRef.current) return

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasNextPage && !isFetchingNextPage) {
          fetchNextPage()
        }
      },
      { rootMargin: '200px' }
    )

    observer.observe(loadMoreRef.current)
    return () => observer.disconnect()
  }, [hasNextPage, isFetchingNextPage, fetchNextPage])

  // Get unique categories from all suggestions
  const suggestionCategories = useMemo(() => {
    if (!allGroups.length) return []
    const cats = new Set<string>()
    allGroups.forEach((group: SuggestionGroup) => {
      if (group.primary.category) cats.add(group.primary.category)
      group.alternates?.forEach((alt: DownloadSuggestion) => {
        if (alt.category) cats.add(alt.category)
      })
    })
    return Array.from(cats).sort()
  }, [allGroups])

  // Filter and sort groups (client-side on loaded data)
  const filteredGroups = useMemo(() => {
    let results = [...allGroups]

    if (categoryFilter !== 'all') {
      results = results.filter((group: SuggestionGroup) =>
        group.primary.category?.toLowerCase().includes(categoryFilter.toLowerCase()) ||
        group.alternates?.some((alt: DownloadSuggestion) =>
          alt.category?.toLowerCase().includes(categoryFilter.toLowerCase())
        )
      )
    }

    const mult = sortDirection === 'asc' ? 1 : -1
    results.sort((a: SuggestionGroup, b: SuggestionGroup) => {
      switch (sortBy) {
        case 'seeders':
          return mult * (a.primary.seeders - b.primary.seeders)
        case 'size':
          return mult * (a.primary.size - b.primary.size)
        case 'name':
          return mult * (a.title || a.primary.title).localeCompare(b.title || b.primary.title)
        case 'date':
          return mult * (new Date(a.primary.upload_date || 0).getTime() - new Date(b.primary.upload_date || 0).getTime())
        case 'quality': {
          const getQuality = (title: string) => {
            const t = title.toLowerCase()
            if (t.includes('2160p') || t.includes('4k')) return 5
            if (t.includes('1080p')) return 4
            if (t.includes('720p')) return 3
            if (t.includes('480p')) return 2
            return 1
          }
          return mult * (getQuality(a.primary.title) - getQuality(b.primary.title))
        }
        default:
          return mult * (a.primary.seeders - b.primary.seeders)
      }
    })

    return results
  }, [allGroups, categoryFilter, sortBy, sortDirection])

  const handleGenerate = async () => {
    try {
      await generateSuggestions.mutateAsync()
      addToast('Suggestions generated successfully', 'success')
    } catch (error: any) {
      addToast(`Failed to generate: ${getErrorMessage(error)}`, 'error')
    }
  }

  const handleApprove = async (id: number) => {
    try {
      await approveSuggestion.mutateAsync({ id, autoStart: autoStartEnabled })
      addToast(autoStartEnabled ? 'Torrent approved and auto-started' : 'Torrent approved for download', 'success')
    } catch (error: any) {
      addToast(`Failed to approve: ${getErrorMessage(error)}`, 'error')
    }
  }

  const handleReject = async (id: number) => {
    try {
      await rejectSuggestion.mutateAsync(id)
      addToast('Torrent rejected', 'success')
    } catch (error: any) {
      addToast(`Failed to reject: ${getErrorMessage(error)}`, 'error')
    }
  }

  const handleBulkApprove = async () => {
    if (selectedSuggestions.size === 0) return
    try {
      await bulkApprove.mutateAsync(Array.from(selectedSuggestions))
      addToast(`Approved ${selectedSuggestions.size} torrents`, 'success')
      setSelectedSuggestions(new Set())
      setSelectAll(false)
    } catch (error: any) {
      addToast(`Failed to bulk approve: ${getErrorMessage(error)}`, 'error')
    }
  }

  const handleBulkReject = async () => {
    if (selectedSuggestions.size === 0) return
    try {
      await bulkReject.mutateAsync(Array.from(selectedSuggestions))
      addToast(`Rejected ${selectedSuggestions.size} torrents`, 'success')
      setSelectedSuggestions(new Set())
      setSelectAll(false)
    } catch (error: any) {
      addToast(`Failed to bulk reject: ${getErrorMessage(error)}`, 'error')
    }
  }

  const toggleSuggestionSelection = (id: number) => {
    setSelectedSuggestions((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const handleSelectAll = () => {
    if (selectAll) {
      setSelectedSuggestions(new Set())
    } else {
      const allIds = filteredGroups.flatMap((group: SuggestionGroup) => [
        group.primary.id,
        ...(group.alternates || []).map((alt: DownloadSuggestion) => alt.id)
      ])
      setSelectedSuggestions(new Set(allIds))
    }
    setSelectAll(!selectAll)
  }

  const toggleGroupExpanded = (groupId: number) => {
    setExpandedGroups((prev) => {
      const next = new Set(prev)
      if (next.has(groupId)) {
        next.delete(groupId)
      } else {
        next.add(groupId)
      }
      return next
    })
  }

  const statusTabs: { label: string; value: StatusFilter; count?: number }[] = [
    { label: 'All', value: 'all', count: stats?.total },
    { label: 'Pending', value: 'pending', count: stats?.pending },
    { label: 'Approved', value: 'approved', count: stats?.approved },
    { label: 'Rejected', value: 'rejected', count: stats?.rejected },
  ]

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Suggestions</h1>
          <p className="text-muted-foreground">AI-powered download suggestions</p>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex items-center gap-2 mr-4">
            <input
              type="checkbox"
              id="autoStart"
              checked={autoStartEnabled}
              onChange={(e) => setAutoStartEnabled(e.target.checked)}
              className="h-4 w-4"
            />
            <label htmlFor="autoStart" className="text-sm text-muted-foreground">
              Auto-start on approve
            </label>
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

      {/* Status Tabs */}
      <div className="flex gap-2 border-b">
        {statusTabs.map((tab) => (
          <button
            key={tab.value}
            className={`px-4 py-2 ${
              statusFilter === tab.value
                ? 'border-b-2 border-primary font-medium'
                : 'text-muted-foreground hover:text-foreground'
            }`}
            onClick={() => {
              setStatusFilter(tab.value)
              setSelectedSuggestions(new Set())
              setSelectAll(false)
            }}
          >
            {tab.label}
            {tab.count !== undefined && ` (${tab.count})`}
          </button>
        ))}
      </div>

      {/* Controls */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowFilters(!showFilters)}
          >
            <Filter className="mr-2 h-4 w-4" />
            Filters
            {categoryFilter !== 'all' && (
              <Badge variant="secondary" className="ml-1">Active</Badge>
            )}
          </Button>

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
              <option value="quality">Quality</option>
            </select>
          </div>

          <button
            onClick={() => setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')}
            className="p-2 rounded text-muted-foreground hover:bg-muted"
            title={`Sort ${sortDirection === 'asc' ? 'descending' : 'ascending'}`}
          >
            {sortDirection === 'asc' ? '↑' : '↓'}
          </button>
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
            {suggestionCategories.length > 0 && (
              <div>
                <label className="text-sm font-medium mb-1 block">Category</label>
                <select
                  value={categoryFilter}
                  onChange={(e) => setCategoryFilter(e.target.value)}
                  className="w-full h-9 rounded-md border border-input bg-background px-3 py-1 text-sm"
                >
                  <option value="all">All Categories</option>
                  {suggestionCategories.map((cat) => (
                    <option key={cat} value={cat}>
                      {cat}
                    </option>
                  ))}
                </select>
              </div>
            )}
            <div className="flex justify-end">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setCategoryFilter('all')}
              >
                Reset Filters
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Bulk Actions Bar */}
      {selectedSuggestions.size > 0 && (
        <div className="flex items-center justify-between p-3 bg-muted rounded-lg">
          <span className="text-sm font-medium">
            {selectedSuggestions.size} selected
          </span>
          <div className="flex gap-2">
            <Button
              size="sm"
              className="bg-green-600 hover:bg-green-700"
              onClick={handleBulkApprove}
              disabled={bulkApprove.isPending}
            >
              <Check className="mr-2 h-4 w-4" />
              Approve Selected
            </Button>
            <Button
              size="sm"
              variant="outline"
              className="border-red-600 text-red-600"
              onClick={handleBulkReject}
              disabled={bulkReject.isPending}
            >
              <X className="mr-2 h-4 w-4" />
              Reject Selected
            </Button>
          </div>
        </div>
      )}

      {/* Select All */}
      {filteredGroups.length > 0 && (
        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={selectAll}
            onChange={handleSelectAll}
            className="h-4 w-4"
          />
          <span className="text-sm text-muted-foreground">Select all</span>
        </div>
      )}

      {/* Loading */}
      {suggestionsLoading && (
        <div className="flex justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      )}

      {/* Grid View */}
      {!suggestionsLoading && viewMode === 'grid' && (
        <div className="grid gap-4 grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6">
          {filteredGroups.map((group: SuggestionGroup) => (
            <div
              key={group.primary.id}
              className="relative aspect-[2/3] rounded-lg overflow-hidden group bg-slate-800 min-h-[200px]"
            >
              {/* Selection checkbox */}
              <div className="absolute top-2 left-2 z-10">
                <input
                  type="checkbox"
                  checked={selectedSuggestions.has(group.primary.id)}
                  onChange={() => toggleSuggestionSelection(group.primary.id)}
                  className="h-4 w-4 rounded border-gray-300"
                  onClick={(e) => e.stopPropagation()}
                />
              </div>

              {/* Background image - use group poster_url for best cover */}
              <PosterImage
                src={group.poster_url || group.primary.poster_url}
                alt={group.title || group.primary.title}
                title={group.title || group.primary.title}
                className="absolute inset-0 w-full h-full transition-transform duration-300 group-hover:scale-105"
              />

              {/* Gradient overlay */}
              <div className="absolute inset-0 bg-gradient-to-t from-black/90 via-black/40 to-transparent" />

              {/* Top badges */}
              <div className="absolute top-2 left-8 right-2 flex gap-1 flex-wrap">
                <Badge className="bg-black/60 text-white border-0 text-[10px] backdrop-blur-sm">
                  {formatBytes(group.primary.size)}
                </Badge>
                {group.total_options > 1 && (
                  <Badge className="bg-purple-600/80 text-white border-0 text-[10px] backdrop-blur-sm">
                    {group.total_options} versions
                  </Badge>
                )}
                {group.primary.source?.name && (
                  <Badge className="bg-blue-600/80 text-white border-0 text-[10px] backdrop-blur-sm">
                    {group.primary.source.name}
                  </Badge>
                )}
                <Badge className={`border-0 text-[10px] backdrop-blur-sm ${
                  group.primary.status === 'approved' ? 'bg-green-600 text-white' :
                  group.primary.status === 'rejected' ? 'bg-red-600 text-white' :
                  group.primary.status === 'downloaded' ? 'bg-blue-600 text-white' :
                  'bg-yellow-600 text-white'
                }`}>
                  {group.primary.status}
                </Badge>
              </div>

              {/* Bottom content */}
              <div className="absolute bottom-0 left-0 right-0 p-3">
                <h3 className="text-white font-semibold text-sm leading-tight line-clamp-2 mb-1.5 drop-shadow-lg">
                  {group.title || group.primary.title}
                </h3>

                <div className="flex items-center gap-1.5 flex-wrap">
                  <Badge className="bg-green-600/80 text-white border-0 text-[10px] backdrop-blur-sm">
                    {group.primary.seeders}S
                  </Badge>
                  <Badge className="bg-red-600/80 text-white border-0 text-[10px] backdrop-blur-sm">
                    {group.primary.leechers}L
                  </Badge>
                </div>
              </div>

              {/* Hover overlay with action buttons */}
              {group.primary.status === 'pending' && (
                <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                  <Button
                    size="sm"
                    className="bg-green-600 hover:bg-green-700 text-white"
                    onClick={() => handleApprove(group.primary.id)}
                    disabled={approveSuggestion.isPending}
                  >
                    <Check className="h-4 w-4 mr-1" />
                    Approve Best
                  </Button>
                  {(group.alternates || []).length > 0 && (
                    <Button
                      size="sm"
                      variant="outline"
                      className="border-white text-white hover:bg-white/20"
                      onClick={() => toggleGroupExpanded(group.primary.id)}
                    >
                      {expandedGroups.has(group.primary.id) ? (
                        <ChevronUp className="h-4 w-4 mr-1" />
                      ) : (
                        <ChevronDown className="h-4 w-4 mr-1" />
                      )}
                      Versions
                    </Button>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Table View */}
      {!suggestionsLoading && viewMode === 'table' && (
        <div className="border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-muted">
              <tr>
                <th className="text-left p-3 font-medium w-10">
                  <input
                    type="checkbox"
                    checked={selectAll}
                    onChange={handleSelectAll}
                    className="h-4 w-4"
                  />
                </th>
                <th className="text-left p-3 font-medium w-14">Preview</th>
                <th className="text-left p-3 font-medium">Title</th>
                <th className="text-left p-3 font-medium">Size</th>
                <th className="text-left p-3 font-medium">Seeders</th>
                <th className="text-left p-3 font-medium">Versions</th>
                <th className="text-left p-3 font-medium">Date</th>
                <th className="text-left p-3 font-medium">Quality</th>
                <th className="text-left p-3 font-medium">Category</th>
                <th className="text-left p-3 font-medium">Status</th>
                <th className="text-left p-3 font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredGroups.length === 0 && (
                <tr>
                  <td colSpan={11} className="p-8 text-center text-muted-foreground">
                    No suggestions found. Try adjusting your filters or generate new suggestions.
                  </td>
                </tr>
              )}
              {filteredGroups.map((group: SuggestionGroup) => (
                <Fragment key={group.primary.id}>
                  <tr className="border-t hover:bg-muted/50">
                    <td className="p-2">
                      <input
                        type="checkbox"
                        checked={selectedSuggestions.has(group.primary.id)}
                        onChange={() => toggleSuggestionSelection(group.primary.id)}
                        className="h-4 w-4"
                      />
                    </td>
                    <td className="p-2">
                      <div className="w-8 h-10 rounded overflow-hidden bg-muted flex items-center justify-center flex-shrink-0">
                        {(group.poster_url || group.primary.poster_url) ? (
                          <img
                            src={group.poster_url || group.primary.poster_url}
                            alt={group.title || group.primary.title}
                            className="w-full h-full object-cover"
                            loading="lazy"
                          />
                        ) : (
                          <span className="text-[10px] font-bold text-muted-foreground">
                            {(group.title || group.primary.title).charAt(0).toUpperCase()}
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="p-2">
                      <div className="font-medium truncate max-w-[400px]" title={group.title || group.primary.title}>
                        {group.title || group.primary.title}
                      </div>
                    </td>
                    <td className="p-2 whitespace-nowrap">{formatBytes(group.primary.size)}</td>
                    <td className="p-2">
                      <Badge className="bg-green-600 text-white">{group.primary.seeders}</Badge>
                    </td>
                    <td className="p-2">
                      {group.total_options > 1 ? (
                        <button
                          onClick={() => toggleGroupExpanded(group.primary.id)}
                          className="flex items-center gap-1 text-purple-400 hover:text-purple-300"
                        >
                          <Badge variant="outline" className="text-[10px]">
                            {group.total_options} versions
                          </Badge>
                          {expandedGroups.has(group.primary.id) ? (
                            <ChevronUp className="h-3 w-3" />
                          ) : (
                            <ChevronDown className="h-3 w-3" />
                          )}
                        </button>
                      ) : (
                        <span className="text-muted-foreground text-xs">1</span>
                      )}
                    </td>
                    <td className="p-2 whitespace-nowrap">
                      {group.primary.upload_date ? new Date(group.primary.upload_date).toLocaleDateString() : '-'}
                    </td>
                    <td className="p-2">
                      {group.primary.title.includes('2160p') || group.primary.title.includes('4K') ? (
                        <Badge className="bg-purple-600 text-white">4K</Badge>
                      ) : group.primary.title.includes('1080p') ? (
                        <Badge className="bg-blue-600 text-white">1080p</Badge>
                      ) : group.primary.title.includes('720p') ? (
                        <Badge className="bg-yellow-600 text-white">720p</Badge>
                      ) : (
                        <span className="text-muted-foreground">-</span>
                      )}
                    </td>
                    <td className="p-2">
                      {group.primary.category ? (
                        <Badge variant="outline" className="text-[10px]">
                          {group.primary.category}
                        </Badge>
                      ) : (
                        <span className="text-muted-foreground">-</span>
                      )}
                    </td>
                    <td className="p-2">
                      <Badge className={
                        group.primary.status === 'approved' ? 'bg-green-600 text-white' :
                        group.primary.status === 'rejected' ? 'bg-red-600 text-white' :
                        group.primary.status === 'downloaded' ? 'bg-blue-600 text-white' :
                        'bg-yellow-600 text-white'
                      }>
                        {group.primary.status}
                      </Badge>
                    </td>
                    <td className="p-2">
                      {group.primary.status === 'pending' && (
                        <div className="flex gap-1">
                          <Button
                            size="sm"
                            className="bg-green-600 hover:bg-green-700 text-white h-7 px-2"
                            onClick={() => handleApprove(group.primary.id)}
                            disabled={approveSuggestion.isPending}
                          >
                            <Check className="h-3 w-3" />
                          </Button>
                          <Button
                            size="sm"
                            className="bg-red-600 hover:bg-red-700 text-white h-7 px-2"
                            onClick={() => handleReject(group.primary.id)}
                            disabled={rejectSuggestion.isPending}
                          >
                            <X className="h-3 w-3" />
                          </Button>
                        </div>
                      )}
                    </td>
                  </tr>
                  {/* Expanded alternates */}
                  {expandedGroups.has(group.primary.id) && (group.alternates || []).length > 0 && (
                    <tr className="bg-muted/30">
                      <td colSpan={11} className="p-0">
                        <div className="p-3 space-y-2">
                          <div className="text-xs font-medium text-muted-foreground mb-2">
                            Alternate Versions
                          </div>
                          {(group.alternates || []).map((alt: DownloadSuggestion) => (
                            <div
                              key={alt.id}
                              className="flex items-center justify-between p-2 bg-background rounded border"
                            >
                              <div className="flex items-center gap-3 flex-1 min-w-0">
                                <input
                                  type="checkbox"
                                  checked={selectedSuggestions.has(alt.id)}
                                  onChange={() => toggleSuggestionSelection(alt.id)}
                                  className="h-4 w-4"
                                />
                                <div className="w-6 h-8 rounded overflow-hidden bg-muted flex-shrink-0">
                                  {alt.poster_url ? (
                                    <img
                                      src={alt.poster_url}
                                      alt={alt.title}
                                      className="w-full h-full object-cover"
                                      loading="lazy"
                                    />
                                  ) : (
                                    <span className="text-[8px] font-bold text-muted-foreground flex items-center justify-center h-full">
                                      {alt.title.charAt(0).toUpperCase()}
                                    </span>
                                  )}
                                </div>
                                <div className="truncate flex-1" title={alt.title}>
                                  <span className="font-medium text-sm">{alt.title}</span>
                                </div>
                              </div>
                              <div className="flex items-center gap-3 flex-shrink-0">
                                <Badge className="bg-green-600 text-white text-[10px]">{alt.seeders}S</Badge>
                                <span className="text-xs text-muted-foreground whitespace-nowrap">
                                  {formatBytes(alt.size)}
                                </span>
                                <div className="flex gap-1">
                                  <Button
                                    size="sm"
                                    className="bg-green-600 hover:bg-green-700 text-white h-6 px-2"
                                    onClick={() => handleApprove(alt.id)}
                                    disabled={approveSuggestion.isPending}
                                  >
                                    <Check className="h-3 w-3" />
                                  </Button>
                                  <Button
                                    size="sm"
                                    className="bg-red-600 hover:bg-red-700 text-white h-6 px-2"
                                    onClick={() => handleReject(alt.id)}
                                    disabled={rejectSuggestion.isPending}
                                  >
                                    <X className="h-3 w-3" />
                                  </Button>
                                </div>
                              </div>
                            </div>
                          ))}
                        </div>
                      </td>
                    </tr>
                  )}
                </Fragment>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Load More Sentinel */}
      {hasNextPage && (
        <div ref={loadMoreRef} className="flex justify-center py-8">
          {isFetchingNextPage ? (
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          ) : (
            <span className="text-sm text-muted-foreground">Scroll to load more...</span>
          )}
        </div>
      )}

      {/* Empty state */}
      {!suggestionsLoading && filteredGroups.length === 0 && (
        <div className="text-center py-12">
          <p className="text-muted-foreground text-lg">No suggestions found</p>
          <p className="text-sm text-muted-foreground mt-1">
            {statusFilter !== 'all'
              ? `No ${statusFilter} suggestions. Try a different filter or generate new suggestions.`
              : 'Generate suggestions to see results here.'}
          </p>
        </div>
      )}
    </div>
  )
}
