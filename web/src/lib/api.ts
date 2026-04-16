import axios, { AxiosError, AxiosInstance } from 'axios';
import {
  DownloadSource,
  DownloadRule,
  DownloadTask,
  DownloadSuggestion,
  SuggestionStats,
  DownloadStats,
  MediaItem,
  PaginatedResponse,
  SearchParams,
  ApiError,
} from '@/types';

const api: AxiosInstance = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Error interceptor
api.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ApiError>) => {
    const message = error.response?.data?.message || error.message || 'An error occurred';
    console.error('[API Error]', message);
    return Promise.reject(error);
  }
);

// Search API
export const searchApi = {
  search: async (query: string): Promise<void> => {
    await api.get('/search', { params: { q: query } });
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
};

// Suggestions API
export const suggestionsApi = {
  getSuggestions: async (params: SearchParams): Promise<PaginatedResponse<DownloadSuggestion>> => {
    const response = await api.get('/suggestions', { params });
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

  approve: async (id: number): Promise<void> => {
    await api.post(`/suggestions/approve?id=${id}`);
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
    return response.data;
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
    return response.data;
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
};

export default api;
