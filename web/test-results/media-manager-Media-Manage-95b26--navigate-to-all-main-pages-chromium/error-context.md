# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: media-manager.spec.ts >> Media Manager - Navigation >> should navigate to all main pages
- Location: e2e/media-manager.spec.ts:38:3

# Error details

```
Error: expect(received).toContain(expected) // indexOf

Expected substring: "home"
Received string:    "http://localhost:5178/"
```

# Page snapshot

```yaml
- generic [ref=e3]:
  - banner [ref=e4]:
    - generic [ref=e5]:
      - button [ref=e6] [cursor=pointer]:
        - img [ref=e7]
      - generic [ref=e9]: 📺 Media Manager
      - generic [ref=e10]:
        - generic "VPN is not active" [ref=e11]:
          - img [ref=e12]
          - generic [ref=e14]: VPN Disconnected
        - button [ref=e15] [cursor=pointer]:
          - img [ref=e16]
        - button [ref=e22] [cursor=pointer]:
          - img [ref=e23]
  - generic [ref=e26]:
    - navigation [ref=e27]:
      - generic [ref=e28]:
        - link "Home" [active] [ref=e29] [cursor=pointer]:
          - /url: /
          - img [ref=e30]
          - text: Home
        - link "Discover" [ref=e33] [cursor=pointer]:
          - /url: /discover
          - img [ref=e34]
          - text: Discover
        - link "Watchlist" [ref=e37] [cursor=pointer]:
          - /url: /watchlist
          - img [ref=e38]
          - text: Watchlist
        - link "Downloads" [ref=e40] [cursor=pointer]:
          - /url: /downloads
          - img [ref=e41]
          - text: Downloads
        - link "Library" [ref=e44] [cursor=pointer]:
          - /url: /library
          - img [ref=e45]
          - text: Library
        - link "Search" [ref=e47] [cursor=pointer]:
          - /url: /search
          - img [ref=e48]
          - text: Search
        - link "Suggestions" [ref=e51] [cursor=pointer]:
          - /url: /suggestions
          - img [ref=e52]
          - text: Suggestions
        - link "Settings" [ref=e54] [cursor=pointer]:
          - /url: /settings
          - img [ref=e55]
          - text: Settings
    - main [ref=e58]:
      - generic [ref=e59]:
        - img [ref=e60]
        - generic [ref=e62]:
          - paragraph [ref=e63]: VPN Disconnected
          - paragraph [ref=e64]: Downloads are disabled for your security. Please connect to your VPN to resume downloading.
        - button "Dismiss warning" [ref=e65] [cursor=pointer]:
          - img [ref=e66]
      - generic [ref=e69]:
        - generic [ref=e70]:
          - heading "Dashboard" [level=1] [ref=e71]
          - paragraph [ref=e72]: Welcome to your media manager. Monitor your downloads and library.
        - generic [ref=e73]:
          - generic [ref=e74]:
            - generic [ref=e75]:
              - heading "Pending" [level=3] [ref=e76]
              - img [ref=e77]
            - generic [ref=e81]: "0"
          - generic [ref=e82]:
            - generic [ref=e83]:
              - heading "Downloading" [level=3] [ref=e84]
              - img [ref=e85]
            - generic [ref=e88]: "0"
          - generic [ref=e89]:
            - generic [ref=e90]:
              - heading "Completed" [level=3] [ref=e91]
              - img [ref=e92]
            - generic [ref=e96]: "0"
          - generic [ref=e97]:
            - generic [ref=e98]:
              - heading "Failed" [level=3] [ref=e99]
              - img [ref=e100]
            - generic [ref=e105]: "0"
        - generic [ref=e106]:
          - generic [ref=e107] [cursor=pointer]:
            - generic [ref=e108]:
              - heading "Downloads" [level=3] [ref=e109]:
                - img [ref=e110]
                - text: Downloads
              - paragraph [ref=e113]: Manage your download queue
            - paragraph [ref=e115]: View active downloads, cancel or restart tasks, and monitor progress.
          - generic [ref=e116] [cursor=pointer]:
            - generic [ref=e117]:
              - heading "Library" [level=3] [ref=e118]:
                - img [ref=e119]
                - text: Library
              - paragraph [ref=e121]: Browse your media collection
            - paragraph [ref=e123]: View downloaded movies, fetch posters, and organize your library.
          - generic [ref=e124] [cursor=pointer]:
            - generic [ref=e125]:
              - heading "Search" [level=3] [ref=e126]:
                - img [ref=e127]
                - text: Search
              - paragraph [ref=e130]: Find new content
            - paragraph [ref=e132]: Search across multiple indexers and approve downloads.
```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test'
  2   | 
  3   | const BASE_URL = process.env.BASE_URL || 'http://localhost:5178'
  4   | 
  5   | // Helper to wait for API response
  6   | async function waitForAPIResponse(page: import('@playwright/test').Page, urlPattern: string) {
  7   |   return page.waitForResponse(response => response.url().includes(urlPattern))
  8   | }
  9   | 
  10  | test.describe('Media Manager - Auth', () => {
  11  |   test('should show login page', async ({ page }) => {
  12  |     await page.goto(`${BASE_URL}/login`)
  13  |     
  14  |     // Check login form elements
  15  |     await expect(page.locator('h1:has-text("Media Manager")')).toBeVisible()
  16  |     await expect(page.locator('input[placeholder="Enter your username"]')).toBeVisible()
  17  |     await expect(page.locator('input[placeholder="Enter your password"]')).toBeVisible()
  18  |     await expect(page.locator('button:has-text("Sign In")')).toBeVisible()
  19  |   })
  20  | 
  21  |   test('should toggle between login and register', async ({ page }) => {
  22  |     await page.goto(`${BASE_URL}/login`)
  23  |     
  24  |     // Click to register
  25  |     await page.click('text=Don\'t have an account? Create one')
  26  |     
  27  |     await expect(page.locator('h2:has-text("Create Account")')).toBeVisible()
  28  |     await expect(page.locator('button:has-text("Create Account")')).toBeVisible()
  29  |     
  30  |     // Click back to login
  31  |     await page.click('text=Already have an account? Sign in')
  32  |     
  33  |     await expect(page.locator('h2:has-text("Sign In")')).toBeVisible()
  34  |   })
  35  | })
  36  | 
  37  | test.describe('Media Manager - Navigation', () => {
  38  |   test('should navigate to all main pages', async ({ page }) => {
  39  |     await page.goto(BASE_URL)
  40  |     
  41  |     // Check sidebar navigation - includes Watchlist now
  42  |     const navItems = ['Home', 'Discover', 'Watchlist', 'Downloads', 'Library', 'Search', 'Suggestions', 'Settings']
  43  |     
  44  |     for (const item of navItems) {
  45  |       await page.click(`text=${item}`)
  46  |       await page.waitForLoadState('networkidle')
  47  |       
  48  |       // Verify we're on the right page by checking URL
  49  |       const url = page.url()
> 50  |       expect(url).toContain(item.toLowerCase())
      |                   ^ Error: expect(received).toContain(expected) // indexOf
  51  |     }
  52  |   })
  53  | })
  54  | 
  55  | test.describe('Media Manager - Discover Page', () => {
  56  |   test('should load discover page with content', async ({ page }) => {
  57  |     await page.goto(`${BASE_URL}/discover`)
  58  |     
  59  |     // Wait for API responses
  60  |     await waitForAPIResponse(page, '/api/discover/movies/trending')
  61  |     
  62  |     // Check page title
  63  |     await expect(page.locator('h1:has-text("Discover")')).toBeVisible()
  64  |     
  65  |     // Check tabs
  66  |     await expect(page.locator('button:has-text("All")')).toBeVisible()
  67  |     await expect(page.locator('button:has-text("MOVIES")')).toBeVisible()
  68  |     await expect(page.locator('button:has-text("TV SHOWS")')).toBeVisible()
  69  |     
  70  |     // Check content sections
  71  |     await expect(page.locator('text=Trending Movies')).toBeVisible()
  72  |     await expect(page.locator('text=Popular Movies')).toBeVisible()
  73  |   })
  74  | 
  75  |   test('should filter discover by movies tab', async ({ page }) => {
  76  |     await page.goto(`${BASE_URL}/discover`)
  77  |     
  78  |     await waitForAPIResponse(page, '/api/discover/movies/trending')
  79  |     
  80  |     // Click Movies tab
  81  |     await page.click('button:has-text("MOVIES")')
  82  |     
  83  |     await page.waitForTimeout(500)
  84  |     
  85  |     // Should show movie sections only
  86  |     await expect(page.locator('text=Trending Movies')).toBeVisible()
  87  |     await expect(page.locator('text=Popular Movies')).toBeVisible()
  88  |     
  89  |     // Should NOT show TV sections
  90  |     const tvContent = await page.locator('text=Trending TV Shows').count()
  91  |     expect(tvContent).toBe(0)
  92  |   })
  93  | 
  94  |   test('should filter discover by TV tab', async ({ page }) => {
  95  |     await page.goto(`${BASE_URL}/discover`)
  96  |     
  97  |     await waitForAPIResponse(page, '/api/discover/movies/trending')
  98  |     
  99  |     // Click TV Shows tab
  100 |     await page.click('button:has-text("TV SHOWS")')
  101 |     
  102 |     await page.waitForTimeout(500)
  103 |     
  104 |     // Should show TV sections
  105 |     await expect(page.locator('text=Trending TV Shows')).toBeVisible()
  106 |     await expect(page.locator('text=Popular TV Shows')).toBeVisible()
  107 |   })
  108 | 
  109 |   test('should show movie cards with ratings', async ({ page }) => {
  110 |     await page.goto(`${BASE_URL}/discover`)
  111 |     
  112 |     await waitForAPIResponse(page, '/api/discover/movies/trending')
  113 |     
  114 |     // Check for movie cards with star ratings
  115 |     const starRatings = page.locator('.text-yellow-400')
  116 |     await expect(starRatings.first()).toBeVisible()
  117 |     
  118 |     // Check for movie titles
  119 |     const movieTitles = page.locator('h3.text-white')
  120 |     await expect(movieTitles.first()).toBeVisible()
  121 |   })
  122 | })
  123 | 
  124 | test.describe('Media Manager - Movie Detail', () => {
  125 |   test('should navigate to movie detail page', async ({ page }) => {
  126 |     await page.goto(`${BASE_URL}/discover`)
  127 |     
  128 |     await waitForAPIResponse(page, '/api/discover/movies/trending')
  129 |     
  130 |     // Click first movie card
  131 |     await page.click('a[href^="/movie/"]')
  132 |     
  133 |     // Wait for movie detail API
  134 |     await waitForAPIResponse(page, '/api/discover/movie/')
  135 |     
  136 |     // Check movie detail elements
  137 |     await expect(page.locator('h1')).toBeVisible()
  138 |     await expect(page.locator('text=Overview')).toBeVisible()
  139 |     await expect(page.locator('text=Cast')).toBeVisible()
  140 |   })
  141 | 
  142 |   test('should show movie info sidebar', async ({ page }) => {
  143 |     await page.goto(`${BASE_URL}/movie/550`) // Fight Club as example
  144 |     
  145 |     await waitForAPIResponse(page, '/api/discover/movie/')
  146 |     
  147 |     // Check info sidebar
  148 |     await expect(page.locator('h3:has-text("Movie Info")')).toBeVisible()
  149 |     await expect(page.locator('text=Release Date')).toBeVisible()
  150 |     await expect(page.locator('text=Runtime')).toBeVisible()
```