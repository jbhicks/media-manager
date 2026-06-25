import { test, expect } from '@playwright/test'

const BASE_URL = process.env.BASE_URL || 'http://localhost:5174'

// Helper to wait for API response
async function waitForAPIResponse(page: import('@playwright/test').Page, urlPattern: string) {
  return page.waitForResponse(response => response.url().includes(urlPattern))
}

test.describe('Media Manager - Auth', () => {
  test('should show login page', async ({ page }) => {
    await page.goto(`${BASE_URL}/login`)
    
    // Check login form elements
    await expect(page.locator('h1:has-text("Media Manager")')).toBeVisible()
    await expect(page.locator('input[placeholder="Enter your username"]')).toBeVisible()
    await expect(page.locator('input[placeholder="Enter your password"]')).toBeVisible()
    await expect(page.locator('button:has-text("Sign In")')).toBeVisible()
  })

  test('should toggle between login and register', async ({ page }) => {
    await page.goto(`${BASE_URL}/login`)
    
    // Click to register
    await page.click('text=Don\'t have an account? Create one')
    
    await expect(page.locator('h2:has-text("Create Account")')).toBeVisible()
    await expect(page.locator('button:has-text("Create Account")')).toBeVisible()
    
    // Click back to login
    await page.click('text=Already have an account? Sign in')
    
    await expect(page.locator('h2:has-text("Sign In")')).toBeVisible()
  })
})

test.describe('Media Manager - Navigation', () => {
  test('should navigate to all main pages', async ({ page }) => {
    await page.goto(BASE_URL)
    
    // Check sidebar navigation
    const navItems = ['Home', 'Discover', 'Downloads', 'Library', 'Search', 'Suggestions', 'Settings']
    
    for (const item of navItems) {
      await page.click(`text=${item}`)
      await page.waitForLoadState('networkidle')
      
      // Verify we're on the right page by checking URL
      const url = page.url()
      expect(url).toContain(item.toLowerCase())
    }
  })
})

test.describe('Media Manager - Discover Page', () => {
  test('should load discover page with content', async ({ page }) => {
    await page.goto(`${BASE_URL}/discover`)
    
    // Wait for API responses
    await waitForAPIResponse(page, '/api/discover/movies/trending')
    
    // Check page title
    await expect(page.locator('h1:has-text("Discover")')).toBeVisible()
    
    // Check tabs
    await expect(page.locator('button:has-text("All")')).toBeVisible()
    await expect(page.locator('button:has-text("Movies")')).toBeVisible()
    await expect(page.locator('button:has-text("TV Shows")')).toBeVisible()
    
    // Check content sections
    await expect(page.locator('text=Trending Movies')).toBeVisible()
    await expect(page.locator('text=Popular Movies')).toBeVisible()
  })

  test('should filter discover by movies tab', async ({ page }) => {
    await page.goto(`${BASE_URL}/discover`)
    
    await waitForAPIResponse(page, '/api/discover/movies/trending')
    
    // Click Movies tab
    await page.click('button:has-text("Movies")')
    
    await page.waitForTimeout(500)
    
    // Should show movie sections only
    await expect(page.locator('text=Trending Movies')).toBeVisible()
    await expect(page.locator('text=Popular Movies')).toBeVisible()
    
    // Should NOT show TV sections
    const tvContent = await page.locator('text=Trending TV Shows').count()
    expect(tvContent).toBe(0)
  })

  test('should filter discover by TV tab', async ({ page }) => {
    await page.goto(`${BASE_URL}/discover`)
    
    await waitForAPIResponse(page, '/api/discover/movies/trending')
    
    // Click TV Shows tab
    await page.click('button:has-text("TV Shows")')
    
    await page.waitForTimeout(500)
    
    // Should show TV sections
    await expect(page.locator('text=Trending TV Shows')).toBeVisible()
    await expect(page.locator('text=Popular TV Shows')).toBeVisible()
  })

  test('should show movie cards with ratings', async ({ page }) => {
    await page.goto(`${BASE_URL}/discover`)
    
    await waitForAPIResponse(page, '/api/discover/movies/trending')
    
    // Check for movie cards with star ratings
    const starRatings = page.locator('.text-yellow-400')
    await expect(starRatings.first()).toBeVisible()
    
    // Check for movie titles
    const movieTitles = page.locator('h3.text-white')
    await expect(movieTitles.first()).toBeVisible()
  })
})

test.describe('Media Manager - Movie Detail', () => {
  test('should navigate to movie detail page', async ({ page }) => {
    await page.goto(`${BASE_URL}/discover`)
    
    await waitForAPIResponse(page, '/api/discover/movies/trending')
    
    // Click first movie card
    await page.click('a[href^="/movie/"]')
    
    // Wait for movie detail API
    await waitForAPIResponse(page, '/api/discover/movie/')
    
    // Check movie detail elements
    await expect(page.locator('h1')).toBeVisible()
    await expect(page.locator('text=Overview')).toBeVisible()
    await expect(page.locator('text=Cast')).toBeVisible()
  })

  test('should show movie info sidebar', async ({ page }) => {
    await page.goto(`${BASE_URL}/movie/550`) // Fight Club as example
    
    await waitForAPIResponse(page, '/api/discover/movie/')
    
    // Check info sidebar
    await expect(page.locator('h3:has-text("Movie Info")')).toBeVisible()
    await expect(page.locator('text=Release Date')).toBeVisible()
    await expect(page.locator('text=Runtime')).toBeVisible()
    await expect(page.locator('text=Budget')).toBeVisible()
    await expect(page.locator('text=Revenue')).toBeVisible()
  })

  test('should show similar movies', async ({ page }) => {
    await page.goto(`${BASE_URL}/movie/550`)
    
    await waitForAPIResponse(page, '/api/discover/movie/')
    
    // Check similar movies section
    await expect(page.locator('h2:has-text("Similar Movies")')).toBeVisible()
  })
})

test.describe('Media Manager - TV Detail', () => {
  test('should navigate to TV detail page', async ({ page }) => {
    await page.goto(`${BASE_URL}/discover`)
    
    await waitForAPIResponse(page, '/api/discover/tv/trending')
    
    // Click TV Shows tab first
    await page.click('button:has-text("TV Shows")')
    await page.waitForTimeout(500)
    
    // Click first TV card
    await page.click('a[href^="/tv/"]')
    
    // Wait for TV detail API
    await waitForAPIResponse(page, '/api/discover/tv/')
    
    // Check TV detail elements
    await expect(page.locator('h1')).toBeVisible()
    await expect(page.locator('text=Overview')).toBeVisible()
    await expect(page.locator('text=Episodes')).toBeVisible()
  })

  test('should show season selector', async ({ page }) => {
    await page.goto(`${BASE_URL}/tv/1399`) // Game of Thrones as example
    
    await waitForAPIResponse(page, '/api/discover/tv/')
    
    // Check season selector
    await expect(page.locator('select')).toBeVisible()
    await expect(page.locator('text=Show Info')).toBeVisible()
  })
})

test.describe('Media Manager - Search', () => {
  test('should search for movies', async ({ page }) => {
    await page.goto(`${BASE_URL}/search`)
    
    // Type search query
    await page.fill('input[type="text"]', 'Inception')
    
    // Submit search
    await page.keyboard.press('Enter')
    
    // Wait for search results
    await page.waitForTimeout(2000)
    
    // Check results loaded
    await expect(page.locator('text=Inception').first()).toBeVisible()
  })
})

test.describe('Media Manager - Responsive', () => {
  test('should be responsive on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto(`${BASE_URL}/discover`)
    
    await waitForAPIResponse(page, '/api/discover/movies/trending')
    
    // Check that content is visible on mobile
    await expect(page.locator('h1:has-text("Discover")')).toBeVisible()
    
    // Check horizontal scrolling on movie rows
    const movieRow = page.locator('.overflow-x-auto').first()
    await expect(movieRow).toBeVisible()
  })

  test('should be responsive on tablet', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 })
    await page.goto(`${BASE_URL}/discover`)
    
    await waitForAPIResponse(page, '/api/discover/movies/trending')
    
    await expect(page.locator('h1:has-text("Discover")')).toBeVisible()
  })
})

test.describe('Media Manager - Accessibility', () => {
  test('should have proper heading structure', async ({ page }) => {
    await page.goto(`${BASE_URL}/discover`)
    
    await waitForAPIResponse(page, '/api/discover/movies/trending')
    
    // Check for h1
    const h1 = await page.locator('h1').count()
    expect(h1).toBeGreaterThan(0)
    
    // Check for h2 sections
    const h2 = await page.locator('h2').count()
    expect(h2).toBeGreaterThan(0)
  })

  test('should have clickable navigation links', async ({ page }) => {
    await page.goto(BASE_URL)
    
    // Check all nav links are clickable
    const links = await page.locator('nav a').all()
    expect(links.length).toBeGreaterThan(0)
    
    for (const link of links) {
      await expect(link).toBeVisible()
    }
  })
})
