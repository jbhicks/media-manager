import { describe, it, expect, vi, beforeEach } from 'vitest'

// Create mock functions BEFORE vi.mock (which is hoisted)
const mockGet = vi.fn()
const mockPost = vi.fn()
const mockPut = vi.fn()
const mockDelete = vi.fn()

// Mock axios before importing the API module
vi.mock('axios', () => ({
  default: {
    create: vi.fn(() => ({
      get: (...args: any[]) => mockGet(...args),
      post: (...args: any[]) => mockPost(...args),
      put: (...args: any[]) => mockPut(...args),
      delete: (...args: any[]) => mockDelete(...args),
      interceptors: {
        response: {
          use: vi.fn(),
        },
      },
    })),
  },
}))

// Import the API module AFTER the mock is set up
import {
  searchApi,
  suggestionsApi,
  tasksApi,
  libraryApi,
  statsApi,
  vpnApi,
} from './api'

describe('API Layer', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('searchApi', () => {
    it('searches with query parameter', async () => {
      mockGet.mockResolvedValue({ data: { results: [], query: 'test' } })
      await searchApi.search('test')
      expect(mockGet).toHaveBeenCalledWith('/search', { params: { q: 'test' } })
    })

    it('fetches poster by id', async () => {
      mockGet.mockResolvedValue({ data: { poster_url: 'http://example.com/poster.jpg' } })
      const result = await searchApi.fetchPoster(1)
      expect(mockGet).toHaveBeenCalledWith('/search/poster', { params: { id: 1 } })
      expect(result).toBe('http://example.com/poster.jpg')
    })

    it('approves a suggestion', async () => {
      mockPost.mockResolvedValue({ data: { success: true } })
      await searchApi.approve(1)
      expect(mockPost).toHaveBeenCalledWith('/search/approve?id=1')
    })

    it('rejects a suggestion', async () => {
      mockPost.mockResolvedValue({ data: { success: true } })
      await searchApi.reject(1)
      expect(mockPost).toHaveBeenCalledWith('/search/reject?id=1')
    })

    it('bulk approves suggestions', async () => {
      mockPost.mockResolvedValue({ data: { success: true } })
      await searchApi.bulkApprove([1, 2, 3])
      expect(mockPost).toHaveBeenCalledWith('/search/bulk-approve', expect.any(FormData))
    })
  })

  describe('suggestionsApi', () => {
    it('gets suggestions with params', async () => {
      mockGet.mockResolvedValue({
        data: { suggestions: [], total: 0, limit: 20, offset: 0 },
      })
      await suggestionsApi.getSuggestions({ status: 'pending', limit: 10 })
      expect(mockGet).toHaveBeenCalledWith('/suggestions', {
        params: { status: 'pending', limit: 10 },
      })
    })

    it('gets stats', async () => {
      mockGet.mockResolvedValue({
        data: { pending: 1, approved: 2, rejected: 3, total: 6 },
      })
      const result = await suggestionsApi.getStats()
      expect(mockGet).toHaveBeenCalledWith('/suggestions/stats')
      expect(result).toEqual({ pending: 1, approved: 2, rejected: 3, total: 6 })
    })

    it('generates suggestions', async () => {
      mockPost.mockResolvedValue({ data: { created: 5 } })
      await suggestionsApi.generate()
      expect(mockPost).toHaveBeenCalledWith('/suggestions/generate')
    })
  })

  describe('tasksApi', () => {
    it('gets tasks and extracts tasks array', async () => {
      const tasks = [
        { id: 1, title: 'Test', status: 'pending', progress: 0 },
      ]
      mockGet.mockResolvedValue({
        data: { tasks, stats: { pending: 1, downloading: 0, completed: 0, failed: 0 } },
      })
      const result = await tasksApi.getTasks()
      expect(mockGet).toHaveBeenCalledWith('/tasks')
      expect(result).toEqual(tasks)
    })

    it('returns empty array when tasks is missing', async () => {
      mockGet.mockResolvedValue({ data: { stats: {} } })
      const result = await tasksApi.getTasks()
      expect(result).toEqual([])
    })

    it('cancels a task', async () => {
      mockPost.mockResolvedValue({ data: { success: true } })
      await tasksApi.cancel(1)
      expect(mockPost).toHaveBeenCalledWith('/tasks/cancel', { id: 1 })
    })

    it('restarts a task', async () => {
      mockPost.mockResolvedValue({ data: { success: true } })
      await tasksApi.restart(1)
      expect(mockPost).toHaveBeenCalledWith('/tasks/restart', { id: 1 })
    })

    it('deletes a task', async () => {
      mockPost.mockResolvedValue({ data: { success: true } })
      await tasksApi.delete(1)
      expect(mockPost).toHaveBeenCalledWith('/tasks/delete', { id: 1 })
    })

    it('clears completed tasks', async () => {
      mockPost.mockResolvedValue({ data: { success: true } })
      await tasksApi.clearCompleted()
      expect(mockPost).toHaveBeenCalledWith('/tasks/clear-completed')
    })

    it('clears failed tasks', async () => {
      mockPost.mockResolvedValue({ data: { success: true } })
      await tasksApi.clearFailed()
      expect(mockPost).toHaveBeenCalledWith('/tasks/clear-failed')
    })

    it('reprocesses completed tasks', async () => {
      mockPost.mockResolvedValue({ data: { count: 5, message: 'Reprocessed 5' } })
      const result = await tasksApi.reprocess()
      expect(mockPost).toHaveBeenCalledWith('/tasks/reprocess')
      expect(result.count).toBe(5)
    })
  })

  describe('libraryApi', () => {
    it('gets movies', async () => {
      const movies = [{ title: 'Test Movie', year: 2024, poster_url: '', size: '1.5 GB', path: '/test' }]
      mockGet.mockResolvedValue({ data: movies })
      const result = await libraryApi.getMovies()
      expect(mockGet).toHaveBeenCalledWith('/library/movies')
      expect(result).toEqual(movies)
    })

    it('fetches poster by title', async () => {
      mockGet.mockResolvedValue({ data: { poster_url: 'http://example.com/poster.jpg' } })
      const result = await libraryApi.fetchPosterByTitle('Test Movie')
      expect(mockGet).toHaveBeenCalledWith('/library/poster-by-title', {
        params: { title: 'Test Movie' },
      })
      expect(result).toBe('http://example.com/poster.jpg')
    })

    it('reprocesses library', async () => {
      mockPost.mockResolvedValue({ data: { count: 3, message: 'Reprocessed 3' } })
      const result = await libraryApi.reprocess()
      expect(mockPost).toHaveBeenCalledWith('/library/reprocess')
      expect(result.count).toBe(3)
    })

    it('deletes a movie by id', async () => {
      mockPost.mockResolvedValue({ data: { success: true } })
      await libraryApi.deleteMovie(1)
      expect(mockPost).toHaveBeenCalledWith('/library/delete', { id: 1 })
    })
  })

  describe('statsApi', () => {
    it('gets stats', async () => {
      const stats = { pending: 1, downloading: 0, completed: 2, failed: 0 }
      mockGet.mockResolvedValue({ data: stats })
      const result = await statsApi.getStats()
      expect(mockGet).toHaveBeenCalledWith('/stats')
      expect(result).toEqual(stats)
    })
  })

  describe('vpnApi', () => {
    it('gets VPN status', async () => {
      const status = { active: true, status: 'connected', message: 'Connected' }
      mockGet.mockResolvedValue({ data: status })
      const result = await vpnApi.getStatus()
      expect(mockGet).toHaveBeenCalledWith('/vpn/status')
      expect(result).toEqual(status)
    })

    it('connects VPN', async () => {
      mockPost.mockResolvedValue({ data: { status: 'connecting', message: 'Connecting...' } })
      const result = await vpnApi.connect()
      expect(mockPost).toHaveBeenCalledWith('/vpn/connect')
      expect(result.status).toBe('connecting')
    })

    it('disconnects VPN', async () => {
      mockPost.mockResolvedValue({ data: { status: 'disconnected', message: 'Disconnected' } })
      const result = await vpnApi.disconnect()
      expect(mockPost).toHaveBeenCalledWith('/vpn/disconnect')
      expect(result.status).toBe('disconnected')
    })
  })
})
