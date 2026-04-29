import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from '@tanstack/react-query'
import {
  suggestionsApi,
  tasksApi,
  libraryApi,
  sourcesApi,
  rulesApi,
  statsApi,
  vpnApi,
  searchApi,
} from '@/lib/api'
import { SearchParams, DownloadSource } from '@/types'

// Query keys for cache management
export const queryKeys = {
  suggestions: (params?: SearchParams) => ['suggestions', params] as const,
  groupedSuggestions: (params?: SearchParams) => ['groupedSuggestions', params] as const,
  suggestionStats: () => ['suggestionStats'] as const,
  tasks: () => ['tasks'] as const,
  library: () => ['library'] as const,
  sources: () => ['sources'] as const,
  rules: () => ['rules'] as const,
  stats: () => ['stats'] as const,
  vpnStatus: () => ['vpnStatus'] as const,
  search: (query: string, indexers?: string[], limit?: number, offset?: number) => ['search', query, indexers, limit, offset] as const,
}

// Suggestions hooks
export function useSuggestions(params?: SearchParams) {
  return useQuery({
    queryKey: queryKeys.suggestions(params),
    queryFn: () => suggestionsApi.getSuggestions(params || {}),
  })
}

export function useGroupedSuggestions(params?: SearchParams) {
  return useQuery({
    queryKey: ['groupedSuggestions', params],
    queryFn: () => suggestionsApi.getGroupedSuggestions(params || {}),
  })
}

export function useInfiniteGroupedSuggestions(params?: SearchParams) {
  return useInfiniteQuery({
    queryKey: ['infiniteGroupedSuggestions', params],
    queryFn: ({ pageParam = 0 }) =>
      suggestionsApi.getGroupedSuggestions({
        ...params,
        limit: 100,
        offset: pageParam,
      }),
    initialPageParam: 0,
    getNextPageParam: (lastPage) => {
      const nextOffset = lastPage.offset + lastPage.data.length
      return nextOffset < lastPage.total ? nextOffset : undefined
    },
  })
}

export function useSuggestionStats() {
  return useQuery({
    queryKey: queryKeys.suggestionStats(),
    queryFn: () => suggestionsApi.getStats(),
  })
}

export function useGenerateSuggestions() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: () => suggestionsApi.generate(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.suggestions() })
      queryClient.invalidateQueries({ queryKey: queryKeys.suggestionStats() })
    },
  })
}

export function useApproveSuggestion() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: ({ id, autoStart }: { id: number; autoStart?: boolean }) => suggestionsApi.approve(id, autoStart || false),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.suggestions() })
      queryClient.invalidateQueries({ queryKey: queryKeys.suggestionStats() })
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks() })
    },
  })
}

// Search hooks
export function useSearch(query: string, indexers?: string[], limit?: number, offset?: number) {
  return useQuery({
    queryKey: queryKeys.search(query, indexers, limit, offset),
    queryFn: () => searchApi.search(query, indexers, limit, offset),
    enabled: query.length > 0,
    staleTime: 30000,
    refetchIntervalInBackground: false,
  })
}

export function useSearchActivity() {
  return useQuery({
    queryKey: ['searchActivity'],
    queryFn: () => searchApi.getActivity(),
    refetchInterval: 5000,
  })
}

export function useExtractTorrentImages() {
  return useMutation({
    mutationFn: (magnetLink: string) => searchApi.extractImages(magnetLink),
  })
}

export function useRejectSuggestion() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: (id: number) => suggestionsApi.reject(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.suggestions() })
      queryClient.invalidateQueries({ queryKey: queryKeys.suggestionStats() })
    },
  })
}

export function useBulkApproveSuggestions() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: (ids: number[]) => suggestionsApi.approveBatch(ids),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.suggestions() })
      queryClient.invalidateQueries({ queryKey: queryKeys.suggestionStats() })
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks() })
    },
  })
}

export function useBulkRejectSuggestions() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: (ids: number[]) => suggestionsApi.rejectBatch(ids),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.suggestions() })
      queryClient.invalidateQueries({ queryKey: queryKeys.suggestionStats() })
    },
  })
}

// Tasks hooks
export function useTasks() {
  return useQuery({
    queryKey: queryKeys.tasks(),
    queryFn: () => tasksApi.getTasks(),
  })
}

export function useCancelTask() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: (id: number) => tasksApi.cancel(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks() })
      queryClient.invalidateQueries({ queryKey: queryKeys.stats() })
    },
  })
}

export function useRestartTask() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: (id: number) => tasksApi.restart(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks() })
      queryClient.invalidateQueries({ queryKey: queryKeys.stats() })
    },
  })
}

export function useDeleteTask() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: (id: number) => tasksApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks() })
      queryClient.invalidateQueries({ queryKey: queryKeys.stats() })
    },
  })
}

export function useClearCompletedTasks() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: () => tasksApi.clearCompleted(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks() })
      queryClient.invalidateQueries({ queryKey: queryKeys.stats() })
    },
  })
}

export function useClearFailedTasks() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: () => tasksApi.clearFailed(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks() })
      queryClient.invalidateQueries({ queryKey: queryKeys.stats() })
    },
  })
}

// Library hooks
export function useLibrary() {
  return useQuery({
    queryKey: queryKeys.library(),
    queryFn: () => libraryApi.getMovies(),
  })
}

export function useReprocessLibrary() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: () => libraryApi.reprocess(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.library() })
    },
  })
}

export function useFetchAllPosters() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => libraryApi.fetchAllPosters(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.library() })
    },
  })
}

export function useRefreshJellyfinPosters() {
  return useMutation({
    mutationFn: () => libraryApi.refreshJellyfinPosters(),
  })
}

// Sources & Rules hooks
export function useSources() {
  return useQuery({
    queryKey: queryKeys.sources(),
    queryFn: () => sourcesApi.getSources(),
  })
}

export function useJackettIndexers(sourceId: number | null) {
  return useQuery({
    queryKey: ['jackettIndexers', sourceId],
    queryFn: () => sourcesApi.getJackettIndexers(sourceId!),
    enabled: sourceId !== null,
  })
}

export function useAllJackettIndexers() {
  return useQuery({
    queryKey: ['allJackettIndexers'],
    queryFn: () => sourcesApi.getAllJackettIndexers(),
    refetchInterval: 30000, // Refetch every 30 seconds
  })
}

export function useCreateSource() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (source: {
      name: string;
      type: string;
      url: string;
      api_key?: string;
      enabled: boolean;
      priority: number;
    }) => sourcesApi.createSource(source),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sources() })
      queryClient.invalidateQueries({ queryKey: ['allJackettIndexers'] })
    },
  })
}

export function useUpdateSource() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (source: DownloadSource) => sourcesApi.updateSource(source),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sources() })
      queryClient.invalidateQueries({ queryKey: ['allJackettIndexers'] })
    },
  })
}

export function useDeleteSource() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => sourcesApi.deleteSource(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sources() })
      queryClient.invalidateQueries({ queryKey: ['allJackettIndexers'] })
    },
  })
}

export function useRules() {
  return useQuery({
    queryKey: queryKeys.rules(),
    queryFn: () => rulesApi.getRules(),
  })
}

// Stats hook
export function useStats() {
  return useQuery({
    queryKey: queryKeys.stats(),
    queryFn: () => statsApi.getStats(),
  })
}

// VPN hooks
export function useVPNStatus() {
  return useQuery({
    queryKey: queryKeys.vpnStatus(),
    queryFn: () => vpnApi.getStatus(),
    refetchInterval: 30000, // Refetch every 30 seconds
  })
}

export function useVPNConnect() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: () => vpnApi.connect(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.vpnStatus() })
    },
  })
}

export function useVPNDisconnect() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: () => vpnApi.disconnect(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.vpnStatus() })
    },
  })
}
