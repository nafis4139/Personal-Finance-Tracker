import { test, expect } from '@playwright/test';

const BASE = process.env.E2E_BASE ?? 'http://localhost:8080';

test.beforeEach(async ({ context }) => {
  // start clean: no cookies, no localStorage JWT
  await context.clearCookies();
});

test('login page renders', async ({ page }) => {
  await page.goto(`${BASE}/login`);

  // Clear localStorage after navigation to ensure app doesn’t instantly redirect
  await page.evaluate(() => localStorage.clear());

  // Accept either the hero text or the Login header, depending on markup
  const loginHeading = page.getByRole('heading', { name: /login|welcome back/i });
  await expect(loginHeading).toBeVisible();
});

test('register -> login flow (happy path)', async ({ page }) => {
  const email = `e2e+${Date.now()}@example.com`;

  // Go to Register
  await page.goto(`${BASE}/register`);
  await page.evaluate(() => localStorage.clear());

  await page.getByPlaceholder(/name/i).fill('Test User');
  await page.getByPlaceholder(/email/i).fill(email);
  await page.getByPlaceholder(/password/i).fill('secret123');
  await page.getByRole('button', { name: /register/i }).click();

  // After registering, apps either:
  //  A) redirect to /login, or
  //  B) auto-login and redirect to /dashboard
  // Wait for either outcome.
  await page.waitForLoadState('networkidle');
  const url = page.url();

  if (/\/dashboard/.test(url)) {
    // Already logged in
  } else {
    // Not logged in yet → go to login (some UIs stay on /register with a link)
    if (!/\/login/.test(url)) {
      // click a "Login" link if present; otherwise navigate directly
      const maybeLoginLink = page.getByRole('link', { name: /login/i });
      if (await maybeLoginLink.count()) {
        await maybeLoginLink.first().click();
      } else {
        await page.goto(`${BASE}/login`);
      }
    }

    await page.getByPlaceholder(/email/i).fill(email);
    await page.getByPlaceholder(/password/i).fill('secret123');
    await page.getByRole('button', { name: /log in|login/i }).click();
  }

  // Arrive at dashboard
  await expect(page).toHaveURL(/\/dashboard/);
  // Use a unique, unambiguous element: the main Dashboard heading
  await expect(page.getByRole('heading', { name: /dashboard/i })).toBeVisible();
});
