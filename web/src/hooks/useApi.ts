import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  suggestionsApi,
  tasksApi,
  libraryApi,
  sourcesApi,
  rulesApi,
  statsApi,
  vpnApi,
} from '@/lib/api'
import { SearchParams } from '@/types'

// Query keys for cache management
export const queryKeys = {
  suggestions: (params?: SearchParams) => ['suggestions', params] as const,
  suggestionStats: () => ['suggestionStats'] as const,
  tasks: () => ['tasks'] as const,
  library: () => ['library'] as const,
  sources: () => ['sources'] as const,
  rules: () => ['rules'] as const,
  stats: () => ['stats'] as const,
  vpnStatus: () => ['vpnStatus'] as const,
}

// Suggestions hooks
export function useSuggestions(params?: SearchParams) {
  return useQuery({
    queryKey: queryKeys.suggestions(params),
    queryFn: () => suggestionsApi.getSuggestions(params || {}),
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
    mutationFn: (id: number) => suggestionsApi.approve(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.suggestions() })
      queryClient.invalidateQueries({ queryKey: queryKeys.suggestionStats() })
      queryClient.invalidateQueries({ queryKey: queryKeys.tasks() })
    },
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

// Sources & Rules hooks
export function useSources() {
  return useQuery({
    queryKey: queryKeys.sources(),
    queryFn: () => sourcesApi.getSources(),
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

// VPN hook
export function useVPNStatus() {
  return useQuery({
    queryKey: queryKeys.vpnStatus(),
    queryFn: () => vpnApi.getStatus(),
    refetchInterval: 30000, // Refetch every 30 seconds
  })
}
