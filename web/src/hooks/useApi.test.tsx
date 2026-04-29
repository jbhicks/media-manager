import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactNode } from 'react'
import {
  queryKeys,
  useSearch,
  useTasks,
  useLibrary,
  useStats,
  useVPNStatus,
  useSources,
  useSuggestions,
  useApproveSuggestion,
  useRejectSuggestion,
  useCancelTask,
  useDeleteTask,
  useReprocessLibrary,
  useVPNConnect,
  useVPNDisconnect,
} from './useApi'

// Mock the API modules
vi.mock('@/lib/api', () => ({
  suggestionsApi: {
    getSuggestions: vi.fn(),
    getStats: vi.fn(),
    generate: vi.fn(),
    approve: vi.fn(),
    reject: vi.fn(),
    approveBatch: vi.fn(),
    rejectBatch: vi.fn(),
  },
  tasksApi: {
    getTasks: vi.fn(),
    cancel: vi.fn(),
    restart: vi.fn(),
    delete: vi.fn(),
    clearCompleted: vi.fn(),
    clearFailed: vi.fn(),
    reprocessCompleted: vi.fn(),
  },
  libraryApi: {
    getMovies: vi.fn(),
    reprocess: vi.fn(),
    fetchAllPosters: vi.fn(),
    refreshJellyfinPosters: vi.fn(),
  },
  sourcesApi: {
    getSources: vi.fn(),
    getJackettIndexers: vi.fn(),
    getAllJackettIndexers: vi.fn(),
    createSource: vi.fn(),
    updateSource: vi.fn(),
    deleteSource: vi.fn(),
  },
  rulesApi: {
    getRules: vi.fn(),
  },
  statsApi: {
    getStats: vi.fn(),
  },
  vpnApi: {
    getStatus: vi.fn(),
    connect: vi.fn(),
    disconnect: vi.fn(),
  },
  searchApi: {
    search: vi.fn(),
    getActivity: vi.fn(),
    extractImages: vi.fn(),
  },
}))

import {
  suggestionsApi,
  tasksApi,
  libraryApi,
  sourcesApi,
  statsApi,
  vpnApi,
  searchApi,
} from '@/lib/api'

function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  })
}

function Wrapper({ children }: { children: ReactNode }) {
  const queryClient = createTestQueryClient()
  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  )
}

describe('queryKeys', () => {
  it('should generate correct suggestion keys', () => {
    expect(queryKeys.suggestions()).toEqual(['suggestions', undefined])
    expect(queryKeys.suggestions({ q: 'test' })).toEqual(['suggestions', { q: 'test' }])
  })

  it('should generate correct task keys', () => {
    expect(queryKeys.tasks()).toEqual(['tasks'])
  })

  it('should generate correct library keys', () => {
    expect(queryKeys.library()).toEqual(['library'])
  })

  it('should generate correct search keys', () => {
    expect(queryKeys.search('batman')).toEqual(['search', 'batman', undefined, undefined, undefined])
    expect(queryKeys.search('batman', ['indexer1'], 20, 40)).toEqual([
      'search', 'batman', ['indexer1'], 20, 40,
    ])
  })

  it('should generate correct stats keys', () => {
    expect(queryKeys.stats()).toEqual(['stats'])
  })

  it('should generate correct vpn keys', () => {
    expect(queryKeys.vpnStatus()).toEqual(['vpnStatus'])
  })
})

describe('useSearch', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should not fetch when query is empty', () => {
    renderHook(() => useSearch(''), { wrapper: Wrapper })
    expect(searchApi.search).not.toHaveBeenCalled()
  })

  it('should fetch search results when query is provided', async () => {
    const mockResults = [{ id: 1, source_id: 1, title: 'Batman', size: 1000, seeders: 10, leechers: 5, expires_at: '', created_at: '' }]
    vi.mocked(searchApi.search).mockResolvedValue(mockResults as any)

    const { result } = renderHook(() => useSearch('batman'), { wrapper: Wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(searchApi.search).toHaveBeenCalledWith('batman', undefined, undefined, undefined)
    expect(result.current.data).toEqual(mockResults)
  })

  it('should pass indexers, limit, and offset to search', async () => {
    const mockResults: any[] = []
    vi.mocked(searchApi.search).mockResolvedValue(mockResults)

    const { result } = renderHook(
      () => useSearch('superman', ['idx1', 'idx2'], 50, 100),
      { wrapper: Wrapper }
    )

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(searchApi.search).toHaveBeenCalledWith('superman', ['idx1', 'idx2'], 50, 100)
  })
})

describe('useTasks', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch tasks', async () => {
    const mockTasks = [{ id: 1, title: 'Download 1', size: 1000, seeders: 10, leechers: 5, status: 'downloading', progress: 50, created_at: '', updated_at: '' }]
    vi.mocked(tasksApi.getTasks).mockResolvedValue(mockTasks as any)

    const { result } = renderHook(() => useTasks(), { wrapper: Wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(tasksApi.getTasks).toHaveBeenCalled()
    expect(result.current.data).toEqual(mockTasks)
  })
})

describe('useLibrary', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch movies', async () => {
    const mockMovies = [{ path: '/movies/test.mp4', title: 'Test Movie', year: 2024, poster_url: '', size: '1 GB' }]
    vi.mocked(libraryApi.getMovies).mockResolvedValue(mockMovies as any)

    const { result } = renderHook(() => useLibrary(), { wrapper: Wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(libraryApi.getMovies).toHaveBeenCalled()
    expect(result.current.data).toEqual(mockMovies)
  })
})

describe('useStats', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch stats', async () => {
    const mockStats = { pending: 5, downloading: 2, completed: 10, failed: 1 }
    vi.mocked(statsApi.getStats).mockResolvedValue(mockStats as any)

    const { result } = renderHook(() => useStats(), { wrapper: Wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(statsApi.getStats).toHaveBeenCalled()
    expect(result.current.data).toEqual(mockStats)
  })
})

describe('useVPNStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch VPN status', async () => {
    const mockStatus = { active: true, status: 'connected', message: 'VPN active' }
    vi.mocked(vpnApi.getStatus).mockResolvedValue(mockStatus as any)

    const { result } = renderHook(() => useVPNStatus(), { wrapper: Wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(vpnApi.getStatus).toHaveBeenCalled()
    expect(result.current.data).toEqual(mockStatus)
  })
})

describe('useSources', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch sources', async () => {
    const mockSources = [{ id: 1, name: 'Jackett', type: 'jackett', url: 'http://localhost', enabled: true, priority: 1, last_checked: '', created_at: '', updated_at: '' }]
    vi.mocked(sourcesApi.getSources).mockResolvedValue(mockSources as any)

    const { result } = renderHook(() => useSources(), { wrapper: Wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(sourcesApi.getSources).toHaveBeenCalled()
    expect(result.current.data).toEqual(mockSources)
  })
})

describe('useSuggestions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch suggestions without params', async () => {
    const mockSuggestions = { data: [{ id: 1, source_id: 1, title: 'Movie 1', info_hash: 'abc', size: 1000, seeders: 10, leechers: 5, status: 'pending', created_at: '', updated_at: '' }], total: 1, limit: 20, offset: 0 }
    vi.mocked(suggestionsApi.getSuggestions).mockResolvedValue(mockSuggestions as any)

    const { result } = renderHook(() => useSuggestions(), { wrapper: Wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(suggestionsApi.getSuggestions).toHaveBeenCalledWith({})
  })

  it('should fetch suggestions with params', async () => {
    const mockSuggestions = { data: [], total: 0, limit: 20, offset: 0 }
    vi.mocked(suggestionsApi.getSuggestions).mockResolvedValue(mockSuggestions as any)

    const { result } = renderHook(() => useSuggestions({ q: 'action' }), { wrapper: Wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(suggestionsApi.getSuggestions).toHaveBeenCalledWith({ q: 'action' })
  })
})

describe('useApproveSuggestion', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should call approve API', async () => {
    vi.mocked(suggestionsApi.approve).mockResolvedValue(undefined)

    const { result } = renderHook(() => useApproveSuggestion(), { wrapper: Wrapper })

    await result.current.mutateAsync({ id: 1, autoStart: true })
    expect(suggestionsApi.approve).toHaveBeenCalledWith(1, true)
  })
})

describe('useRejectSuggestion', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should call reject API', async () => {
    vi.mocked(suggestionsApi.reject).mockResolvedValue(undefined)

    const { result } = renderHook(() => useRejectSuggestion(), { wrapper: Wrapper })

    await result.current.mutateAsync(1)
    expect(suggestionsApi.reject).toHaveBeenCalledWith(1)
  })
})

describe('useCancelTask', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should call cancel API', async () => {
    vi.mocked(tasksApi.cancel).mockResolvedValue(undefined)

    const { result } = renderHook(() => useCancelTask(), { wrapper: Wrapper })

    await result.current.mutateAsync(1)
    expect(tasksApi.cancel).toHaveBeenCalledWith(1)
  })
})

describe('useDeleteTask', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should call delete API', async () => {
    vi.mocked(tasksApi.delete).mockResolvedValue(undefined)

    const { result } = renderHook(() => useDeleteTask(), { wrapper: Wrapper })

    await result.current.mutateAsync(1)
    expect(tasksApi.delete).toHaveBeenCalledWith(1)
  })
})

describe('useReprocessLibrary', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should call reprocess API', async () => {
    vi.mocked(libraryApi.reprocess).mockResolvedValue({ count: 5, message: 'Reprocessed 5 files' } as any)

    const { result } = renderHook(() => useReprocessLibrary(), { wrapper: Wrapper })

    await result.current.mutateAsync()
    expect(libraryApi.reprocess).toHaveBeenCalled()
  })
})

describe('useVPNConnect', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should call connect API', async () => {
    vi.mocked(vpnApi.connect).mockResolvedValue({ status: 'connected', message: 'Connected' } as any)

    const { result } = renderHook(() => useVPNConnect(), { wrapper: Wrapper })

    await result.current.mutateAsync()
    expect(vpnApi.connect).toHaveBeenCalled()
  })
})

describe('useVPNDisconnect', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should call disconnect API', async () => {
    vi.mocked(vpnApi.disconnect).mockResolvedValue({ status: 'disconnected', message: 'Disconnected' } as any)

    const { result } = renderHook(() => useVPNDisconnect(), { wrapper: Wrapper })

    await result.current.mutateAsync()
    expect(vpnApi.disconnect).toHaveBeenCalled()
  })
})