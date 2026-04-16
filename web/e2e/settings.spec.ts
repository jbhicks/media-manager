import { test, expect } from '@playwright/test'

test.describe('Settings Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/settings')
  })

  test('should display settings page title and description', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible()
    await expect(page.getByText('Configure your media manager')).toBeVisible()
  })

  test('should display Download Sources card', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Download Sources' })).toBeVisible()
    await expect(page.getByText(/sources configured/i)).toBeVisible()
    await expect(page.getByRole('button', { name: 'Manage Sources' })).toBeVisible()
  })

  test('should display Download Rules card', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Download Rules' })).toBeVisible()
    await expect(page.getByText(/rules configured/i)).toBeVisible()
    await expect(page.getByRole('button', { name: 'Manage Rules' })).toBeVisible()
  })

  test('should display Server Settings card', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Server Settings' })).toBeVisible()
    await expect(page.getByText(/Configure server and connection settings/i)).toBeVisible()
    await expect(page.getByRole('button', { name: 'Configure Server' })).toBeVisible()
  })

  test('should display Security card', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Security' })).toBeVisible()
    await expect(page.getByText(/Manage API keys and authentication/i)).toBeVisible()
    await expect(page.getByRole('button', { name: 'Security Settings' })).toBeVisible()
  })

  test('should show correct source count from API', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/sources') && response.status() === 200
    )
    
    const data = await response.json()
    const sourceCount = data.sources?.length || 0
    
    await expect(page.getByText(`${sourceCount} sources configured`)).toBeVisible()
  })

  test('should show correct rules count from API', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/rules') && response.status() === 200
    )
    
    const data = await response.json()
    const ruleCount = data.rules?.length || 0
    
    await expect(page.getByText(`${ruleCount} rules configured`)).toBeVisible()
  })

  test('settings cards should have icons', async ({ page }) => {
    // Check for icons in each card (icons are SVGs)
    const cards = ['Download Sources', 'Download Rules', 'Server Settings', 'Security']
    
    for (const cardName of cards) {
      const card = page.locator('h3', { hasText: cardName }).locator('..').locator('..')
      const hasIcon = await card.locator('svg').count() > 0
      expect(hasIcon).toBeTruthy()
    }
  })

  test('Manage Sources button should be clickable', async ({ page }) => {
    const button = page.getByRole('button', { name: 'Manage Sources' })
    await expect(button).toBeEnabled()
    
    // Click (functionality not yet implemented, but button should respond)
    await button.click()
    
    // Since it's not implemented, page shouldn't navigate away
    await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible()
  })

  test('Manage Rules button should be clickable', async ({ page }) => {
    const button = page.getByRole('button', { name: 'Manage Rules' })
    await expect(button).toBeEnabled()
    
    await button.click()
    
    await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible()
  })

  test('Configure Server button should be clickable', async ({ page }) => {
    const button = page.getByRole('button', { name: 'Configure Server' })
    await expect(button).toBeEnabled()
    
    await button.click()
    
    await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible()
  })

  test('Security Settings button should be clickable', async ({ page }) => {
    const button = page.getByRole('button', { name: 'Security Settings' })
    await expect(button).toBeEnabled()
    
    await button.click()
    
    await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible()
  })

  test('settings page should load data from multiple APIs', async ({ page }) => {
    // Wait for both sources and rules APIs
    const [sourcesResponse, rulesResponse] = await Promise.all([
      page.waitForResponse(response => response.url().includes('/api/sources')),
      page.waitForResponse(response => response.url().includes('/api/rules')),
    ])
    
    expect(sourcesResponse.status()).toBe(200)
    expect(rulesResponse.status()).toBe(200)
  })

  test('should display cards in 2-column grid on desktop', async ({ page }) => {
    // Set desktop viewport
    await page.setViewportSize({ width: 1280, height: 720 })
    
    // Grid should be visible
    const grid = page.locator('.grid.md\\:grid-cols-2')
    await expect(grid).toBeVisible()
  })
})
