import { test, expect } from '@playwright/test';

test.describe('Performance hints on roll pages', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
    await page.goto('/roll/matin-brumeux');
    await page.locator('.styled-image').first().waitFor({ state: 'attached' });
  });

  test('first image is eager and high-priority', async ({ page }) => {
    const firstImg = page.locator('.styled-image').first();
    await expect(firstImg).toHaveAttribute('loading', 'eager');
    await expect(firstImg).toHaveAttribute('fetchpriority', 'high');
  });

  test('subsequent images are lazy', async ({ page }) => {
    const images = page.locator('.styled-image');
    const count = await images.count();
    test.skip(count < 2, 'roll has fewer than 2 images');
    await expect(images.nth(1)).toHaveAttribute('loading', 'lazy');
  });

  test('skeleton shimmer is removed once the image finishes loading', async ({ page }) => {
    const firstContainer = page.locator('.image-container').first();
    await expect(firstContainer).toHaveClass(/loaded/, { timeout: 5000 });
  });
});
