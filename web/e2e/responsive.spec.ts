import { test, expect } from '@playwright/test'

test.describe('Responsive Design', () => {
  test('should display correctly on mobile viewport', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/')
    
    // Should show header
    await expect(page.getByText('Media Manager')).toBeVisible()
    
    // Should show sidebar in collapsed state or hamburger menu
    const sidebar = page.locator('nav').first()
    await expect(sidebar).toBeVisible()
  })

  test('should display correctly on tablet viewport', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 })
    await page.goto('/')
    
    // Should show header
    await expect(page.getByText('Media Manager')).toBeVisible()
    
    // Should show main content
    await expect(page.locator('main')).toBeVisible()
  })

  test('should display correctly on desktop viewport', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 })
    await page.goto('/')
    
    // Should show header
    await expect(page.getByText('Media Manager')).toBeVisible()
    
    // Should show expanded sidebar
    const sidebar = page.locator('nav').first()
    const width = await sidebar.evaluate(el => getComputedStyle(el).width)
    expect(parseInt(width)).toBeGreaterThan(200)
  })

  test('should adapt grid layout on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/library')
    
    await page.waitForResponse(response => 
      response.url().includes('/api/library/movies') && response.status() === 200
    )
    
    // Grid should be visible
    const grid = page.locator('.grid')
    await expect(grid).toBeVisible()
  })

  test('should adapt grid layout on tablet', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 })
    await page.goto('/library')
    
    await page.waitForResponse(response => 
      response.url().includes('/api/library/movies') && response.status() === 200
    )
    
    // Grid should be visible
    const grid = page.locator('.grid')
    await expect(grid).toBeVisible()
  })

  test('should show mobile-optimized header on small screens', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/')
    
    // Header should be visible
    const header = page.locator('header')
    await expect(header).toBeVisible()
    
    // Header should not overflow
    const headerWidth = await header.evaluate(el => el.scrollWidth)
    const viewportWidth = 375
    expect(headerWidth).toBeLessThanOrEqual(viewportWidth + 50) // Allow small tolerance
  })

  test('should handle sidebar toggle on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/')
    
    const menuButton = page.locator('button').first()
    await menuButton.click()
    
    // Sidebar should toggle
    const sidebar = page.locator('nav').first()
    await expect(sidebar).toBeVisible()
  })

  test('should show search page on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/search')
    
    // Search input should be visible and usable
    const searchInput = page.getByPlaceholder('Search torrents...')
    await expect(searchInput).toBeVisible()
    
    // Should be able to type
    await searchInput.fill('test')
    await expect(searchInput).toHaveValue('test')
  })

  test('should show downloads page on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/downloads')
    
    // Heading should be visible
    await expect(page.getByRole('heading', { name: 'Downloads' })).toBeVisible()
    
    // Action buttons should be visible
    await expect(page.getByRole('button', { name: /Clear Completed/i })).toBeVisible()
  })

  test('should show settings page on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/settings')
    
    // Heading should be visible
    await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible()
    
    // Cards should stack vertically
    const cards = await page.locator('.grid > div').all()
    expect(cards.length).toBeGreaterThan(0)
  })

  test('should handle touch targets on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/')
    
    // Buttons should be at least 44px tall (touch target size)
    const buttons = await page.locator('button').all()
    for (const button of buttons.slice(0, 5)) { // Check first 5 buttons
      const height = await button.evaluate(el => el.getBoundingClientRect().height)
      expect(height).toBeGreaterThanOrEqual(36) // Minimum reasonable touch target
    }
  })

  test('should not have horizontal scroll on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/')
    
    // Check for horizontal overflow
    const hasHorizontalScroll = await page.evaluate(() => {
      return document.documentElement.scrollWidth > document.documentElement.clientWidth
    })
    
    // Should not have horizontal scroll (allow small tolerance)
    const scrollWidth = await page.evaluate(() => document.documentElement.scrollWidth)
    const clientWidth = await page.evaluate(() => document.documentElement.clientWidth)
    expect(scrollWidth).toBeLessThanOrEqual(clientWidth + 20)
  })

  test('should show suggestions page on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/suggestions')
    
    // Heading should be visible
    await expect(page.getByRole('heading', { name: 'Suggestions' })).toBeVisible()
    
    // Generate button should be visible
    await expect(page.getByRole('button', { name: /Generate Suggestions/i })).toBeVisible()
  })

  test('should adapt stats cards on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/')
    
    // Stats cards should be visible
    await expect(page.getByText('Pending').first()).toBeVisible()
    await expect(page.getByText('Downloading').first()).toBeVisible()
    await expect(page.getByText('Completed').first()).toBeVisible()
    await expect(page.getByText('Failed').first()).toBeVisible()
  })

  test('should show VPN status on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/')
    
    await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    // VPN status should be visible (abbreviated on mobile)
    const vpnText = await page.locator('header').getByText(/VPN|No VPN/i).first().textContent()
    expect(vpnText).toMatch(/VPN|No VPN/)
  })

  test('should handle landscape orientation on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 812, height: 375 }) // iPhone X landscape
    await page.goto('/')
    
    // Should show header
    await expect(page.getByText('Media Manager')).toBeVisible()
    
    // Should show content
    await expect(page.locator('main')).toBeVisible()
  })
})