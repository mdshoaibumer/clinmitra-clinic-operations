import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.evaluate(() => { sessionStorage.setItem('_setupComplete', 'true') });
    await page.goto('/login');
  });

  test('should login successfully', async ({ page }) => {
    await page.fill('input[id="username"]', 'admin');
    await page.fill('input[id="password"]', 'password123');
    await page.click('button:has-text("Sign In")');
    await expect(page).toHaveURL(/\/dashboard/);
  });

  test('should fail with wrong credentials', async ({ page }) => {
    await page.fill('input[id="username"]', 'admin');
    await page.fill('input[id="password"]', 'wrongpass');
    await page.click('button:has-text("Sign In")');
    await expect(page.locator('text=Invalid credentials')).toBeVisible();
  });

  test('should handle SQL injection attempt gracefully', async ({ page }) => {
    await page.fill('input[id="username"]', 'sql-inj');
    await page.fill('input[id="password"]', 'anypassword');
    await page.click('button:has-text("Sign In")');
    // Be specific to the error div to avoid strict mode violation
    await expect(page.locator('div.text-red-600').filter({ hasText: 'SQL Exception' })).toBeVisible();
  });
});
