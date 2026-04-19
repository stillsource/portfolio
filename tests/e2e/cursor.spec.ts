import { test, expect } from '@playwright/test';

test.describe('Custom cursor', () => {
  test.beforeEach(async ({ page }) => {
    // Skip the splash screen
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
  });

  test('is invisible before the first mouse movement', async ({ page }) => {
    await page.goto('/');
    const cursor = page.locator('#custom-cursor');
    await expect(cursor).toHaveCSS('opacity', '0');
  });

  test('appears and follows the mouse on the index page', async ({ page }) => {
    await page.goto('/');
    const cursor = page.locator('#custom-cursor');

    await page.mouse.move(400, 300);
    await page.waitForTimeout(100);

    await expect(cursor).toHaveCSS('opacity', '1');
    const transform = await cursor.evaluate(el => el.style.transform);
    expect(transform).toContain('translate3d');
  });

  test('stays visible and follows the mouse after navigating to a roll', async ({ page }) => {
    await page.goto('/');
    await page.mouse.move(400, 300);
    await page.waitForTimeout(100);

    // Navigate to the first available roll
    const firstRollLink = page.locator('a[href^="/roll/"]').first();
    await firstRollLink.click();
    await page.waitForLoadState('networkidle');

    // Move the mouse on the new page — wait for the lerp to converge
    await page.mouse.move(600, 400);
    await page.waitForTimeout(500);

    const cursor = page.locator('#custom-cursor');
    await expect(cursor).toHaveCSS('opacity', '1');
    const transform = await cursor.evaluate(el => el.style.transform);
    expect(transform).toContain('translate3d');
  });

  test('does not disappear after a mouseleave then a navigation', async ({ page }) => {
    await page.goto('/');
    await page.mouse.move(400, 300);
    await page.waitForTimeout(100);

    // Simulate a mouseleave (exiting the window)
    await page.mouse.move(-10, -10);
    await page.waitForTimeout(50);

    const cursor = page.locator('#custom-cursor');
    await expect(cursor).toHaveCSS('opacity', '0');

    // Navigate
    const firstRollLink = page.locator('a[href^="/roll/"]').first();
    await firstRollLink.click();
    await page.waitForLoadState('networkidle');

    // Re-enter the window. After a navigation the cursor listeners are
    // re-attached, so allow enough time for both the lerp to converge and the
    // opacity transition (0.18s) to settle before asserting.
    await page.mouse.move(500, 500);
    await expect(cursor).toHaveCSS('opacity', '1', { timeout: 3000 });
  });
});

test.describe('Custom cursor — interactive states', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
  });

  test('gets the is-clickable class when hovering a link', async ({ page }) => {
    await page.goto('/');
    const link = page.locator('a[href^="/roll/"]').first();
    await link.waitFor({ state: 'visible' });
    await link.hover();
    await expect(page.locator('#custom-cursor')).toHaveClass(/is-clickable/);
  });

  test('drops the is-clickable class when leaving a link', async ({ page }) => {
    await page.goto('/');
    const link = page.locator('a[href^="/roll/"]').first();
    await link.waitFor({ state: 'visible' });
    await link.hover();
    await expect(page.locator('#custom-cursor')).toHaveClass(/is-clickable/);
    // Move to an empty area away from links.
    await page.mouse.move(2, 2);
    await expect(page.locator('#custom-cursor')).not.toHaveClass(/is-clickable/);
  });

  test('gets the is-pressed class on pointerdown', async ({ page }) => {
    await page.goto('/');
    await page.mouse.move(400, 300);
    await page.mouse.down();
    await expect(page.locator('#custom-cursor')).toHaveClass(/is-pressed/);
    await page.mouse.up();
    await expect(page.locator('#custom-cursor')).not.toHaveClass(/is-pressed/);
  });
});
