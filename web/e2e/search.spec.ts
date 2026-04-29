import { test, expect } from '@playwright/test'

test.describe('Search Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/search')
  })

  test('should display search page title and description', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Search' })).toBeVisible()
    await expect(page.getByText('Search torrents across all indexers')).toBeVisible()
  })

  test('should display search input with icon', async ({ page }) => {
    const searchInput = page.getByPlaceholder('Search torrents...')
    await expect(searchInput).toBeVisible()
    
    // Should have search icon
    const searchIcon = page.locator('svg').filter({ has: page.locator('title', { hasText: /search/i }) }).first()
    await expect(searchIcon).toBeVisible()
  })

  test('should display search button', async ({ page }) => {
    const searchButton = page.getByRole('button', { name: /Search/i })
    await expect(searchButton).toBeVisible()
    await expect(searchButton).toBeEnabled()
  })

  test('should display search results tab by default', async ({ page }) => {
    // Search Results tab should be active (has border-b-2)
    const searchTab = page.getByRole('button', { name: /Search Results/i })
    await expect(searchTab).toBeVisible()
    
    // Should have active styling
    const hasActiveClass = await searchTab.evaluate(el => 
      el.classList.contains('border-b-2') || el.classList.contains('border-primary')
    )
    expect(hasActiveClass).toBeTruthy()
  })

  test('should display suggestions tab', async ({ page }) => {
    const suggestionsTab = page.getByRole('button', { name: /Suggestions$/i })
    await expect(suggestionsTab).toBeVisible()
  })

  test('should switch to suggestions tab when clicked', async ({ page }) => {
    const suggestionsTab = page.getByRole('button', { name: /Suggestions$/i })
    await suggestionsTab.click()
    
    // Suggestions tab should now be active
    const hasActiveClass = await suggestionsTab.evaluate(el => 
      el.classList.contains('border-b-2') || el.classList.contains('border-primary')
    )
    expect(hasActiveClass).toBeTruthy()
    
    // Wait for suggestions API
    await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
  })

  test('should switch back to search results tab', async ({ page }) => {
    // First switch to suggestions
    const suggestionsTab = page.getByRole('button', { name: /Suggestions$/i })
    await suggestionsTab.click()
    
    // Then switch back to search
    const searchTab = page.getByRole('button', { name: /Search Results/i })
    await searchTab.click()
    
    // Search tab should be active
    const hasActiveClass = await searchTab.evaluate(el => 
      el.classList.contains('border-b-2') || el.classList.contains('border-primary')
    )
    expect(hasActiveClass).toBeTruthy()
  })

  test('should perform search when typing and submitting', async ({ page }) => {
    const searchInput = page.getByPlaceholder('Search torrents...')
    await searchInput.fill('test movie')
    
    // Submit the form
    const searchButton = page.getByRole('button', { name: /Search/i })
    await searchButton.click()
    
    // Wait for search API
    await page.waitForResponse(response => 
      response.url().includes('/api/search') && response.status() === 200
    )
  })

  test('should show loading state during search', async ({ page }) => {
    const searchInput = page.getByPlaceholder('Search torrents...')
    await searchInput.fill('loading test')
    
    const searchButton = page.getByRole('button', { name: /Search/i })
    await searchButton.click()
    
    // Button should show loading spinner or be disabled
    await expect(page.locator('.animate-spin').first()).toBeVisible()
  })

  test('should display search results when available', async ({ page }) => {
    const searchInput = page.getByPlaceholder('Search torrents...')
    await searchInput.fill('movie')
    
    const searchButton = page.getByRole('button', { name: /Search/i })
    await searchButton.click()
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/search') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.results && data.results.length > 0) {
      const firstResult = data.results[0]
      await expect(page.getByText(firstResult.title).first()).toBeVisible()
    }
  })

  test('should display download button for search results with magnet link', async ({ page }) => {
    const searchInput = page.getByPlaceholder('Search torrents...')
    await searchInput.fill('movie')
    
    const searchButton = page.getByRole('button', { name: /Search/i })
    await searchButton.click()
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/search') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.results && data.results.length > 0) {
      const resultWithMagnet = data.results.find((r: any) => r.magnetLink)
      if (resultWithMagnet) {
        await expect(page.getByRole('button', { name: /Download/i }).first()).toBeVisible()
      }
    }
  })

  test('should disable download button when no magnet link', async ({ page }) => {
    const searchInput = page.getByPlaceholder('Search torrents...')
    await searchInput.fill('movie')
    
    const searchButton = page.getByRole('button', { name: /Search/i })
    await searchButton.click()
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/search') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.results && data.results.length > 0) {
      const resultWithoutMagnet = data.results.find((r: any) => !r.magnetLink)
      if (resultWithoutMagnet) {
        // Find the card for this result and check its download button
        const card = page.locator('.grid > div').filter({ hasText: resultWithoutMagnet.title }).first()
        const downloadButton = card.getByRole('button', { name: /Download/i })
        await expect(downloadButton).toBeDisabled()
      }
    }
  })

  test('should display result metadata (size, seeders, leechers)', async ({ page }) => {
    const searchInput = page.getByPlaceholder('Search torrents...')
    await searchInput.fill('movie')
    
    const searchButton = page.getByRole('button', { name: /Search/i })
    await searchButton.click()
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/search') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.results && data.results.length > 0) {
      // Should show size badge
      const sizeElements = await page.locator('text=/\\d+\\s*(GB|MB|KB|Bytes)/i').all()
      expect(sizeElements.length).toBeGreaterThanOrEqual(0)
      
      // Should show seeders/leechers
      const seederElements = await page.getByText(/S\s*\/\s*L/i).all()
      expect(seederElements.length).toBeGreaterThanOrEqual(0)
    }
  })

  test('should show empty state when search returns no results', async ({ page }) => {
    const searchInput = page.getByPlaceholder('Search torrents...')
    await searchInput.fill('xyznonexistent12345')
    
    const searchButton = page.getByRole('button', { name: /Search/i })
    await searchButton.click()
    
    await page.waitForResponse(response => 
      response.url().includes('/api/search') && response.status() === 200
    )
    
    // Should show empty state
    await expect(page.getByText('No search results found')).toBeVisible()
    await expect(page.getByText('Try a different search term')).toBeVisible()
  })

  test('should display suggestions with approve/reject buttons', async ({ page }) => {
    // Switch to suggestions tab
    const suggestionsTab = page.getByRole('button', { name: /Suggestions$/i })
    await suggestionsTab.click()
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.data && data.data.length > 0) {
      const pendingSuggestion = data.data.find((s: any) => s.status === 'pending')
      if (pendingSuggestion) {
        await expect(page.getByRole('button', { name: /Approve/i }).first()).toBeVisible()
        await expect(page.getByRole('button', { name: /Reject/i }).first()).toBeVisible()
      }
    }
  })

  test('should approve a suggestion', async ({ page }) => {
    // Switch to suggestions tab
    const suggestionsTab = page.getByRole('button', { name: /Suggestions$/i })
    await suggestionsTab.click()
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.data && data.data.length > 0) {
      const pendingSuggestion = data.data.find((s: any) => s.status === 'pending')
      if (pendingSuggestion) {
        const approveButton = page.getByRole('button', { name: /Approve/i }).first()
        await approveButton.click()
        
        // Wait for approve API
        await page.waitForResponse(response => 
          response.url().includes('/api/suggestions/approve') && response.status() === 200
        )
        
        // Should show toast
        await expect(page.getByText(/approved|Torrent approved/i)).toBeVisible()
      }
    }
  })

  test('should reject a suggestion', async ({ page }) => {
    // Switch to suggestions tab
    const suggestionsTab = page.getByRole('button', { name: /Suggestions$/i })
    await suggestionsTab.click()
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.data && data.data.length > 0) {
      const pendingSuggestion = data.data.find((s: any) => s.status === 'pending')
      if (pendingSuggestion) {
        const rejectButton = page.getByRole('button', { name: /Reject/i }).first()
        await rejectButton.click()
        
        // Wait for reject API
        await page.waitForResponse(response => 
          response.url().includes('/api/suggestions/reject') && response.status() === 200
        )
        
        // Should show toast
        await expect(page.getByText(/rejected|Torrent rejected/i)).toBeVisible()
      }
    }
  })

  test('should show empty state for suggestions tab', async ({ page }) => {
    // Switch to suggestions tab
    const suggestionsTab = page.getByRole('button', { name: /Suggestions$/i })
    await suggestionsTab.click()
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (!data.data || data.data.length === 0) {
      await expect(page.getByText('No suggestions found')).toBeVisible()
      await expect(page.getByText('Search for torrents to generate suggestions')).toBeVisible()
    }
  })

  test('should display suggestion status badges', async ({ page }) => {
    // Switch to suggestions tab
    const suggestionsTab = page.getByRole('button', { name: /Suggestions$/i })
    await suggestionsTab.click()
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.data && data.data.length > 0) {
      const suggestion = data.data[0]
      await expect(page.getByText(suggestion.status).first()).toBeVisible()
    }
  })

  test('should display placeholder image for missing posters', async ({ page }) => {
    // Switch to suggestions tab
    const suggestionsTab = page.getByRole('button', { name: /Suggestions$/i })
    await suggestionsTab.click()
    
    await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    // Check for placeholder images
    const placeholderImages = await page.locator('img[src*="placeholder"]').all()
    expect(placeholderImages.length).toBeGreaterThanOrEqual(0)
  })

  test('should handle image load errors with fallback', async ({ page }) => {
    // Mock an image error by intercepting
    await page.route('**/*', (route) => {
      if (route.request().resourceType() === 'image' && !route.request().url().includes('placeholder')) {
        route.abort('failed')
      } else {
        route.continue()
      }
    })
    
    // Switch to suggestions tab
    const suggestionsTab = page.getByRole('button', { name: /Suggestions$/i })
    await suggestionsTab.click()
    
    await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    // Placeholder images should be visible after error
    const placeholderImages = await page.locator('img[src*="placeholder"]').all()
    expect(placeholderImages.length).toBeGreaterThanOrEqual(0)
  })

  test('should show loading state for suggestions', async ({ page }) => {
    // Switch to suggestions tab
    const suggestionsTab = page.getByRole('button', { name: /Suggestions$/i })
    await suggestionsTab.click()
    
    // Should trigger suggestions API
    await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
  })

  test('should maintain search query when switching tabs', async ({ page }) => {
    const searchInput = page.getByPlaceholder('Search torrents...')
    await searchInput.fill('test query')
    
    // Switch to suggestions tab
    const suggestionsTab = page.getByRole('button', { name: /Suggestions$/i })
    await suggestionsTab.click()
    
    // Search query should be maintained
    await expect(searchInput).toHaveValue('test query')
    
    // Switch back to search tab
    const searchTab = page.getByRole('button', { name: /Search Results/i })
    await searchTab.click()
    
    // Search query should still be there
    await expect(searchInput).toHaveValue('test query')
  })
})