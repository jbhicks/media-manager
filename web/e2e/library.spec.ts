import { test, expect } from '@playwright/test'

test.describe('Library Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/library')
  })

  test('should display library page title and description', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Library' })).toBeVisible()
    
    // Description with item count
    const description = await page.getByText(/Browse your media collection/i)
    await expect(description).toBeVisible()
  })

  test('should display Clean Filenames button', async ({ page }) => {
    const cleanButton = page.getByRole('button', { name: /Clean Filenames/i })
    await expect(cleanButton).toBeVisible()
  })

  test('should display movie grid or empty state', async ({ page }) => {
    // Wait for library to load
    await page.waitForResponse(response => 
      response.url().includes('/api/library/movies') && response.status() === 200
    )
    
    // Either movies are shown or empty state
    const hasMovies = await page.locator('.grid > div').count() > 0
    const hasEmptyState = await page.getByText('No media files found').isVisible()
    
    expect(hasMovies || hasEmptyState).toBeTruthy()
  })

  test('should display movie cards with posters when movies exist', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/library/movies') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.movies && data.movies.length > 0) {
      // Should show movie title
      const firstMovie = data.movies[0]
      await expect(page.getByText(firstMovie.title).first()).toBeVisible()
      
      // Should have poster image
      const poster = page.locator('img[alt="' + firstMovie.title + '"]').first()
      await expect(poster).toBeVisible()
    }
  })

  test('should display movie metadata when available', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/library/movies') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.movies && data.movies.length > 0) {
      const movie = data.movies[0]
      
      // Should show size
      if (movie.size) {
        await expect(page.getByText(movie.size).first()).toBeVisible()
      }
      
      // Should show rating if available
      if (movie.rating) {
        await expect(page.getByText(movie.rating).first()).toBeVisible()
      }
    }
  })

  test('should show correct aspect ratio for movie posters', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/library/movies') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.movies && data.movies.length > 0) {
      // Check for poster container with aspect ratio
      const posterContainer = page.locator('.aspect-\\[2\\/3\\]').first()
      await expect(posterContainer).toBeVisible()
    }
  })

  test('empty state should show helpful message', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/library/movies') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (!data.movies || data.movies.length === 0) {
      await expect(page.getByText('No media files found')).toBeVisible()
      await expect(page.getByText('Your library will appear here once you download some media')).toBeVisible()
    }
  })

  test('Clean Filenames button should trigger reprocessing', async ({ page }) => {
    const cleanButton = page.getByRole('button', { name: /Clean Filenames/i })
    
    // Click the button
    await cleanButton.click()
    
    // Should show loading state
    await expect(cleanButton).toBeDisabled()
    
    // Wait for API call
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/library/reprocess') && response.status() === 200
    )
    
    // Should show toast notification
    await expect(page.getByText(/Library reprocessed|reprocessed/i)).toBeVisible()
  })

  test('should handle images with loading="lazy" attribute', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/library/movies') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.movies && data.movies.length > 0) {
      // Check for lazy loading attribute
      const lazyImage = page.locator('img[loading="lazy"]').first()
      await expect(lazyImage).toBeVisible()
    }
  })

  test('should display correct number of items in description', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/library/movies') && response.status() === 200
    )
    
    const data = await response.json()
    const itemCount = data.movies?.length || 0
    
    // Description should show correct count
    await expect(page.getByText(`${itemCount} items`)).toBeVisible()
  })
})
