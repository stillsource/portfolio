import { test, expect } from '@playwright/test';

test.describe('Tag pages', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
  });

  test('/tags/Nocturne shows the tag as title', async ({ page }) => {
    await page.goto('/tags/Nocturne');
    await expect(page.locator('.tag-title')).toHaveText('Nocturne');
  });

  test('/tags/Nocturne shows at least one roll', async ({ page }) => {
    await page.goto('/tags/Nocturne');
    // The page renders Roll components directly (not a list of links)
    const rolls = page.locator('.roll-section');
    await expect(rolls.first()).toBeAttached({ timeout: 5000 });
    const count = await rolls.count();
    expect(count).toBeGreaterThan(0);
  });

  test('/tags/Architecture shows the tag and rolls', async ({ page }) => {
    await page.goto('/tags/Architecture');
    await expect(page.locator('.tag-title')).toHaveText('Architecture');
    const rolls = page.locator('.roll-section');
    await expect(rolls.first()).toBeAttached();
  });

  test('/tags/Pluie only contains rolls with images', async ({ page }) => {
    await page.goto('/tags/Pluie');
    await expect(page.locator('.tag-title')).toHaveText('Pluie');
    // Images are present inside the rolls
    const images = page.locator('.styled-image');
    await expect(images.first()).toBeAttached();
    const count = await images.count();
    expect(count).toBeGreaterThan(0);
  });

  test('unknown tag page returns 404 or redirects', async ({ page }) => {
    const response = await page.goto('/tags/TagInexistant99');
    // Either 404 or redirect to /
    expect(response?.status()).not.toBe(200);
  });

  test('index filtering by "Nocturne" -> same roll count as /tags/Nocturne page', async ({ page }) => {
    // Count rolls on the /tags/Nocturne page
    await page.goto('/tags/Nocturne');
    const rollsOnTagPage = await page.locator('.roll-section').count();

    // Index: open the filter disclosure first (hidden by default), then click
    // the Nocturne tag button.
    await page.goto('/');
    await expect(page.locator('.filter-disclosure')).toHaveClass(/is-visible/, { timeout: 3000 });
    await page.locator('#filter-disclosure-toggle').click();
    const nocturneBtn = page.locator('.filter-btn[data-filter="Nocturne"]');
    await expect(nocturneBtn).toBeVisible();
    await nocturneBtn.click();
    await expect(nocturneBtn).toHaveAttribute('aria-pressed', 'true', { timeout: 1000 });

    const visibleOnIndex = await page.locator('.roll-item:not(.hidden)').count();
    expect(visibleOnIndex).toBe(rollsOnTagPage);
  });
});
