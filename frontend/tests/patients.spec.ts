import { test, expect } from '@playwright/test';

test.describe('Patient Management', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.evaluate(() => { sessionStorage.setItem('_setupComplete', 'true') });

    // Login first
    await page.goto('/login');
    await page.fill('input[id="username"]', 'admin');
    await page.fill('input[id="password"]', 'password123');
    await page.click('button:has-text("Sign In")');
    await expect(page).toHaveURL(/\/dashboard/);
    
    // Navigate to patients
    await page.click('nav >> text=Patients');
    await expect(page).toHaveURL(/\/patients/);
  });

  test('should add and then search for a patient', async ({ page }) => {
    // Add patient
    await page.click('button:has-text("New Patient")');
    const uniqueName = `Jane Doe ${Date.now()}`;
    await page.fill('input[name="name"]', uniqueName);
    await page.fill('input[name="phone"]', '9876543211');
    await page.selectOption('select[name="gender"]', 'female');
    await page.fill('input[name="age"]', '30');
    await page.click('button:has-text("Save Patient")');
    
    // Wait for the form to close
    await expect(page.locator('text=Register New Patient')).not.toBeVisible();

    // Search for the added patient
    const searchInput = page.locator('input[placeholder="Search by name or phone..."]');
    await searchInput.fill(uniqueName);
    
    // Check table content
    await expect(page.locator('table')).toContainText(uniqueName);
  });

  test('should show validation error for missing required fields', async ({ page }) => {
    await page.click('button:has-text("New Patient")');
    await page.click('button:has-text("Save Patient")');

    await expect(page.locator('text=Name must be at least 2 characters')).toBeVisible();
    await expect(page.locator('text=Valid 10-digit phone number required')).toBeVisible();
  });
});
