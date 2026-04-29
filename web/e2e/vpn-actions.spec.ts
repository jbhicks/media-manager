import { test, expect } from '@playwright/test'

test.describe('VPN Actions', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/settings')
    
    // Wait for VPN status to load
    await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
  })

  test('should display VPN status card', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'VPN Status' })).toBeVisible()
  })

  test('should show Connect VPN button when disconnected', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (!data.active) {
      await expect(page.getByRole('button', { name: /Connect VPN/i })).toBeVisible()
      await expect(page.getByRole('button', { name: /Connect VPN/i })).toBeEnabled()
    }
  })

  test('should show Disconnect VPN button when connected', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.active) {
      await expect(page.getByRole('button', { name: /Disconnect VPN/i })).toBeVisible()
      await expect(page.getByRole('button', { name: /Disconnect VPN/i })).toBeEnabled()
    }
  })

  test('should call connect API when Connect VPN is clicked', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (!data.active) {
      const connectButton = page.getByRole('button', { name: /Connect VPN/i })
      await connectButton.click()
      
      // Wait for connect API
      await page.waitForResponse(response => 
        response.url().includes('/api/vpn/connect') && response.status() === 200
      )
    }
  })

  test('should call disconnect API when Disconnect VPN is clicked', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.active) {
      const disconnectButton = page.getByRole('button', { name: /Disconnect VPN/i })
      await disconnectButton.click()
      
      // Wait for disconnect API
      await page.waitForResponse(response => 
        response.url().includes('/api/vpn/disconnect') && response.status() === 200
      )
    }
  })

  test('should show loading state during VPN connect', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (!data.active) {
      const connectButton = page.getByRole('button', { name: /Connect VPN/i })
      await connectButton.click()
      
      // Button should show loading or be disabled
      await expect(page.getByText(/Connecting/i)).toBeVisible()
    }
  })

  test('should show loading state during VPN disconnect', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.active) {
      const disconnectButton = page.getByRole('button', { name: /Disconnect VPN/i })
      await disconnectButton.click()
      
      // Button should show loading or be disabled
      await expect(page.getByText(/Disconnecting/i)).toBeVisible()
    }
  })

  test('should refresh VPN status after connect', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (!data.active) {
      const connectButton = page.getByRole('button', { name: /Connect VPN/i })
      await connectButton.click()
      
      // Wait for connect API
      await page.waitForResponse(response => 
        response.url().includes('/api/vpn/connect') && response.status() === 200
      )
      
      // Should refresh VPN status
      await page.waitForResponse(response => 
        response.url().includes('/api/vpn/status') && response.status() === 200
      )
    }
  })

  test('should refresh VPN status after disconnect', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.active) {
      const disconnectButton = page.getByRole('button', { name: /Disconnect VPN/i })
      await disconnectButton.click()
      
      // Wait for disconnect API
      await page.waitForResponse(response => 
        response.url().includes('/api/vpn/disconnect') && response.status() === 200
      )
      
      // Should refresh VPN status
      await page.waitForResponse(response => 
        response.url().includes('/api/vpn/status') && response.status() === 200
      )
    }
  })

  test('should disable VPN buttons during operation', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (!data.active) {
      const connectButton = page.getByRole('button', { name: /Connect VPN/i })
      await connectButton.click()
      
      // Button should be disabled during operation
      await expect(connectButton).toBeDisabled()
    }
  })

  test('should display VPN card with colored border', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/vpn/status') && response.status() === 200
    )
    
    const data = await response.json()
    
    const vpnCard = page.getByRole('heading', { name: 'VPN Status' }).locator('..').locator('..')
    
    if (data.active) {
      const hasGreenBorder = await vpnCard.evaluate(el => 
        el.classList.contains('border-green-500') || 
        el.classList.contains('border-green-500/50') ||
        getComputedStyle(el).borderColor.includes('green')
      )
      expect(hasGreenBorder).toBeTruthy()
    } else {
      const hasRedBorder = await vpnCard.evaluate(el => 
        el.classList.contains('border-red-500') || 
        el.classList.contains('border-red-500/50') ||
        getComputedStyle(el).borderColor.includes('red')
      )
      expect(hasRedBorder).toBeTruthy()
    }
  })

  test('should display auto-refresh message', async ({ page }) => {
    await expect(page.getByText(/automatically checked every 30 seconds/i)).toBeVisible()
  })
})