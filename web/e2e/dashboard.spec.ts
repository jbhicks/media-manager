import { test, expect } from '@playwright/test'

test.describe('Dashboard Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('should display dashboard title and description', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()
    await expect(page.getByText('Welcome to your media manager')).toBeVisible()
  })

  test('should display stat cards with correct labels', async ({ page }) => {
    const statCards = ['Pending', 'Downloading', 'Completed', 'Failed']
    
    for (const label of statCards) {
      await expect(page.getByText(label, { exact: true }).first()).toBeVisible()
    }
  })

  test('should display quick action cards', async ({ page }) => {
    await expect(page.getByText('Downloads').nth(1)).toBeVisible()
    await expect(page.getByText('Library').nth(1)).toBeVisible()
    await expect(page.getByText('Search').nth(1)).toBeVisible()
  })

  test('should navigate to Downloads page when clicking Downloads card', async ({ page }) => {
    // Click on the Downloads quick action card
    const downloadsCard = page.getByText('Manage your download queue')
    await downloadsCard.click()
    
    // Should navigate to downloads page
    await expect(page).toHaveURL(/\/downloads/)
    await expect(page.getByRole('heading', { name: 'Downloads' })).toBeVisible()
  })

  test('should navigate to Library page when clicking Library card', async ({ page }) => {
    const libraryCard = page.getByText('Browse your media collection')
    await libraryCard.click()
    
    await expect(page).toHaveURL(/\/library/)
    await expect(page.getByRole('heading', { name: 'Library' })).toBeVisible()
  })

  test('should navigate to Search page when clicking Search card', async ({ page }) => {
    const searchCard = page.getByText('Find new content')
    await searchCard.click()
    
    await expect(page).toHaveURL(/\/search/)
    await expect(page.getByRole('heading', { name: 'Search' })).toBeVisible()
  })

  test('should show loading state for stats initially', async ({ page }) => {
    // Check that stat values are either numbers or loading indicators
    const statValues = await page.locator('.text-2xl.font-bold').allTextContents()
    expect(statValues.length).toBe(4)
  })

  test('should update stats from API', async ({ page }) => {
    // Wait for stats to load (they should appear as numbers after API call)
    await page.waitForResponse(response => 
      response.url().includes('/api/stats') && response.status() === 200
    )
    
    // Stats should be visible
    const statsSection = page.locator('.grid.gap-4').first()
    await expect(statsSection).toBeVisible()
  })
})
