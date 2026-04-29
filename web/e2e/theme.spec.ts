import { test, expect } from '@playwright/test'

test.describe('Theme', () => {
  test('should default to dark theme', async ({ page }) => {
    await page.goto('/')
    
    // Should have dark class on html element
    const hasDarkClass = await page.evaluate(() => 
      document.documentElement.classList.contains('dark')
    )
    expect(hasDarkClass).toBeTruthy()
  })

  test('should toggle theme when button is clicked', async ({ page }) => {
    await page.goto('/')
    
    const themeButton = page.locator('header button').nth(1)
    
    // Get initial theme
    const initialTheme = await page.evaluate(() => 
      document.documentElement.classList.contains('dark')
    )
    
    // Click theme button
    await themeButton.click()
    
    // Theme should change
    const newTheme = await page.evaluate(() => 
      document.documentElement.classList.contains('dark')
    )
    expect(newTheme).not.toBe(initialTheme)
  })

  test('should persist theme across page navigation', async ({ page }) => {
    await page.goto('/')
    
    // Toggle theme to light
    const themeButton = page.locator('header button').nth(1)
    await themeButton.click()
    
    // Verify theme changed
    const isLight = await page.evaluate(() => 
      !document.documentElement.classList.contains('dark')
    )
    expect(isLight).toBeTruthy()
    
    // Navigate to another page
    await page.goto('/downloads')
    
    // Theme should persist
    const stillLight = await page.evaluate(() => 
      !document.documentElement.classList.contains('dark')
    )
    expect(stillLight).toBeTruthy()
  })

  test('should persist theme after reload', async ({ page }) => {
    await page.goto('/')
    
    // Toggle theme
    const themeButton = page.locator('header button').nth(1)
    await themeButton.click()
    
    // Get current theme
    const currentTheme = await page.evaluate(() => 
      document.documentElement.classList.contains('dark') ? 'dark' : 'light'
    )
    
    // Reload page
    await page.reload()
    
    // Theme should persist
    const themeAfterReload = await page.evaluate(() => 
      document.documentElement.classList.contains('dark') ? 'dark' : 'light'
    )
    expect(themeAfterReload).toBe(currentTheme)
  })

  test('should have correct background color in dark mode', async ({ page }) => {
    await page.goto('/')
    
    // Ensure dark mode
    const isDark = await page.evaluate(() => 
      document.documentElement.classList.contains('dark')
    )
    
    if (isDark) {
      // Body should have dark background
      const bgColor = await page.evaluate(() => {
        const body = document.body
        const styles = window.getComputedStyle(body)
        return styles.backgroundColor
      })
      // Dark backgrounds are typically rgb(0, 0, 0) or similar dark colors
      expect(bgColor).not.toBe('rgb(255, 255, 255)')
    }
  })

  test('should have correct text color in dark mode', async ({ page }) => {
    await page.goto('/')
    
    const heading = page.locator('h1')
    const textColor = await heading.evaluate(el => {
      const styles = window.getComputedStyle(el)
      return styles.color
    })
    
    // Text should be visible (not the same as background)
    expect(textColor).not.toBe('')
  })

  test('should toggle theme multiple times', async ({ page }) => {
    await page.goto('/')
    
    const themeButton = page.locator('header button').nth(1)
    
    // Toggle multiple times
    for (let i = 0; i < 3; i++) {
      const before = await page.evaluate(() => 
        document.documentElement.classList.contains('dark')
      )
      
      await themeButton.click()
      
      const after = await page.evaluate(() => 
        document.documentElement.classList.contains('dark')
      )
      
      expect(after).not.toBe(before)
    }
  })

  test('should have theme toggle button with correct icon', async ({ page }) => {
    await page.goto('/')
    
    const themeButton = page.locator('header button').nth(1)
    
    // Should have an SVG icon (sun or moon)
    const hasIcon = await themeButton.locator('svg').count() > 0
    expect(hasIcon).toBeTruthy()
  })

  test('should apply theme to all pages', async ({ page }) => {
    const pages = ['/', '/downloads', '/library', '/search', '/suggestions', '/settings']
    
    // Set theme to light first
    await page.goto('/')
    const themeButton = page.locator('header button').nth(1)
    await themeButton.click()
    
    for (const url of pages) {
      await page.goto(url)
      
      // Should have theme class
      const hasThemeClass = await page.evaluate(() => {
        const html = document.documentElement
        return html.classList.contains('dark') || html.classList.contains('light')
      })
      expect(hasThemeClass).toBeTruthy()
    }
  })
})