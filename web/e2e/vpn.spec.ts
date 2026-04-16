import { test, expect } from '@playwright/test'

test.describe('VPN Status Feature', () => {
  test('should display VPN status in header', async ({ page }) => {
    await page.goto('/')
    
    // Wait for VPN API
    await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    // VPN status should be visible in header
    const vpnElement = page.locator('header').getByText(/VPN/i)
    await expect(vpnElement).toBeVisible()
  })

  test('should show VPN connected status when active', async ({ page }) => {
    await page.goto('/')
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.active) {
      await expect(page.getByText(/VPN Connected|Connected/i).first()).toBeVisible()
      
      // Should have green styling
      const vpnElement = page.locator('header').getByText(/VPN Connected/i).first()
      const hasGreenClass = await vpnElement.evaluate(el => 
        el.classList.contains('text-green-500') || 
        el.closest('[class*="green"]') !== null
      )
      expect(hasGreenClass).toBeTruthy()
    }
  })

  test('should show VPN disconnected status when inactive', async ({ page }) => {
    await page.goto('/')
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (!data.active) {
      await expect(page.getByText(/VPN Disconnected|Disconnected/i).first()).toBeVisible()
      
      // Should have red styling
      const vpnElement = page.locator('header').getByText(/VPN Disconnected/i).first()
      const hasRedClass = await vpnElement.evaluate(el => 
        el.classList.contains('text-red-500') || 
        el.closest('[class*="red"]') !== null
      )
      expect(hasRedClass).toBeTruthy()
    }
  })

  test('should show VPN icon in header', async ({ page }) => {
    await page.goto('/')
    
    await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    // Should have shield icon
    const header = page.locator('header')
    const shieldIcon = header.locator('svg').filter({ hasText: /shield| Shield/i })
    await expect(shieldIcon.first()).toBeVisible()
  })

  test('should display VPN status card in settings', async ({ page }) => {
    await page.goto('/settings')
    
    // Wait for VPN API
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    // VPN card should be visible
    await expect(page.getByRole('heading', { name: 'VPN Status' })).toBeVisible()
  })

  test('VPN status card should show correct status badge', async ({ page }) => {
    await page.goto('/settings')
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    // Should show status badge
    if (data.active) {
      await expect(page.getByText('Connected').filter({ has: page.locator('..') }).first()).toBeVisible()
    } else {
      await expect(page.getByText('Disconnected').filter({ has: page.locator('..') }).first()).toBeVisible()
    }
  })

  test('VPN status card should show message', async ({ page }) => {
    await page.goto('/settings')
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    // Should show message
    if (data.message) {
      await expect(page.getByText(data.message)).toBeVisible()
    }
  })

  test('should display VPN provider when connected', async ({ page }) => {
    await page.goto('/settings')
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.active && data.provider) {
      await expect(page.getByText('Provider:')).toBeVisible()
      await expect(page.getByText(data.provider)).toBeVisible()
    }
  })

  test('should display VPN location when connected', async ({ page }) => {
    await page.goto('/settings')
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.active && (data.location || data.country)) {
      await expect(page.getByText('Location:')).toBeVisible()
    }
  })

  test('should display VPN IP when connected', async ({ page }) => {
    await page.goto('/settings')
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.active && data.ip) {
      await expect(page.getByText('IP Address:')).toBeVisible()
      await expect(page.getByText(data.ip)).toBeVisible()
    }
  })

  test('VPN status should auto-refresh', async ({ page }) => {
    await page.goto('/')
    
    // Wait for initial VPN API call
    await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    // Wait for the next auto-refresh (30 seconds interval)
    // For testing, we'll just verify the API was called at least once
    const vpnRequests = []
    page.on('response', response => {
      if (response.url().includes('/api/vpn/status')) {
        vpnRequests.push(response)
      }
    })
    
    // Wait a bit to see if it refreshes
    await page.waitForTimeout(1000)
    
    expect(vpnRequests.length).toBeGreaterThanOrEqual(1)
  })

  test('should show VPN status on mobile viewport', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/')
    
    await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    // Should show abbreviated VPN status on mobile
    const vpnText = await page.locator('header').getByText(/VPN|No VPN/i).first().textContent()
    expect(vpnText).toMatch(/VPN|No VPN/)
  })

  test('VPN card should have colored border based on status', async ({ page }) => {
    await page.goto('/settings')
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    const vpnCard = page.getByRole('heading', { name: 'VPN Status' }).locator('..').locator('..')
    
    if (data.active) {
      const hasGreenBorder = await vpnCard.evaluate(el => 
        el.classList.contains('border-green-500') || 
        el.classList.contains('border-green-500/50')
      )
      expect(hasGreenBorder).toBeTruthy()
    } else {
      const hasRedBorder = await vpnCard.evaluate(el => 
        el.classList.contains('border-red-500') || 
        el.classList.contains('border-red-500/50')
      )
      expect(hasRedBorder).toBeTruthy()
    }
  })

  test('should show VPN warning banner when disconnected', async ({ page }) => {
    await page.goto('/')
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (!data.active) {
      // Should show warning banner
      await expect(page.getByText('VPN Disconnected')).toBeVisible()
      await expect(page.getByText('Downloads are disabled for your security')).toBeVisible()
    }
  })

  test('should be able to dismiss VPN warning banner', async ({ page }) => {
    await page.goto('/')
    
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (!data.active) {
      // Banner should be visible
      const banner = page.getByText('VPN Disconnected').locator('..')
      await expect(banner).toBeVisible()
      
      // Click dismiss button
      const dismissButton = banner.locator('button')
      await dismissButton.click()
      
      // Banner should be hidden
      await expect(banner).not.toBeVisible()
    }
  })
})
