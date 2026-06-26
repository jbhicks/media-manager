import { test, expect } from '@playwright/test';

/**
 * TMDB API Key Registration Automation
 * This script navigates to TMDB and creates an API key
 */

test('Register for TMDB API key', async ({ page }) => {
  // Step 1: Navigate to TMDB signup
  await page.goto('https://www.themoviedb.org/signup');
  
  // Wait for the page to load
  await page.waitForLoadState('networkidle');
  
  // Check if we're on the signup page
  await expect(page).toHaveTitle(/Join TMDB/);
  
  console.log('On TMDB signup page');
  
  // Step 2: Fill in registration form
  // Note: TMDB requires email verification, so we'll need a temp email
  
  // Take screenshot for debugging
  await page.screenshot({ path: 'tmdb-signup.png' });
  
  // The form fields
  const usernameInput = page.locator('input[name="username"]').or(page.locator('#username'));
  const passwordInput = page.locator('input[name="password"]').or(page.locator('#password'));
  const confirmPasswordInput = page.locator('input[name="password_confirm"]').or(page.locator('#password_confirm'));
  const emailInput = page.locator('input[name="email"]').or(page.locator('#email'));
  
  // Check what fields are available
  const hasUsername = await usernameInput.isVisible().catch(() => false);
  const hasPassword = await passwordInput.isVisible().catch(() => false);
  const hasEmail = await emailInput.isVisible().catch(() => false);
  
  console.log(`Form fields - Username: ${hasUsername}, Password: ${hasPassword}, Email: ${hasEmail}`);
  
  // If we can't find the form, take a screenshot and log the HTML
  if (!hasUsername && !hasEmail) {
    const html = await page.content();
    console.log('Page HTML:', html.substring(0, 1000));
    await page.screenshot({ path: 'tmdb-signup-error.png' });
    throw new Error('Could not find signup form');
  }
});

test('Navigate to TMDB API settings', async ({ page }) => {
  // Navigate to API settings (requires login)
  await page.goto('https://www.themoviedb.org/settings/api');
  
  await page.waitForLoadState('networkidle');
  
  // Check if we're redirected to login
  const currentUrl = page.url();
  console.log('Current URL:', currentUrl);
  
  if (currentUrl.includes('/login')) {
    console.log('Need to login first');
    await page.screenshot({ path: 'tmdb-login.png' });
  } else {
    console.log('On API settings page');
    await page.screenshot({ path: 'tmdb-api-settings.png' });
  }
});
