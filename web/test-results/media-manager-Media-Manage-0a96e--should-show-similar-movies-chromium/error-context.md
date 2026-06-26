# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: media-manager.spec.ts >> Media Manager - Movie Detail >> should show similar movies
- Location: e2e/media-manager.spec.ts:155:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('h2:has-text("Similar Movies")')
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('h2:has-text("Similar Movies")')

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
        - link "Home" [ref=e29] [cursor=pointer]:
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
          - link [ref=e73] [cursor=pointer]:
            - /url: /discover
            - img [ref=e74]
          - generic [ref=e77]:
            - heading "Fight Club" [level=1] [ref=e78]
            - paragraph [ref=e79]: Mischief. Mayhem. Soap.
            - generic [ref=e80]:
              - generic [ref=e81]:
                - img [ref=e82]
                - generic [ref=e84]: "8.4"
                - generic [ref=e85]: (32,209 votes)
              - generic [ref=e86]: "|"
              - generic [ref=e87]: "1999"
              - generic [ref=e88]: "|"
              - generic [ref=e89]: 139 min
              - generic [ref=e90]: "|"
              - generic [ref=e91]: Released
            - generic [ref=e92]:
              - generic [ref=e93]: Drama
              - generic [ref=e94]: Thriller
            - generic [ref=e95]:
              - button "Add to Watchlist" [ref=e96] [cursor=pointer]:
                - img [ref=e97]
                - text: Add to Watchlist
              - button "Download" [ref=e99] [cursor=pointer]:
                - img [ref=e100]
                - text: Download
        - generic [ref=e104]:
          - generic [ref=e105]:
            - generic [ref=e106]:
              - heading "Overview" [level=2] [ref=e107]
              - paragraph [ref=e108]: A ticking-time-bomb insomniac and a slippery soap salesman channel primal male aggression into a shocking new form of therapy. Their concept catches on, with underground "fight clubs" forming in every town, until an eccentric gets in the way and ignites an out-of-control spiral toward oblivion.
            - generic [ref=e109]:
              - heading "Cast" [level=2] [ref=e110]
              - generic [ref=e111]:
                - generic [ref=e112]:
                  - img "Edward Norton" [ref=e114]
                  - paragraph [ref=e115]: Edward Norton
                  - paragraph [ref=e116]: Narrator
                - generic [ref=e117]:
                  - img "Brad Pitt" [ref=e119]
                  - paragraph [ref=e120]: Brad Pitt
                  - paragraph [ref=e121]: Tyler Durden
                - generic [ref=e122]:
                  - img "Helena Bonham Carter" [ref=e124]
                  - paragraph [ref=e125]: Helena Bonham Carter
                  - paragraph [ref=e126]: Marla Singer
                - generic [ref=e127]:
                  - img "Meat Loaf" [ref=e129]
                  - paragraph [ref=e130]: Meat Loaf
                  - paragraph [ref=e131]: Robert Paulson
                - generic [ref=e132]:
                  - img "Jared Leto" [ref=e134]
                  - paragraph [ref=e135]: Jared Leto
                  - paragraph [ref=e136]: Angel Face
                - generic [ref=e137]:
                  - img "Zach Grenier" [ref=e139]
                  - paragraph [ref=e140]: Zach Grenier
                  - paragraph [ref=e141]: Richard Chesler (Regional Manager)
                - generic [ref=e142]:
                  - img "Holt McCallany" [ref=e144]
                  - paragraph [ref=e145]: Holt McCallany
                  - paragraph [ref=e146]: The Mechanic
                - generic [ref=e147]:
                  - img "Eion Bailey" [ref=e149]
                  - paragraph [ref=e150]: Eion Bailey
                  - paragraph [ref=e151]: Ricky
                - generic [ref=e152]:
                  - img "Richmond Arquette" [ref=e154]
                  - paragraph [ref=e155]: Richmond Arquette
                  - paragraph [ref=e156]: Intern at Hospital
                - generic [ref=e157]:
                  - img [ref=e160]
                  - paragraph [ref=e162]: David Andrews
                  - paragraph [ref=e163]: Thomas at Remaining Men Together
            - generic [ref=e164]:
              - heading "Crew" [level=2] [ref=e165]
              - generic [ref=e166]:
                - generic [ref=e167]:
                  - paragraph [ref=e168]: Director
                  - paragraph [ref=e169]: David Fincher
                - generic [ref=e170]:
                  - paragraph [ref=e171]: Writers
                  - paragraph [ref=e172]: Jim Uhls
          - generic [ref=e174]:
            - heading "Movie Info" [level=3] [ref=e175]
            - generic [ref=e176]:
              - generic [ref=e177]:
                - paragraph [ref=e178]: Release Date
                - paragraph [ref=e179]: 1999-10-15
              - generic [ref=e180]:
                - paragraph [ref=e181]: Runtime
                - paragraph [ref=e182]: 2h 19m
              - generic [ref=e183]:
                - paragraph [ref=e184]: Budget
                - paragraph [ref=e185]: $63.0M
              - generic [ref=e186]:
                - paragraph [ref=e187]: Revenue
                - paragraph [ref=e188]: $100.9M
              - generic [ref=e189]:
                - paragraph [ref=e190]: Status
                - paragraph [ref=e191]: Released
```

# Test source

```ts
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
  151 |     await expect(page.locator('text=Budget')).toBeVisible()
  152 |     await expect(page.locator('text=Revenue')).toBeVisible()
  153 |   })
  154 | 
  155 |   test('should show similar movies', async ({ page }) => {
  156 |     await page.goto(`${BASE_URL}/movie/550`)
  157 |     
  158 |     await waitForAPIResponse(page, '/api/discover/movie/')
  159 |     
  160 |     // Check similar movies section
> 161 |     await expect(page.locator('h2:has-text("Similar Movies")')).toBeVisible()
      |                                                                 ^ Error: expect(locator).toBeVisible() failed
  162 |   })
  163 | })
  164 | 
  165 | test.describe('Media Manager - TV Detail', () => {
  166 |   test('should navigate to TV detail page', async ({ page }) => {
  167 |     await page.goto(`${BASE_URL}/discover`)
  168 |     
  169 |     await waitForAPIResponse(page, '/api/discover/tv/trending')
  170 |     
  171 |     // Click TV Shows tab first
  172 |     await page.click('button:has-text("TV SHOWS")')
  173 |     await page.waitForTimeout(500)
  174 |     
  175 |     // Click first TV card
  176 |     await page.click('a[href^="/tv/"]')
  177 |     
  178 |     // Wait for TV detail API
  179 |     await waitForAPIResponse(page, '/api/discover/tv/')
  180 |     
  181 |     // Check TV detail elements
  182 |     await expect(page.locator('h1')).toBeVisible()
  183 |     await expect(page.locator('text=Overview')).toBeVisible()
  184 |     await expect(page.locator('text=Episodes')).toBeVisible()
  185 |   })
  186 | 
  187 |   test('should show season selector', async ({ page }) => {
  188 |     await page.goto(`${BASE_URL}/tv/1399`) // Game of Thrones as example
  189 |     
  190 |     await waitForAPIResponse(page, '/api/discover/tv/')
  191 |     
  192 |     // Check season selector
  193 |     await expect(page.locator('select')).toBeVisible()
  194 |     await expect(page.locator('text=Show Info')).toBeVisible()
  195 |   })
  196 | })
  197 | 
  198 | test.describe('Media Manager - Search', () => {
  199 |   test('should search for movies', async ({ page }) => {
  200 |     await page.goto(`${BASE_URL}/search`)
  201 |     
  202 |     // Type search query
  203 |     await page.fill('input[type="text"]', 'Inception')
  204 |     
  205 |     // Submit search
  206 |     await page.keyboard.press('Enter')
  207 |     
  208 |     // Wait for search results
  209 |     await page.waitForTimeout(2000)
  210 |     
  211 |     // Check results loaded
  212 |     await expect(page.locator('text=Inception').first()).toBeVisible()
  213 |   })
  214 | })
  215 | 
  216 | test.describe('Media Manager - Responsive', () => {
  217 |   test('should be responsive on mobile', async ({ page }) => {
  218 |     await page.setViewportSize({ width: 375, height: 667 })
  219 |     await page.goto(`${BASE_URL}/discover`)
  220 |     
  221 |     await waitForAPIResponse(page, '/api/discover/movies/trending')
  222 |     
  223 |     // Check that content is visible on mobile
  224 |     await expect(page.locator('h1:has-text("Discover")')).toBeVisible()
  225 |     
  226 |     // Check horizontal scrolling on movie rows
  227 |     const movieRow = page.locator('.overflow-x-auto').first()
  228 |     await expect(movieRow).toBeVisible()
  229 |   })
  230 | 
  231 |   test('should be responsive on tablet', async ({ page }) => {
  232 |     await page.setViewportSize({ width: 768, height: 1024 })
  233 |     await page.goto(`${BASE_URL}/discover`)
  234 |     
  235 |     await waitForAPIResponse(page, '/api/discover/movies/trending')
  236 |     
  237 |     await expect(page.locator('h1:has-text("Discover")')).toBeVisible()
  238 |   })
  239 | })
  240 | 
  241 | test.describe('Media Manager - Accessibility', () => {
  242 |   test('should have proper heading structure', async ({ page }) => {
  243 |     await page.goto(`${BASE_URL}/discover`)
  244 |     
  245 |     await waitForAPIResponse(page, '/api/discover/movies/trending')
  246 |     
  247 |     // Check for h1
  248 |     const h1 = await page.locator('h1').count()
  249 |     expect(h1).toBeGreaterThan(0)
  250 |     
  251 |     // Check for h2 sections
  252 |     const h2 = await page.locator('h2').count()
  253 |     expect(h2).toBeGreaterThan(0)
  254 |   })
  255 | 
  256 |   test('should have clickable navigation links', async ({ page }) => {
  257 |     await page.goto(BASE_URL)
  258 |     
  259 |     // Check all nav links are clickable
  260 |     const links = await page.locator('nav a').all()
  261 |     expect(links.length).toBeGreaterThan(0)
```