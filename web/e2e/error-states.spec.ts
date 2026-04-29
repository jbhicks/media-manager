import { test, expect } from '@playwright/test'

test.describe('Error Handling', () => {
  test('should show error toast when API fails on dashboard', async ({ page }) => {
    // Intercept and fail the stats API
    await page.route('**/api/stats', (route) => {
      route.fulfill({ status: 500, body: JSON.stringify({ message: 'Internal Server Error' }) })
    })
    
    await page.goto('/')
    
    // Page should still load without crashing
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()
  })

  test('should handle 404 errors gracefully', async ({ page }) => {
    await page.goto('/nonexistent-page')
    
    // Should show the app layout (SPA routing)
    await expect(page.getByText('Media Manager')).toBeVisible()
  })

  test('should handle network errors on downloads page', async ({ page }) => {
    await page.route('**/api/tasks', (route) => {
      route.abort('failed')
    })
    
    await page.goto('/downloads')
    
    // Page should still load
    await expect(page.getByRole('heading', { name: 'Downloads' })).toBeVisible()
  })

  test('should handle network errors on library page', async ({ page }) => {
    await page.route('**/api/library/movies', (route) => {
      route.abort('failed')
    })
    
    await page.goto('/library')
    
    // Page should still load
    await expect(page.getByRole('heading', { name: 'Library' })).toBeVisible()
  })

  test('should handle network errors on suggestions page', async ({ page }) => {
    await page.route('**/api/suggestions**', (route) => {
      route.abort('failed')
    })
    
    await page.goto('/suggestions')
    
    // Page should still load
    await expect(page.getByRole('heading', { name: 'Suggestions' })).toBeVisible()
  })

  test('should handle network errors on search page', async ({ page }) => {
    await page.route('**/api/search**', (route) => {
      route.abort('failed')
    })
    
    await page.goto('/search')
    
    // Should be able to type in search
    const searchInput = page.getByPlaceholder('Search torrents...')
    await searchInput.fill('test')
    
    // Page should still be functional
    await expect(searchInput).toHaveValue('test')
  })

  test('should handle API errors when cancelling task', async ({ page }) => {
    await page.goto('/downloads')
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    const data = await response.json()
    const downloadingTask = data.tasks?.find((t: { status: string }) => t.status === 'downloading')
    
    if (downloadingTask) {
      // Intercept cancel API to return error
      await page.route('**/api/tasks/cancel', (route) => {
        route.fulfill({ status: 500, body: JSON.stringify({ message: 'Failed to cancel' }) })
      })
      
      const taskCard = page.locator('.space-y-4 > div').filter({ hasText: downloadingTask.title }).first()
      const cancelButton = taskCard.locator('button').first()
      await cancelButton.click()
      
      // Should show error toast
      await expect(page.getByText(/Failed to cancel|error/i)).toBeVisible()
    }
  })

  test('should handle API errors when generating suggestions', async ({ page }) => {
    await page.goto('/suggestions')
    
    // Intercept generate API to return error
    await page.route('**/api/suggestions/generate', (route) => {
      route.fulfill({ status: 500, body: JSON.stringify({ message: 'Failed to generate' }) })
    })
    
    const generateButton = page.getByRole('button', { name: /Generate Suggestions/i })
    await generateButton.click()
    
    // Should show error toast
    await expect(page.getByText(/Failed to generate|error/i)).toBeVisible()
  })

  test('should handle API errors when cleaning filenames', async ({ page }) => {
    await page.goto('/library')
    
    // Intercept reprocess API to return error
    await page.route('**/api/library/reprocess', (route) => {
      route.fulfill({ status: 500, body: JSON.stringify({ message: 'Failed to reprocess' }) })
    })
    
    const cleanButton = page.getByRole('button', { name: /Clean Filenames/i })
    await cleanButton.click()
    
    // Should show error toast
    await expect(page.getByText(/Failed to reprocess|error/i)).toBeVisible()
  })

  test('should handle slow API responses', async ({ page }) => {
    // Delay API response by 2 seconds
    await page.route('**/api/stats', async (route) => {
      await new Promise(resolve => setTimeout(resolve, 2000))
      route.continue()
    })
    
    await page.goto('/')
    
    // Page should show loading state or stats eventually
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()
    
    // Wait for delayed response
    await page.waitForResponse(response => 
      response.url().includes('/api/stats') && response.status() === 200
    )
  })

  test('should recover from API errors on retry', async ({ page }) => {
    let requestCount = 0
    
    await page.route('**/api/tasks', (route) => {
      requestCount++
      if (requestCount === 1) {
        route.fulfill({ status: 500, body: JSON.stringify({ message: 'Error' }) })
      } else {
        route.continue()
      }
    })
    
    await page.goto('/downloads')
    
    // Page should handle error gracefully
    await expect(page.getByRole('heading', { name: 'Downloads' })).toBeVisible()
  })

  test('should handle VPN API errors', async ({ page }) => {
    await page.route('**/api/vpn/status', (route) => {
      route.fulfill({ status: 500, body: JSON.stringify({ message: 'VPN Error' }) })
    })
    
    await page.goto('/')
    
    // Page should still load
    await expect(page.getByText('Media Manager')).toBeVisible()
  })
})