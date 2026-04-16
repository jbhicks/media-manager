import { test, expect } from '@playwright/test'

test.describe('Suggestions Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/suggestions')
  })

  test('should display suggestions page title and description', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Suggestions' })).toBeVisible()
    await expect(page.getByText('AI-powered download suggestions')).toBeVisible()
  })

  test('should display Generate Suggestions button', async ({ page }) => {
    const generateButton = page.getByRole('button', { name: /Generate Suggestions/i })
    await expect(generateButton).toBeVisible()
  })

  test('should display stats cards', async ({ page }) => {
    const statLabels = ['Pending', 'Approved', 'Rejected', 'Total']
    
    for (const label of statLabels) {
      await expect(page.getByText(label, { exact: true }).first()).toBeVisible()
    }
  })

  test('should show stats values from API', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions/stats') && response.status() === 200
    )
    
    const data = await response.json()
    
    // Stats should be displayed
    await expect(page.locator('.text-2xl.font-bold').first()).toBeVisible()
  })

  test('should display Recent Suggestions section', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Recent Suggestions' })).toBeVisible()
  })

  test('should display suggestion items when available', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.data && data.data.length > 0) {
      const firstSuggestion = data.data[0]
      // Should show suggestion title
      await expect(page.getByText(firstSuggestion.title).first()).toBeVisible()
    }
  })

  test('should display status badges on suggestions', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.data && data.data.length > 0) {
      const suggestion = data.data[0]
      // Should show status
      await expect(page.getByText(suggestion.status).first()).toBeVisible()
    }
  })

  test('should display seeder count on suggestions', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.data && data.data.length > 0) {
      const suggestion = data.data[0]
      if (suggestion.seeders !== undefined) {
        await expect(page.getByText(/seeders/i).first()).toBeVisible()
      }
    }
  })

  test('should show empty state when no pending suggestions', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (!data.data || data.data.length === 0) {
      await expect(page.getByText('No pending suggestions')).toBeVisible()
    }
  })

  test('Generate Suggestions button should trigger API call', async ({ page }) => {
    const generateButton = page.getByRole('button', { name: /Generate Suggestions/i })
    
    await generateButton.click()
    
    // Should show loading spinner
    await expect(generateButton).toBeDisabled()
    
    // Wait for API call
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/suggestions/generate') && response.status() === 200
    )
    
    // Should show success toast
    await expect(page.getByText(/successfully|generated/i)).toBeVisible()
  })

  test('should refresh stats after generating suggestions', async ({ page }) => {
    const generateButton = page.getByRole('button', { name: /Generate Suggestions/i })
    
    await generateButton.click()
    
    // Wait for generate API
    await page.waitForResponse(response => 
      response.url().includes('/api/suggestions/generate') && response.status() === 200
    )
    
    // Should refresh stats
    await page.waitForResponse(response => 
      response.url().includes('/api/suggestions/stats') && response.status() === 200
    )
  })

  test('should refresh suggestions list after generating', async ({ page }) => {
    const generateButton = page.getByRole('button', { name: /Generate Suggestions/i })
    
    await generateButton.click()
    
    // Wait for generate API
    await page.waitForResponse(response => 
      response.url().includes('/api/suggestions/generate') && response.status() === 200
    )
    
    // Should refresh suggestions
    await page.waitForResponse(response => 
      response.url().includes('/api/suggestions') && !response.url().includes('stats')
    )
  })

  test('stats cards should show numeric values', async ({ page }) => {
    await page.waitForResponse(response => 
      response.url().includes('/api/suggestions/stats') && response.status() === 200
    )
    
    // All stat values should be visible
    const statValues = await page.locator('.text-2xl.font-bold').allTextContents()
    expect(statValues.length).toBe(4)
    
    // Each should be a number (or 0)
    for (const value of statValues) {
      expect(parseInt(value)).toBeGreaterThanOrEqual(0)
    }
  })
})
