import { test, expect } from '@playwright/test';

test.describe('Setup Wizard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should complete setup successfully', async ({ page }) => {
    // Step 1: Clinic Details
    await page.fill('input[id="clinicName"]', 'Clinmitra Test Clinic');
    await page.fill('input[id="doctorName"]', 'Dr. Smith');
    await page.fill('input[id="phone"]', '9876543210');
    await page.click('button:has-text("Next")');

    // Step 2: Admin Account
    await expect(page.locator('text=Step 2 of 3')).toBeVisible();
    await page.fill('input[id="adminFullName"]', 'Test Admin');
    await page.fill('input[id="adminUsername"]', 'admin');
    await page.fill('input[id="adminPassword"]', 'password123');
    await page.click('button:has-text("Next")');

    // Step 3: Confirm
    await expect(page.locator('text=Step 3 of 3')).toBeVisible();
    await page.click('button:has-text("Complete Setup")');

    // Should redirect to login
    await expect(page).toHaveURL(/\/login/);
  });

  test('should show validation errors for invalid phone', async ({ page }) => {
    await page.fill('input[id="clinicName"]', 'Clinic');
    await page.fill('input[id="doctorName"]', 'Doctor');
    await page.fill('input[id="phone"]', '123'); // Invalid
    await page.click('button:has-text("Next")');
    
    // Check for error message (assuming the schema triggers it)
    await expect(page.locator('text=Valid phone number required')).toBeVisible();
  });

  test('should handle backend errors during setup', async ({ page }) => {
    // Fill step 1 with a phone that triggers mock error
    await page.fill('input[id="clinicName"]', 'Clinic');
    await page.fill('input[id="doctorName"]', 'Doctor');
    await page.fill('input[id="phone"]', '9999999999'); // Triggers error in mock
    await page.click('button:has-text("Next")');

    // Fill step 2
    await page.fill('input[id="adminFullName"]', 'Admin');
    await page.fill('input[id="adminUsername"]', 'admin');
    await page.fill('input[id="adminPassword"]', 'password123');
    await page.click('button:has-text("Next")');

    // Finish
    await page.click('button:has-text("Complete Setup")');

    // Error from mock should appear
    await expect(page.locator('text=Database Error: Unique constraint failed')).toBeVisible({ timeout: 10000 });
  });
});
