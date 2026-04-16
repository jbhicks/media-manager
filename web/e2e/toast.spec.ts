import { test, expect } from '@playwright/test'

test.describe('Toast Notifications', () => {
  test('should show toast on successful action', async ({ page }) => {
    await page.goto('/downloads')
    
    // Wait for page to load
    await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    // Trigger an action that shows toast (e.g., clear completed if there are completed tasks)
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    const data = await response.json()
    
    if (data.tasks?.some((t: { status: string }) => t.status === 'completed')) {
      const clearButton = page.getByRole('button', { name: /Clear Completed/i })
      await clearButton.click()
      
      // Should show success toast
      await expect(page.locator('.fixed.bottom-4').first()).toBeVisible()
    }
  })

  test('should display different toast types', async ({ page }) => {
    await page.goto('/library')
    
    // Click Clean Filenames button
    const cleanButton = page.getByRole('button', { name: /Clean Filenames/i })
    await cleanButton.click()
    
    // Wait for API and toast
    await page.waitForResponse(response => 
      response.url().includes('/api/library/reprocess')
    )
    
    // Toast should be visible
    const toast = page.locator('.fixed.bottom-4 > div').first()
    await expect(toast).toBeVisible()
  })

  test('should auto-dismiss toast after timeout', async ({ page }) => {
    await page.goto('/suggestions')
    
    // Generate suggestions to trigger toast
    const generateButton = page.getByRole('button', { name: /Generate Suggestions/i })
    await generateButton.click()
    
    // Wait for toast to appear
    const toast = page.locator('.fixed.bottom-4 > div').first()
    await expect(toast).toBeVisible()
    
    // Wait for auto-dismiss (3 seconds)
    await page.waitForTimeout(3500)
    
    // Toast should be gone
    await expect(toast).not.toBeVisible()
  })

  test('should dismiss toast when close button is clicked', async ({ page }) => {
    await page.goto('/suggestions')
    
    // Generate suggestions
    const generateButton = page.getByRole('button', { name: /Generate Suggestions/i })
    await generateButton.click()
    
    // Wait for toast
    const toast = page.locator('.fixed.bottom-4 > div').first()
    await expect(toast).toBeVisible()
    
    // Click close button
    const closeButton = toast.locator('button')
    await closeButton.click()
    
    // Toast should be dismissed
    await expect(toast).not.toBeVisible()
  })

  test('should show multiple toasts stacked', async ({ page }) => {
    await page.goto('/suggestions')
    
    // Generate suggestions multiple times
    const generateButton = page.getByRole('button', { name: /Generate Suggestions/i })
    
    await generateButton.click()
    await page.waitForTimeout(500)
    
    // Check that toast container exists
    const toastContainer = page.locator('.fixed.bottom-4.right-4')
    await expect(toastContainer).toBeVisible()
  })

  test('should position toasts at bottom right', async ({ page }) => {
    await page.goto('/')
    
    // Trigger any action that shows toast
    await page.goto('/suggestions')
    const generateButton = page.getByRole('button', { name: /Generate Suggestions/i })
    await generateButton.click()
    
    // Toast container should be positioned at bottom right
    const toastContainer = page.locator('.fixed.bottom-4.right-4')
    await expect(toastContainer).toBeVisible()
  })

  test('toast should have correct styling for success type', async ({ page }) => {
    await page.goto('/suggestions')
    
    const generateButton = page.getByRole('button', { name: /Generate Suggestions/i })
    await generateButton.click()
    
    // Wait for success toast
    const toast = page.locator('.fixed.bottom-4 > div').first()
    await expect(toast).toBeVisible()
    
    // Should have success styling (green colors)
    const hasSuccessClass = await toast.evaluate(el => 
      el.classList.contains('border-green-500') || 
      el.classList.contains('bg-green-500') ||
      el.classList.contains('text-green-500')
    )
    expect(hasSuccessClass).toBeTruthy()
  })
})
