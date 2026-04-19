import { test, expect } from '@playwright/test';

/**
 * "Chambre noire" immersive reading mode.
 *
 * Pressing `h` (or `.`) anywhere on the page toggles `body.chambre-noire`,
 * which fades out the UI chrome (.nav-container, .next-roll-footer, etc.)
 * and fades in the `.dust-canvas` overlay. Pressing the key again reverts
 * the state. The binding lives in LightboxGlobal.astro.
 */
test.describe('Chambre noire mode', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
  });

  test('pressing "h" toggles body.chambre-noire and reveals .dust-canvas', async ({ page }) => {
    await page.goto('/roll/matin-brumeux');
    await page.waitForLoadState('networkidle');
    // The keydown handler is installed by LightboxGlobal.astro only once per
    // window (guarded by window.__lightbox_initialized). Wait for that flag
    // before dispatching keys.
    await page.waitForFunction(
      () => (window as unknown as { __lightbox_initialized?: boolean }).__lightbox_initialized === true,
      null,
      { timeout: 5000 },
    );

    const body = page.locator('body');
    const dust = page.locator('.dust-canvas');

    // Initial state — chambre-noire is off, dust canvas fully transparent.
    await expect(body).not.toHaveClass(/chambre-noire/);
    const initialOpacity = await dust.evaluate(el => getComputedStyle(el).opacity);
    expect(Number(initialOpacity)).toBeLessThanOrEqual(0.01);

    // First press — chambre-noire ON.
    await page.keyboard.press('h');
    await expect(body).toHaveClass(/chambre-noire/);
    // The CSS transitions opacity over ~2s; poll until it reaches 1.
    await page.waitForFunction(
      () => Number(getComputedStyle(document.querySelector('.dust-canvas')!).opacity) > 0.9,
      null,
      { timeout: 3000 },
    );

    // Second press — back to normal.
    await page.keyboard.press('h');
    await expect(body).not.toHaveClass(/chambre-noire/);
    await page.waitForFunction(
      () => Number(getComputedStyle(document.querySelector('.dust-canvas')!).opacity) < 0.1,
      null,
      { timeout: 3000 },
    );
  });

  test('.next-roll-footer fades out when chambre-noire engages', async ({ page }) => {
    await page.goto('/roll/matin-brumeux');
    await page.waitForLoadState('networkidle');
    // The keydown handler is installed by LightboxGlobal.astro only once per
    // window (guarded by window.__lightbox_initialized). Wait for that flag
    // before dispatching keys.
    await page.waitForFunction(
      () => (window as unknown as { __lightbox_initialized?: boolean }).__lightbox_initialized === true,
      null,
      { timeout: 5000 },
    );

    // Scroll the footer into view and wait for the reveal-on-scroll observer
    // to flip its opacity from 0 to 1.
    const footer = page.locator('.next-roll-footer');
    await footer.scrollIntoViewIfNeeded();
    await page.waitForFunction(
      () => Number(getComputedStyle(document.querySelector('.next-roll-footer')!).opacity) > 0.9,
      null,
      { timeout: 3000 },
    );

    const before = Number(await footer.evaluate(el => getComputedStyle(el).opacity));
    expect(before).toBeGreaterThan(0.5);

    await page.keyboard.press('h');
    // The rule sets `opacity: 0 !important` with a 1s transition.
    await page.waitForFunction(
      () => Number(getComputedStyle(document.querySelector('.next-roll-footer')!).opacity) < 0.05,
      null,
      { timeout: 3000 },
    );

    const after = Number(await footer.evaluate(el => getComputedStyle(el).opacity));
    expect(after).toBeLessThan(before);
  });

  test('the "." key is an alias for "h"', async ({ page }) => {
    await page.goto('/roll/matin-brumeux');
    await page.waitForLoadState('networkidle');
    // The keydown handler is installed by LightboxGlobal.astro only once per
    // window (guarded by window.__lightbox_initialized). Wait for that flag
    // before dispatching keys.
    await page.waitForFunction(
      () => (window as unknown as { __lightbox_initialized?: boolean }).__lightbox_initialized === true,
      null,
      { timeout: 5000 },
    );

    const body = page.locator('body');
    await page.keyboard.press('.');
    await expect(body).toHaveClass(/chambre-noire/);

    await page.keyboard.press('.');
    await expect(body).not.toHaveClass(/chambre-noire/);
  });
});
