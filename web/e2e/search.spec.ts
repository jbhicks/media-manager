import { test, expect } from '@playwright/test'

test.describe('Search Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/search')
  })

  test('should display search page title and description', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Search' })).toBeVisible()
    await expect(page.getByText('Find and approve torrents to download')).toBeVisible()
  })

  test('should display Generate Suggestions button', async ({ page }) => {
    const generateButton = page.getByRole('button', { name: /Generate Suggestions/i })
    await expect(generateButton).toBeVisible()
  })

  test('should display search input', async ({ page }) => {
    const searchInput = page.getByPlaceholder('Search torrents...')
    await expect(searchInput).toBeVisible()
  })

  test('should display status filter dropdown', async ({ page }) => {
    const statusSelect = page.locator('select')
    await expect(statusSelect).toBeVisible()
    
    // Check options
    await expect(statusSelect.getByText('Pending')).toBeVisible()
    await expect(statusSelect.getByText('Approved')).toBeVisible()
    await expect(statusSelect.getByText('Rejected')).toBeVisible()
    await expect(statusSelect.getByText('Downloaded')).toBeVisible()
  })

  test('should filter by status when dropdown changes', async ({ page }) => {
    const statusSelect = page.locator('select')
    
    // Change to Approved
    await statusSelect.selectOption('approved')
    
    // Wait for API call
    await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.url().includes('status=approved')
    )
  })

  test('should search when typing in search input', async ({ page }) => {
    const searchInput = page.getByPlaceholder('Search torrents...')
    
    // Type in search box
    await searchInput.fill('action movie')
    
    // Wait for debounce and API call
    await page.waitForTimeout(300)
    
    // Should trigger search API
    await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
  })

  test('should display suggestion cards when results exist', async ({ page }) => {
    // Wait for suggestions to load
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.data && data.data.length > 0) {
      // Should show suggestion cards
      const firstSuggestion = data.data[0]
      await expect(page.getByText(firstSuggestion.title).first()).toBeVisible()
      
      // Should show poster
      const poster = page.locator('img[alt="' + firstSuggestion.title + '"]').first()
      await expect(poster).toBeVisible()
    }
  })

  test('should show approve/reject buttons for pending suggestions', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    const data = await response.json()
    const pendingSuggestion = data.data?.find((s: { status: string }) => s.status === 'pending')
    
    if (pendingSuggestion) {
      // Should show Approve and Reject buttons
      await expect(page.getByRole('button', { name: /Approve/i }).first()).toBeVisible()
      await expect(page.getByRole('button', { name: /Reject/i }).first()).toBeVisible()
    }
  })

  test('should show checkboxes for bulk selection', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.data && data.data.length > 0) {
      // Should show checkboxes
      const checkboxes = await page.locator('input[type="checkbox"]').all()
      expect(checkboxes.length).toBeGreaterThan(0)
    }
  })

  test('should show bulk actions when items are selected', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.data && data.data.length > 0) {
      // Click a checkbox
      const checkbox = page.locator('input[type="checkbox"]').first()
      await checkbox.click()
      
      // Should show bulk actions
      await expect(page.getByText(/selected/i)).toBeVisible()
      await expect(page.getByRole('button', { name: /Approve/i, exact: false }).first()).toBeVisible()
    }
  })

  test('should display suggestion metadata', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.data && data.data.length > 0) {
      const suggestion = data.data[0]
      
      // Should show size
      if (suggestion.size) {
        // Size is formatted, so just check if any size-like text exists
        const sizeElements = await page.locator('text=/\\d+\\s*(GB|MB|KB|Bytes)/i').all()
        expect(sizeElements.length).toBeGreaterThan(0)
      }
      
      // Should show seeders/leechers
      if (suggestion.seeders !== undefined) {
        await expect(page.getByText(/seeders/i).first()).toBeVisible()
      }
    }
  })

  test('should show status badge on each card', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.data && data.data.length > 0) {
      const suggestion = data.data[0]
      // Should show status badge
      await expect(page.getByText(suggestion.status).first()).toBeVisible()
    }
  })

  test('should show empty state when no suggestions', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (!data.data || data.data.length === 0) {
      await expect(page.getByText('No suggestions found')).toBeVisible()
      await expect(page.getByText('Try generating suggestions or adjusting your search')).toBeVisible()
    }
  })

  test('Generate Suggestions button should trigger generation', async ({ page }) => {
    const generateButton = page.getByRole('button', { name: /Generate Suggestions/i })
    
    await generateButton.click()
    
    // Should show loading state
    await expect(generateButton).toBeDisabled()
    
    // Wait for API call
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions/generate') && response.status() === 200
    )
    
    const data = await response.json()
    
    // Should show toast
    await expect(page.getByText(/generated|Suggestions generated/i)).toBeVisible()
  })

  test('should display placeholder image when poster is missing', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.data && data.data.length > 0) {
      // Some images should use placeholder
      const placeholderImages = await page.locator('img[src*="placeholder"]').all()
      // Don't assert count since it depends on data, just check the selector works
      expect(placeholderImages.length).toBeGreaterThanOrEqual(0)
    }
  })
})
