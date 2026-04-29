import axios, { AxiosError, AxiosInstance } from 'axios';
import {
  DownloadSource,
  DownloadRule,
  DownloadTask,
  DownloadSuggestion,
  SuggestionGroup,
  SuggestionStats,
  DownloadStats,
  MediaItem,
  PaginatedResponse,
  SearchParams,
} from '@/types';

const api: AxiosInstance = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Extract error message from various response formats
function extractErrorMessage(error: AxiosError<any>): string {
  // If backend sent JSON with message field
  if (error.response?.data?.message) {
    return error.response.data.message;
  }
  // If backend sent plain text (http.Error)
  if (typeof error.response?.data === 'string') {
    return error.response.data;
  }
  // Network or other errors
  if (error.message) {
    return error.message;
  }
  return 'An error occurred';
}

// Error interceptor
api.interceptors.response.use(
  (response) => response,
  (error: AxiosError<any>) => {
    const message = extractErrorMessage(error);
    console.error('[API Error]', message, error.response?.status, error.config?.url);
    // Attach extracted message to error for easy access
    (error as any).userMessage = message;
    return Promise.reject(error);
  }
);

export interface ProviderTiming {
  name: string;
  duration_seconds: number;
  result_count: number;
  error?: string;
}

export interface SearchActivity {
  query: string;
  timestamp: string;
  duration_seconds: number;
  result_count: number;
  saved_count: number;
  providers: string[];
  provider_timings: ProviderTiming[];
  error?: string;
}

export interface JackettIndexer {
  id: string;
  name: string;
  status: string;
  results?: number;
  error?: string;
  elapsed_ms?: number;
}

// Search API
export interface SearchResult {
  title: string;
  infoHash: string;
  magnetLink: string;
  torrentURL: string;
  size: number;
  seeders: number;
  leechers: number;
  category: string;
  uploadDate: string;
  posterUrl?: string;
  sourceName: string;
}

export const searchApi = {
  search: async (query: string, indexers?: string[], limit?: number, offset?: number): Promise<SearchResult[]> => {
    const params: any = { q: query };
    if (indexers && indexers.length > 0) {
      params.indexers = indexers.join(',');
    }
    if (limit) params.limit = limit;
    if (offset) params.offset = offset;
    const response = await api.get('/search', { params });
    return response.data.results || [];
  },

  getActivity: async (): Promise<SearchActivity[]> => {
    const response = await api.get('/search/activity');
    return response.data.activities || [];
  },

  getResults: async (params: SearchParams): Promise<PaginatedResponse<DownloadSuggestion>> => {
    const response = await api.get('/suggestions', { params });
    return response.data;
  },

  fetchPoster: async (id: number): Promise<string> => {
    const response = await api.get(`/search/poster`, { params: { id } });
    return response.data.poster_url;
  },

  approve: async (id: number): Promise<void> => {
    await api.post(`/search/approve?id=${id}`);
  },

  reject: async (id: number): Promise<void> => {
    await api.post(`/search/reject?id=${id}`);
  },

  bulkApprove: async (ids: number[]): Promise<void> => {
    const formData = new FormData();
    ids.forEach((id) => formData.append('selected', id.toString()));
    await api.post('/search/bulk-approve', formData);
  },

  bulkReject: async (ids: number[]): Promise<void> => {
    const formData = new FormData();
    ids.forEach((id) => formData.append('selected', id.toString()));
    await api.post('/search/bulk-reject', formData);
  },

  clearByStatus: async (status: string): Promise<void> => {
    await api.post(`/search/clear?status=${status}`);
  },

  clearRejected: async (): Promise<void> => {
    await api.post('/search/clear-rejected');
  },

  extractImages: async (magnetLink: string): Promise<{ images: string[]; count: number; supported: boolean; error?: string }> => {
    const response = await api.post('/search/extract-images', { magnet_link: magnetLink });
    return response.data;
  },

  getLogs: async (lines?: number, filter?: string): Promise<{ logs: string[]; total: number }> => {
    const response = await api.get('/logs', { params: { lines, filter } });
    return response.data;
  },
};

// Suggestions API
export const suggestionsApi = {
  getSuggestions: async (params: SearchParams): Promise<PaginatedResponse<DownloadSuggestion>> => {
    const response = await api.get('/suggestions', { params });
    return response.data;
  },

  getGroupedSuggestions: async (params: SearchParams): Promise<PaginatedResponse<SuggestionGroup>> => {
    const response = await api.get('/suggestions/grouped', { params });
    return response.data;
  },

  searchGroupedSuggestions: async (params: SearchParams): Promise<PaginatedResponse<SuggestionGroup>> => {
    const response = await api.get('/suggestions/grouped/search', { params });
    return response.data;
  },

  getStats: async (): Promise<SuggestionStats> => {
    const response = await api.get('/suggestions/stats');
    return response.data;
  },

  generate: async (): Promise<void> => {
    await api.post('/suggestions/generate');
  },

  searchSuggestions: async (params: SearchParams): Promise<PaginatedResponse<DownloadSuggestion>> => {
    const response = await api.get('/suggestions/search', { params });
    return response.data;
  },

  getRecommendations: async (limit: number = 5): Promise<{ recommendations: Array<{ suggestion: DownloadSuggestion; quality_score: number }> }> => {
    const response = await api.get('/suggestions/recommendations', { params: { limit } });
    return response.data;
  },

  getQualityScore: async (id: number): Promise<number> => {
    const response = await api.get('/suggestions/quality-score', { params: { id } });
    return response.data.quality_score;
  },

  approve: async (id: number, autoStart: boolean = false): Promise<void> => {
    await api.post(`/suggestions/approve?id=${id}&auto_start=${autoStart}`);
  },

  reject: async (id: number): Promise<void> => {
    await api.post(`/suggestions/reject?id=${id}`);
  },

  approveBatch: async (ids: number[]): Promise<void> => {
    const formData = new FormData();
    ids.forEach((id) => formData.append('selected', id.toString()));
    await api.post('/suggestions/approve-batch', formData);
  },

  rejectBatch: async (ids: number[]): Promise<void> => {
    const formData = new FormData();
    ids.forEach((id) => formData.append('selected', id.toString()));
    await api.post('/suggestions/reject-batch', formData);
  },

  clearRejected: async (): Promise<void> => {
    await api.post('/suggestions/clear-rejected');
  },

  refreshPosters: async (): Promise<{ success: number; failed: number }> => {
    const response = await api.post('/suggestions/refresh-posters');
    return response.data;
  },
};

// Tasks API
export const tasksApi = {
  getTasks: async (): Promise<DownloadTask[]> => {
    const response = await api.get('/tasks');
    return response.data.tasks || [];
  },

  cancel: async (id: number): Promise<void> => {
    await api.post('/tasks/cancel', { id });
  },

  restart: async (id: number): Promise<void> => {
    await api.post('/tasks/restart', { id });
  },

  delete: async (id: number): Promise<void> => {
    await api.post('/tasks/delete', { id });
  },

  clearCompleted: async (): Promise<void> => {
    await api.post('/tasks/clear-completed');
  },

  clearFailed: async (): Promise<void> => {
    await api.post('/tasks/clear-failed');
  },

  reprocess: async (): Promise<{ count: number; message: string }> => {
    const response = await api.post('/tasks/reprocess');
    return response.data;
  },
};

// Library API
export const libraryApi = {
  getMovies: async (): Promise<MediaItem[]> => {
    const response = await api.get('/library/movies');
    return response.data.movies || response.data;
  },

  fetchPoster: async (id: number): Promise<string> => {
    const response = await api.get('/library/poster', { params: { id } });
    return response.data.poster_url;
  },

  fetchPosterByTitle: async (title: string): Promise<string> => {
    const response = await api.get('/library/poster-by-title', { params: { title } });
    return response.data.poster_url;
  },

  fetchAllPosters: async (): Promise<{ fetched: number; cached: number; failed: number }> => {
    const response = await api.post('/library/fetch-all-posters');
    return response.data;
  },

  refreshJellyfinPosters: async (): Promise<{ success: number; failed: number; skipped: number }> => {
    const response = await api.post('/library/refresh-jellyfin-posters');
    return response.data;
  },

  reprocess: async (): Promise<{ count: number; message: string }> => {
    const response = await api.post('/library/reprocess');
    return response.data;
  },

  deleteMovie: async (id: number): Promise<void> => {
    await api.post('/library/delete', { id });
  },
};

// Sources API
export const sourcesApi = {
  getSources: async (): Promise<DownloadSource[]> => {
    const response = await api.get('/sources');
    return response.data.sources || [];
  },

  createSource: async (source: {
    name: string;
    type: string;
    url: string;
    api_key?: string;
    enabled: boolean;
    priority: number;
  }): Promise<DownloadSource> => {
    const response = await api.post('/sources/create', source);
    return response.data;
  },

  updateSource: async (source: DownloadSource): Promise<void> => {
    await api.post('/sources/update', source);
  },

  deleteSource: async (id: number): Promise<void> => {
    await api.post('/sources/delete', { id });
  },

  getJackettIndexers: async (sourceId: number): Promise<{ source: string; indexers: JackettIndexer[] }> => {
    const response = await api.get(`/sources/${sourceId}/jackett-indexers`);
    return response.data;
  },

  getAllJackettIndexers: async (): Promise<{ sources: Array<{ source: string; url: string; indexers: JackettIndexer[] }> }> => {
    const response = await api.get('/sources/jackett-indexers');
    return response.data;
  },
};

// Rules API
export const rulesApi = {
  getRules: async (): Promise<DownloadRule[]> => {
    const response = await api.get('/rules');
    return response.data;
  },
};

// Stats API
export const statsApi = {
  getStats: async (): Promise<DownloadStats> => {
    const response = await api.get('/stats');
    return response.data;
  },
};

// Movie details API
export const movieApi = {
  getDetails: async (id: number): Promise<unknown> => {
    const response = await api.get('/movie/details', { params: { id } });
    return response.data;
  },
};

// VPN API
export interface VPNStatus {
  active: boolean;
  status: 'connected' | 'disconnected';
  message: string;
  ip?: string;
  location?: string;
  country?: string;
  provider?: string;
  server?: string;
  type?: string;
}

export const vpnApi = {
  getStatus: async (): Promise<VPNStatus> => {
    const response = await api.get('/vpn/status');
    return response.data;
  },
  connect: async (): Promise<{ status: string; message: string }> => {
    const response = await api.post('/vpn/connect');
    return response.data;
  },
  disconnect: async (): Promise<{ status: string; message: string }> => {
    const response = await api.post('/vpn/disconnect');
    return response.data;
  },
};

export default api;
