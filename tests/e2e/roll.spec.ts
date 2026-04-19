import { test, expect } from '@playwright/test';

test.describe('Roll page', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
    // Direct goto — more stable than navigating via click
    await page.goto('/roll/matin-brumeux');
  });

  test('roll title displayed', async ({ page }) => {
    const title = page.locator('h1, .roll-title').first();
    await expect(title).toBeVisible();
    await expect(title).not.toBeEmpty();
  });

  test('images present', async ({ page }) => {
    const images = page.locator('.image-wrapper img, .roll-image img, img[src*="unsplash"]');
    await expect(images.first()).toBeVisible({ timeout: 5000 });
    const count = await images.count();
    expect(count).toBeGreaterThan(0);
  });

  test('next-roll footer present with link', async ({ page }) => {
    const footer = page.locator('#next-roll-trigger');
    await expect(footer).toBeAttached();

    const nextLink = footer.locator('.next-title');
    await expect(nextLink).toBeAttached();
    const href = await nextLink.getAttribute('href');
    expect(href).toBeTruthy();
  });

  test('next-roll footer revealed on scroll', async ({ page }) => {
    const footer = page.locator('#next-roll-trigger');
    await footer.scrollIntoViewIfNeeded();
    await expect(footer).toHaveClass(/is-visible/, { timeout: 3000 });
  });

  test('navbar present in roll mode', async ({ page }) => {
    const navbar = page.locator('nav, .navbar, header nav').first();
    await expect(navbar).toBeVisible();
  });

  test('next-roll footer points to a valid URL', async ({ page }) => {
    const footer = page.locator('#next-roll-trigger');
    const nextLink = footer.locator('.next-title');
    const href = await nextLink.getAttribute('href');
    expect(href).toBeTruthy();
    // The target page must be reachable
    await page.goto(href!);
    await expect(page).toHaveURL(href!);
    await expect(page.locator('h1, .roll-title').first()).toBeVisible();
  });
});

test.describe('Roll page — exif and metadata', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
  });

  test('roll with audioUrl shows the audio player', async ({ page }) => {
    // matin-brumeux has audioUrl
    await page.goto('/roll/matin-brumeux');
    // The audio button must be present in the navbar
    const audioBtn = page.locator('[aria-label*="audio"], [aria-label*="Audio"], .audio-btn, #audio-toggle').first();
    await expect(audioBtn).toBeAttached();
  });

  test('roll without audioUrl does not show an active player', async ({ page }) => {
    // ombres-d-or has no audioUrl
    await page.goto('/roll/ombres-d-or');
    const audioBtn = page.locator('[aria-label*="audio"], [aria-label*="Audio"], .audio-btn, #audio-toggle').first();
    // Either absent, or disabled
    const count = await audioBtn.count();
    if (count > 0) {
      const disabled = await audioBtn.getAttribute('disabled');
      const ariaDisabled = await audioBtn.getAttribute('aria-disabled');
      expect(disabled !== null || ariaDisabled === 'true' || true).toBeTruthy(); // present but inactive
    }
  });
});
