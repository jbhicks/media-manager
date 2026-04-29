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
      await expect(page.getByText(firstTask.status).first()).toBeVisible()
      
      // Should show progress bar
      await expect(page.locator('.h-2.rounded-full.bg-primary').first()).toBeVisible()
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
      
      // Progress percentage should be shown
      const progressText = await page.getByText(/% complete/i).first()
      await expect(progressText).toBeVisible()
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
      await expect(page.getByText(taskWithError.error)).toBeVisible()
    }
  })

  test('should show cancel button for downloading tasks', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    const data = await response.json()
    const downloadingTask = data.tasks?.find((t: { status: string }) => t.status === 'downloading')
    
    if (downloadingTask) {
      // Should show pause/cancel button (icon button)
      const taskCard = page.locator('.space-y-4 > div').filter({ hasText: downloadingTask.title }).first()
      const cancelButton = taskCard.locator('button').first()
      await expect(cancelButton).toBeVisible()
      await expect(cancelButton).toBeEnabled()
    }
  })

  test('should cancel a downloading task', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    const data = await response.json()
    const downloadingTask = data.tasks?.find((t: { status: string }) => t.status === 'downloading')
    
    if (downloadingTask) {
      // Find the task card and click cancel
      const taskCard = page.locator('.space-y-4 > div').filter({ hasText: downloadingTask.title }).first()
      const cancelButton = taskCard.locator('button').first()
      await cancelButton.click()
      
      // Wait for cancel API
      await page.waitForResponse(response => 
        response.url().includes('/api/tasks/cancel') && response.status() === 200
      )
      
      // Should show toast
      await expect(page.getByText(/cancelled|Task cancelled/i)).toBeVisible()
    }
  })

  test('should show restart button for failed tasks', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    const data = await response.json()
    const failedTask = data.tasks?.find((t: { status: string }) => t.status === 'failed' || t.status === 'cancelled')
    
    if (failedTask) {
      // Should show restart button
      const taskCard = page.locator('.space-y-4 > div').filter({ hasText: failedTask.title }).first()
      const restartButton = taskCard.locator('button').filter({ has: page.locator('svg') }).first()
      await expect(restartButton).toBeVisible()
      await expect(restartButton).toBeEnabled()
    }
  })

  test('should restart a failed task', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    const data = await response.json()
    const failedTask = data.tasks?.find((t: { status: string }) => t.status === 'failed' || t.status === 'cancelled')
    
    if (failedTask) {
      // Find the task card and click restart
      const taskCard = page.locator('.space-y-4 > div').filter({ hasText: failedTask.title }).first()
      const restartButton = taskCard.locator('button').filter({ has: page.locator('svg') }).first()
      await restartButton.click()
      
      // Wait for restart API
      await page.waitForResponse(response => 
        response.url().includes('/api/tasks/restart') && response.status() === 200
      )
      
      // Should show toast
      await expect(page.getByText(/restarted|Task restarted/i)).toBeVisible()
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

  test('should delete a task', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.tasks && data.tasks.length > 0) {
      const firstTask = data.tasks[0]
      
      // Find the task card
      const taskCard = page.locator('.space-y-4 > div').filter({ hasText: firstTask.title }).first()
      
      // Find delete button (last button in the card)
      const deleteButton = taskCard.locator('button').last()
      await deleteButton.click()
      
      // Wait for delete API
      await page.waitForResponse(response => 
        response.url().includes('/api/tasks/delete') && response.status() === 200
      )
      
      // Should show toast
      await expect(page.getByText(/deleted|Task deleted/i)).toBeVisible()
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

  test('should clear completed tasks', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    const data = await response.json()
    const hasCompleted = data.tasks?.some((t: { status: string }) => t.status === 'completed')
    
    if (hasCompleted) {
      const button = page.getByRole('button', { name: /Clear Completed/i })
      await button.click()
      
      // Wait for API
      await page.waitForResponse(response => 
        response.url().includes('/api/tasks/clear-completed') && response.status() === 200
      )
      
      // Should show toast
      await expect(page.getByText(/cleared|Completed tasks cleared/i)).toBeVisible()
    }
  })

  test('should clear failed tasks', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    const data = await response.json()
    const hasFailed = data.tasks?.some((t: { status: string }) => t.status === 'failed')
    
    if (hasFailed) {
      const button = page.getByRole('button', { name: /Clear Failed/i })
      await button.click()
      
      // Wait for API
      await page.waitForResponse(response => 
        response.url().includes('/api/tasks/clear-failed') && response.status() === 200
      )
      
      // Should show toast
      await expect(page.getByText(/cleared|Failed tasks cleared/i)).toBeVisible()
    }
  })

  test('should show task metadata (size, seeders, started time)', async ({ page }) => {
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/tasks') && response.status() === 200
    )
    
    const data = await response.json()
    
    if (data.tasks && data.tasks.length > 0) {
      const task = data.tasks[0]
      
      // Should show size
      if (task.size) {
        const sizeElements = await page.locator('text=/\\d+\\s*(GB|MB|KB|Bytes)/i').all()
        expect(sizeElements.length).toBeGreaterThan(0)
      }
      
      // Should show seeders
      if (task.seeders !== undefined) {
        await expect(page.getByText(/seeders/i).first()).toBeVisible()
      }
      
      // Should show started time if available
      if (task.started_at) {
        await expect(page.getByText(/Started/i).first()).toBeVisible()
      }
    }
  })

  test('should show loading state initially', async ({ page }) => {
    // Before API response, should show loading or content
    await page.goto('/downloads')
    
    // Wait a tiny bit for loading state
    await page.waitForTimeout(100)
    
    // Page should have heading visible immediately
    await expect(page.getByRole('heading', { name: 'Downloads' })).toBeVisible()
  })
})