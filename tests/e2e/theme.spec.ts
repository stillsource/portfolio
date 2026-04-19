import { test, expect } from '@playwright/test';

/**
 * Theme toggle (dark/light) — anchored on the Navbar button #theme-toggle.
 *
 * - Initial theme is read from `prefers-color-scheme` via an anti-FOUC inline
 *   script in Layout.astro, or from localStorage["theme-preference"] if set.
 * - Clicking the button flips html[data-theme] between dark and light.
 * - The choice is persisted in localStorage under "theme-preference".
 * - Reloading the page honours the persisted preference.
 */
test.describe('Theme toggle', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
  });

  test('initial theme matches prefers-color-scheme when no preference is stored', async ({ page, browserName }, testInfo) => {
    // Force the emulated color scheme to light before navigating.
    await page.emulateMedia({ colorScheme: 'light' });
    // Ensure no stored preference exists by clearing localStorage at init.
    await page.addInitScript(() => localStorage.removeItem('theme-preference'));
    await page.goto('/');

    const theme = await page.locator('html').getAttribute('data-theme');
    expect(theme).toBe('light');

    // Now emulate dark and reload — with no stored preference the system
    // default wins again.
    await page.emulateMedia({ colorScheme: 'dark' });
    await page.addInitScript(() => localStorage.removeItem('theme-preference'));
    await page.goto('/');
    const theme2 = await page.locator('html').getAttribute('data-theme');
    expect(theme2).toBe('dark');
    // silence linter for unused testInfo / browserName parameters
    void browserName; void testInfo;
  });

  test('clicking the toggle flips the theme and persists the choice', async ({ page }) => {
    await page.emulateMedia({ colorScheme: 'dark' });
    await page.addInitScript(() => localStorage.removeItem('theme-preference'));
    // The toggle lives in the Navbar which is only visible on non-index
    // pages immediately (index waits on splashFinished). Use a roll page
    // to avoid splash timing.
    await page.goto('/roll/matin-brumeux');
    await page.waitForLoadState('networkidle');

    const html = page.locator('html');
    await expect(html).toHaveAttribute('data-theme', 'dark');

    const toggle = page.locator('#theme-toggle');
    await expect(toggle).toBeVisible();
    // Initial aria-label announces what pressing the button will do.
    await expect(toggle).toHaveAttribute('aria-label', /mode clair/);

    await toggle.click();
    await expect(html).toHaveAttribute('data-theme', 'light');
    await expect(toggle).toHaveAttribute('aria-pressed', 'true');
    await expect(toggle).toHaveAttribute('aria-label', /mode sombre/);

    const stored = await page.evaluate(() => localStorage.getItem('theme-preference'));
    expect(stored).toBe('light');

    // Flip back.
    await toggle.click();
    await expect(html).toHaveAttribute('data-theme', 'dark');
    const storedBack = await page.evaluate(() => localStorage.getItem('theme-preference'));
    expect(storedBack).toBe('dark');
  });

  test('stored preference survives a full reload', async ({ page }) => {
    await page.emulateMedia({ colorScheme: 'dark' });
    await page.addInitScript(() => localStorage.setItem('theme-preference', 'light'));
    await page.goto('/roll/matin-brumeux');
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');

    await page.reload();
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');
  });

  test('one click still flips exactly once after several client navigations', async ({ page }) => {
    // Regression for the navbar's `transition:persist` script re-binding on
    // every `astro:page-load`. Each stale listener turned a single click
    // into N rapid flips, which netted out to "nothing changed".
    await page.addInitScript(() => localStorage.setItem('theme-preference', 'dark'));

    await page.goto('/roll/matin-brumeux');
    await page.waitForLoadState('networkidle');

    // Hop across a few pages using the in-app Astro router so the navbar's
    // persisted script listens through several `astro:page-load` events.
    await page.locator('a[href="/"]').first().click();
    await page.waitForLoadState('networkidle');
    await page.locator('a[href^="/roll/"]').first().click();
    await page.waitForLoadState('networkidle');
    await page.locator('a[href="/about"]').first().click();
    await page.waitForLoadState('networkidle');
    await page.locator('a[href="/"]').first().click();
    await page.waitForLoadState('networkidle');

    // Normalise to a known start (ClientRouter's head merge can temporarily
    // drop the client-set `data-theme` attribute; the toggle recovers from
    // `null` by defaulting to 'dark', but that makes the expected delta
    // ambiguous for the test).
    const html = page.locator('html');
    await page.evaluate(() => document.documentElement.setAttribute('data-theme', 'dark'));

    await page.locator('#theme-toggle').click();
    await expect(html).toHaveAttribute('data-theme', 'light', { timeout: 1000 });

    // If the listener had been stacked by the persisted navbar, a surplus
    // would have flipped the attribute again right after. Hold still for a
    // beat and re-assert.
    await page.waitForTimeout(400);
    await expect(html).toHaveAttribute('data-theme', 'light');
    const stored = await page.evaluate(() => localStorage.getItem('theme-preference'));
    expect(stored).toBe('light');
  });
});
