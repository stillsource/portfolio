import { test, expect } from '@playwright/test';

// Dispatched through JS so the test stays robust against overlay regressions
// (cursor-root mix-blend, noise-overlay, future z-index shifts, etc.). A
// separate test below exercises the native pointer path end to end.
const clickToggle = (page: import('@playwright/test').Page) =>
  page.evaluate(() => (document.getElementById('search-toggle') as HTMLButtonElement)?.click());

const clickClose = (page: import('@playwright/test').Page) =>
  page.evaluate(() => (document.getElementById('search-close') as HTMLButtonElement)?.click());

test.describe('Search', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
    await page.goto('/');
    // Wait for astro:page-load to have triggered setupSearch()
    await page.evaluate(() =>
      new Promise<void>(resolve => {
        if ((document.getElementById('search-results') as HTMLElement & { __searchBound?: boolean })?.__searchBound) {
          resolve();
        } else {
          document.addEventListener('astro:page-load', () => resolve(), { once: true });
        }
      })
    );
  });

  test('search button opens the modal', async ({ page }) => {
    await clickToggle(page);
    await expect(page.locator('#search-overlay')).toHaveClass(/is-open/, { timeout: 2000 });
    // Input visible in the opened modal
    await expect(page.locator('#search-input')).toBeVisible({ timeout: 2000 });
  });

  test('close button hides the modal', async ({ page }) => {
    await clickToggle(page);
    await expect(page.locator('#search-overlay')).toHaveClass(/is-open/, { timeout: 2000 });

    await clickClose(page);
    await expect(page.locator('#search-overlay')).not.toHaveClass(/is-open/, { timeout: 2000 });
  });

  test('Escape closes the modal', async ({ page }) => {
    await clickToggle(page);
    await expect(page.locator('#search-overlay')).toHaveClass(/is-open/, { timeout: 2000 });

    await page.keyboard.press('Escape');
    await expect(page.locator('#search-overlay')).not.toHaveClass(/is-open/, { timeout: 2000 });
  });

  test('typing in the field displays results', async ({ page }) => {
    await clickToggle(page);
    await expect(page.locator('#search-overlay')).toHaveClass(/is-open/, { timeout: 2000 });

    await page.locator('#search-input').fill('Tokyo');
    await page.waitForTimeout(400);

    await expect(page.locator('#search-results')).not.toContainText('Commencez à taper');
  });

  test('searching "Tokyo" returns at least 1 result', async ({ page }) => {
    await clickToggle(page);
    await expect(page.locator('#search-overlay')).toHaveClass(/is-open/, { timeout: 2000 });
    await page.locator('#search-input').fill('Tokyo');
    await page.waitForTimeout(500);

    const results = page.locator('#search-results .search-result-item');
    await expect(results.first()).toBeVisible({ timeout: 3000 });
  });

  test('search with no result shows "Aucun fragment" message', async ({ page }) => {
    await clickToggle(page);
    await expect(page.locator('#search-overlay')).toHaveClass(/is-open/, { timeout: 2000 });
    await page.locator('#search-input').fill('xyz123nonexistent');
    await page.waitForTimeout(500);

    await expect(page.locator('#search-results')).toContainText('Aucun fragment');
  });

  test('empty search re-shows placeholder', async ({ page }) => {
    await clickToggle(page);
    await page.locator('#search-input').fill('Tokyo');
    await page.waitForTimeout(300);
    await page.locator('#search-input').clear();
    await page.waitForTimeout(300);

    await expect(page.locator('#search-results')).toContainText('Commencez à taper');
  });

  test('accessible modal: role dialog and aria-modal', async ({ page }) => {
    const overlay = page.locator('#search-overlay');
    await expect(overlay).toHaveAttribute('role', 'dialog');
    await expect(overlay).toHaveAttribute('aria-modal', 'true');
  });

  test('native pointer click on the search button opens the modal without errors', async ({ page }) => {
    const consoleErrors: string[] = [];
    page.on('pageerror', (err) => consoleErrors.push(err.message));
    page.on('console', (msg) => {
      if (msg.type() === 'error') consoleErrors.push(msg.text());
    });

    await page.locator('#search-toggle').click();

    await expect(page.locator('#search-overlay')).toHaveClass(/is-open/, { timeout: 2000 });
    await expect(page.locator('#search-input')).toBeVisible({ timeout: 2000 });
    expect(consoleErrors, `console errors after search toggle: ${consoleErrors.join(' | ')}`).toEqual([]);
  });
});
