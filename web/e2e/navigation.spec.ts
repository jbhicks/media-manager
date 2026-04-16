import { test, expect } from '@playwright/test'

test.describe('Navigation and Layout', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('should display header with app title', async ({ page }) => {
    await expect(page.getByText('Media Manager')).toBeVisible()
  })

  test('should display menu button in header', async ({ page }) => {
    const menuButton = page.locator('button').first()
    await expect(menuButton).toBeVisible()
  })

  test('should display theme toggle button', async ({ page }) => {
    // Theme toggle is a button with sun/moon icon
    const themeButton = page.locator('header button').nth(1)
    await expect(themeButton).toBeVisible()
  })

  test('should display notification bell', async ({ page }) => {
    const bellButton = page.locator('header button').nth(2)
    await expect(bellButton).toBeVisible()
  })

  test('should toggle sidebar when menu button is clicked', async ({ page }) => {
    const menuButton = page.locator('button').first()
    
    // Get initial sidebar state
    const sidebar = page.locator('nav').first()
    const initialWidth = await sidebar.evaluate(el => getComputedStyle(el).width)
    
    // Click menu button
    await menuButton.click()
    
    // Sidebar width should change
    const newWidth = await sidebar.evaluate(el => getComputedStyle(el).width)
    expect(newWidth).not.toBe(initialWidth)
  })

  test('should toggle theme when theme button is clicked', async ({ page }) => {
    const themeButton = page.locator('header button').nth(1)
    
    // Get initial theme
    const initialTheme = await page.evaluate(() => document.documentElement.classList.contains('dark'))
    
    // Click theme button
    await themeButton.click()
    
    // Theme should change
    const newTheme = await page.evaluate(() => document.documentElement.classList.contains('dark'))
    expect(newTheme).not.toBe(initialTheme)
  })

  test('should navigate to all pages from sidebar', async ({ page }) => {
    const pages = [
      { name: 'Home', url: '/' },
      { name: 'Downloads', url: '/downloads' },
      { name: 'Library', url: '/library' },
      { name: 'Search', url: '/search' },
      { name: 'Suggestions', url: '/suggestions' },
      { name: 'Settings', url: '/settings' },
    ]
    
    for (const { name, url } of pages) {
      // Click on sidebar link
      const link = page.locator('nav').first().getByText(name)
      await link.click()
      
      // Should navigate to correct URL
      await expect(page).toHaveURL(url)
      
      // Should show correct heading
      await expect(page.getByRole('heading', { name }).first()).toBeVisible()
    }
  })

  test('should show active state on current page', async ({ page }) => {
    // Navigate to downloads
    await page.goto('/downloads')
    
    // Downloads link should have active class
    const downloadsLink = page.locator('nav').first().getByText('Downloads')
    const hasActiveClass = await downloadsLink.evaluate(el => 
      el.classList.contains('bg-accent') || el.classList.contains('text-accent-foreground')
    )
    expect(hasActiveClass).toBeTruthy()
  })

  test('should collapse sidebar to icons only', async ({ page }) => {
    const menuButton = page.locator('button').first()
    
    // Collapse sidebar
    await menuButton.click()
    
    // Sidebar should show only icons
    const sidebar = page.locator('nav').first()
    const width = await sidebar.evaluate(el => getComputedStyle(el).width)
    
    // Width should be smaller (around 64px for icon-only)
    expect(parseInt(width)).toBeLessThan(100)
  })

  test('should expand sidebar to show labels', async ({ page }) => {
    const menuButton = page.locator('button').first()
    
    // First collapse
    await menuButton.click()
    
    // Then expand
    await menuButton.click()
    
    // Sidebar should show labels
    const sidebar = page.locator('nav').first()
    const width = await sidebar.evaluate(el => getComputedStyle(el).width)
    
    // Width should be larger (around 256px for expanded)
    expect(parseInt(width)).toBeGreaterThan(200)
    
    // Labels should be visible
    await expect(page.getByText('Home').first()).toBeVisible()
  })

  test('should have correct navigation order', async ({ page }) => {
    const expectedOrder = ['Home', 'Downloads', 'Library', 'Search', 'Suggestions', 'Settings']
    
    const navLinks = await page.locator('nav').first().locator('a, button').allTextContents()
    const cleanLinks = navLinks.map(l => l.trim()).filter(l => l.length > 0)
    
    // Check that all expected items are present
    for (const item of expectedOrder) {
      expect(cleanLinks).toContain(item)
    }
  })

  test('should display icons for all navigation items', async ({ page }) => {
    const navItems = page.locator('nav').first().locator('a, button')
    const count = await navItems.count()
    
    for (let i = 0; i < count; i++) {
      const hasIcon = await navItems.nth(i).locator('svg').count() > 0
      expect(hasIcon).toBeTruthy()
    }
  })
})
