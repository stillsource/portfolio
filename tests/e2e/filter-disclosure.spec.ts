import { test, expect } from '@playwright/test';

test.describe('Filter disclosure', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
    await page.goto('/');
  });

  test('panel is collapsed by default and announces its state', async ({ page }) => {
    const toggle = page.locator('#filter-disclosure-toggle');
    const panel = page.locator('#filter-section');

    await expect(toggle).toBeVisible();
    await expect(toggle).toHaveAttribute('aria-expanded', 'false');
    await expect(panel).toHaveAttribute('aria-hidden', 'true');
    // Collapsed panel has zero effective height.
    const panelBoxHeight = await panel.evaluate((el) => el.getBoundingClientRect().height);
    expect(panelBoxHeight).toBeLessThan(8);
  });

  test('clicking the toggle opens the panel, clicking again closes it', async ({ page }) => {
    const toggle = page.locator('#filter-disclosure-toggle');
    const panel = page.locator('#filter-section');

    await toggle.click();
    await expect(toggle).toHaveAttribute('aria-expanded', 'true');
    await expect(panel).toHaveAttribute('aria-hidden', 'false');
    await expect(panel).toHaveClass(/is-open/);
    await expect(page.locator('.filter-btn[data-filter="Nocturne"]')).toBeVisible();

    await toggle.click();
    await expect(toggle).toHaveAttribute('aria-expanded', 'false');
    await expect(panel).toHaveAttribute('aria-hidden', 'true');
  });

  test('Escape closes the open panel and restores focus to the toggle', async ({ page }) => {
    const toggle = page.locator('#filter-disclosure-toggle');
    const panel = page.locator('#filter-section');

    await toggle.click();
    await expect(panel).toHaveClass(/is-open/);

    await page.keyboard.press('Escape');
    await expect(panel).not.toHaveClass(/is-open/);
    await expect(toggle).toBeFocused();
  });

  test('rolls remain accessible without opening the disclosure', async ({ page }) => {
    // Direct access to the first roll link without touching the disclosure.
    await expect(page.locator('.roll-item').first()).toHaveClass(/is-visible/, { timeout: 3000 });
    const firstLink = page.locator('.roll-item:not(.hidden) .roll-link').first();
    await expect(firstLink).toBeVisible();
  });
});
