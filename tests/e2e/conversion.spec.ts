import { test, expect } from '@playwright/test';
import { site } from '../../src/data/site';

test.describe('Conversion layer', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
  });

  test('hero shows role, statement and access row', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.hero-role')).toHaveText(site.role);
    await expect(page.locator('.hero-statement')).toContainText('beau');
    await expect(page.locator(`.hero-links a[href="mailto:${site.email}"]`)).toBeVisible();
    await expect(page.locator(`.hero-links a[href="${site.instagram.url}"]`)).toHaveCount(1);
  });

  test('Écrire mailto is present on index, about, collaborer and a roll', async ({ page }) => {
    for (const path of ['/', '/about', '/collaborer', '/roll/matin-brumeux']) {
      await page.goto(path);
      await expect(page.locator(`a[href="mailto:${site.email}"]`).first()).toBeAttached();
    }
  });

  test('collaborer page renders its CTA', async ({ page }) => {
    await page.goto('/collaborer');
    await expect(page.locator('.collab-cta')).toHaveAttribute('href', `mailto:${site.email}`);
  });

  test('no placeholder contact remains', async ({ page }) => {
    await page.goto('/about');
    const html = await page.content();
    expect(html).not.toContain('hello@example.com');
  });
});
