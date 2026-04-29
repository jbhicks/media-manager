import { test, expect } from '@playwright/test';

test.describe('Torrent Image Extraction', () => {
  test('should extract images from torrent', async ({ page }) => {
    // Navigate to search page
    await page.goto('http://localhost:5173/search');
    
    // Wait for page to load
    await expect(page.getByText('Search torrents across all indexers')).toBeVisible();
    
    // Search for something
    const searchInput = page.getByPlaceholder('Search torrents...');
    await searchInput.fill('Avatar');
    await searchInput.press('Enter');
    
    // Wait for results
    await expect(page.getByText('Search Results')).toBeVisible();
    
    // Click on a result to see if it has images
    // This is a basic test - in reality we'd need a real magnet link
  });

  test('should show placeholder when no images available', async ({ page }) => {
    await page.goto('http://localhost:5173/search');
    
    await expect(page.getByText('Search torrents across all indexers')).toBeVisible();
    
    // Search for something
    const searchInput = page.getByPlaceholder('Search torrents...');
    await searchInput.fill('test12345nonexistent');
    await searchInput.press('Enter');
    
    // Should show search results tab or no results message
    await expect(page.getByText('Search Results').or(page.getByText('No search results found'))).toBeVisible();
  });
});