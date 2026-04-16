// Auto-generated TypeScript types from Go models
// Matches: /pkg/models/download_models.go and /pkg/models/models.go

export interface DownloadSource {
  id: number;
  name: string;
  type: string;
  url: string;
  api_key?: string;
  enabled: boolean;
  priority: number;
  last_checked: string;
  last_error?: string;
  created_at: string;
  updated_at: string;
}

export interface DownloadRule {
  id: number;
  name: string;
  enabled: boolean;
  media_type: string;
  search_query?: string;
  quality?: string;
  resolution?: string;
  min_seeders: number;
  max_seeders?: number;
  min_size?: number;
  max_size?: number;
  min_upload_age?: number;
  max_upload_age?: number;
  sort_by: string;
  max_results: number;
  max_results_per_title: number;
  auto_download: boolean;
  destination_path: string;
  schedule?: string;
  last_run: string;
  created_at: string;
  updated_at: string;
}

export interface DownloadTask {
  id: number;
  rule_id?: number;
  rule?: DownloadRule;
  source_id?: number;
  source?: DownloadSource;
  title: string;
  info_hash?: string;
  magnet_link?: string;
  torrent_url?: string;
  torrent_id?: number;
  size: number;
  seeders: number;
  leechers: number;
  status: 'pending' | 'downloading' | 'completed' | 'failed' | 'cancelled';
  progress: number;
  download_path?: string;
  poster_url?: string;
  tmdb_id?: number;
  error?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface SearchResult {
  id: number;
  source_id: number;
  source?: DownloadSource;
  title: string;
  info_hash?: string;
  magnet_link?: string;
  torrent_url?: string;
  size: number;
  seeders: number;
  leechers: number;
  category?: string;
  upload_date?: string;
  expires_at: string;
  created_at: string;
}

export interface DownloadSuggestion {
  id: number;
  source_id: number;
  source?: DownloadSource;
  title: string;
  info_hash: string;
  magnet_link?: string;
  torrent_url?: string;
  size: number;
  seeders: number;
  leechers: number;
  category?: string;
  content_type?: string;
  resolution?: string;
  quality?: string;
  upload_date?: string;
  poster_url?: string;
  tmdb_id?: number;
  status: 'pending' | 'approved' | 'rejected' | 'downloaded';
  rejected_at?: string;
  approved_at?: string;
  downloaded_at?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface DownloadHistory {
  id: number;
  info_hash: string;
  title: string;
  size: number;
  downloaded_at: string;
  deleted_at: string;
  reason?: string;
  created_at: string;
}

export interface ServiceConfig {
  download_enabled: boolean;
  schedule_interval: number;
  max_concurrent_downloads: number;
  torrent_client_type: string;
  torrent_client_host: string;
  torrent_client_user?: string;
  torrent_client_pass?: string;
  jellyfin_url?: string;
  jellyfin_api_key?: string;
  download_path: string;
  library_path: string;
}

export interface MediaItem {
  title: string;
  year: number;
  poster_url: string;
  size: string;
  overview?: string;
  rating?: string;
  path: string;
}

export interface SuggestionStats {
  pending: number;
  approved: number;
  rejected: number;
  total: number;
}

export interface DownloadStats {
  pending: number;
  downloading: number;
  completed: number;
  failed: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface SearchParams {
  q?: string;
  status?: string;
  sort_by?: string;
  min_seeders?: number;
  limit?: number;
  offset?: number;
}

export interface ApiError {
  message: string;
  code?: number;
  details?: Record<string, unknown>;
}
