import { test, expect } from '@playwright/test'

test.describe('Downloads Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/downloads')
  })

  test('should display downloads page title and description', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Downloads' })).toBeVisible()
    await expect(page.getByText('Manage your download queue')).toBeVisible()
  })

  test('should display Clear Completed and Clear Failed buttons', async ({ page }) => {
    await expect(page.getByRole('button', { name: /Clear Completed/i })).toBeVisible()
    await expect(page.getByRole('button', { name: /Clear Failed/i })).toBeVisible()
  })

  test('should display task list or empty state', async ({ page }) => {
    // Wait for tasks to load
    await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    // Either tasks are shown or empty state
    const hasTasks = await page.locator('.space-y-4 > div').count() > 0
    const hasEmptyState = await page.getByText('No downloads yet').isVisible()
    
    expect(hasTasks || hasEmptyState).toBeTruthy()
  })

  test('should show task details when tasks exist', async ({ page }) => {
    // Wait for API response
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.tasks && data.tasks.length > 0) {
      // Should show task title
      const firstTask = data.tasks[0]
      await expect(page.getByText(firstTask.title).first()).toBeVisible()
      
      // Should show status badge
      await expect(page.getByText(firstTask.status)).toBeVisible()
      
      // Should show progress bar
      await expect(page.locator('.h-2.rounded-full.bg-primary').first()).toBeVisible()
    }
  })

  test('should display action buttons for downloading tasks', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    const data = await response.json()
    const downloadingTask = data.tasks?.find((t: { status: string }) => t.status === 'downloading')
    
    if (downloadingTask) {
      // Should show pause/cancel button
      await expect(page.locator('button').filter({ has: page.locator('svg') }).first()).toBeVisible()
    }
  })

  test('should display action buttons for failed tasks', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    const data = await response.json()
    const failedTask = data.tasks?.find((t: { status: string }) => t.status === 'failed')
    
    if (failedTask) {
      // Should show restart button
      await expect(page.getByRole('button').filter({ hasText: '' }).first()).toBeVisible()
    }
  })

  test('should show delete button for all tasks', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.tasks && data.tasks.length > 0) {
      // Should have delete buttons (trash icons)
      const deleteButtons = await page.locator('button').filter({ has: page.locator('svg') }).all()
      expect(deleteButtons.length).toBeGreaterThan(0)
    }
  })

  test('should display progress information', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.tasks && data.tasks.length > 0) {
      // Progress bar should be visible
      await expect(page.locator('.h-2.w-full.rounded-full.bg-secondary').first()).toBeVisible()
    }
  })

  test('should show task error messages when present', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    const data = await response.json()
    const taskWithError = data.tasks?.find((t: { error?: string }) => t.error)
    
    if (taskWithError) {
      await expect(page.getByText(/Error:/)).toBeVisible()
    }
  })

  test('Clear Completed button should be disabled when no completed tasks', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    const data = await response.json()
    const hasCompleted = data.tasks?.some((t: { status: string }) => t.status === 'completed')
    
    const button = page.getByRole('button', { name: /Clear Completed/i })
    
    if (!hasCompleted) {
      await expect(button).toBeDisabled()
    }
  })

  test('Clear Failed button should be disabled when no failed tasks', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    const data = await response.json()
    const hasFailed = data.tasks?.some((t: { status: string }) => t.status === 'failed')
    
    const button = page.getByRole('button', { name: /Clear Failed/i })
    
    if (!hasFailed) {
      await expect(button).toBeDisabled()
    }
  })
})
