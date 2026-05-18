import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './helpers';

test.describe('Critical Path: Clinic Workflow', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
  });

  test('should complete end-to-end clinical workflow', async ({ page }) => {
    // 1. Create Patient
    await page.click('nav >> text=Patients');
    await page.waitForURL('**/patients');
    await page.click('button:has-text("New Patient")');
    
    // Fill Patient Form
    const uniquePhone = `9${Math.floor(Math.random() * 1000000000).toString().padStart(9, '0')}`;
    await page.fill('input[name="name"]', 'John Doe E2E');
    await page.fill('input[name="phone"]', uniquePhone);
    await page.selectOption('select[name="gender"]', 'male');
    await page.fill('input[name="age"]', '30');
    await page.click('button:has-text("Save Patient")');
    
    // Verify toast
    await expect(page.getByText('Patient Created', { exact: true }).first()).toBeVisible();

    // 2. Book Appointment
    await page.click('nav >> text=Appointments');
    await page.waitForURL('**/appointments');
    await page.click('button:has-text("New Appointment")');
    
    // Fill Appointment Form
    // Wait for patients to load in the select
    await page.waitForSelector(`select[name="patientId"] option:has-text("John Doe E2E")`, { state: 'attached' });
    const selectLocator = page.locator('select[name="patientId"]');
    await selectLocator.selectOption({ label: `John Doe E2E (${uniquePhone})` });
    
    // Set Time
    await page.fill('input[name="startTime"]', '10:00');
    await page.fill('input[name="endTime"]', '11:00');
    await page.fill('input[name="purpose"]', 'Routine Checkup');
    await page.click('button:has-text("Book")');
    
    // Verify toast
    await expect(page.getByText('Appointment booked', { exact: true }).first()).toBeVisible();

    // 3. Generate Invoice
    await page.click('nav >> text=Billing');
    await page.waitForURL('**/billing');
    await page.click('button:has-text("New Invoice")');
    
    // Select Patient
    const billingSelectLocator = page.locator('select').first(); // First select is patient select
    await billingSelectLocator.selectOption({ label: `John Doe E2E (${uniquePhone})` });
    
    // Add Item (already has one empty item)
    await page.fill('input[placeholder="Description"]', 'Consultation Fee');
    await page.fill('input[placeholder="₹ Price"]', '500');
    
    await page.click('button:has-text("Create Invoice")');
    
    // Confirm dialog
    await page.click('button:has-text("Confirm & Create")');
    
    // Wait to be redirected to invoice detail page
    await page.waitForURL(/\/billing\/[a-zA-Z0-9-]+/);
    
    // Verify on Invoice Detail Page (mock returns hardcoded invoice data)
    await expect(page.locator('h1:has-text("Invoice")')).toBeVisible();
    await expect(page.locator('text=Record Payment')).toBeVisible();
  });
});
