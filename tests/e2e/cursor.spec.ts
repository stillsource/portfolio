import { test, expect } from '@playwright/test';

test.describe('Custom cursor', () => {
  test.beforeEach(async ({ page }) => {
    // Passer le splash screen
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
  });

  test('est invisible avant le premier mouvement de souris', async ({ page }) => {
    await page.goto('/');
    const cursor = page.locator('#custom-cursor');
    await expect(cursor).toHaveCSS('opacity', '0');
  });

  test('apparaît et suit la souris sur la page index', async ({ page }) => {
    await page.goto('/');
    const cursor = page.locator('#custom-cursor');

    await page.mouse.move(400, 300);
    await page.waitForTimeout(100);

    await expect(cursor).toHaveCSS('opacity', '1');
    const transform = await cursor.evaluate(el => el.style.transform);
    expect(transform).toContain('translate3d');
  });

  test('reste visible et suit la souris après navigation vers un roll', async ({ page }) => {
    await page.goto('/');
    await page.mouse.move(400, 300);
    await page.waitForTimeout(100);

    // Naviguer vers le premier roll disponible
    const firstRollLink = page.locator('a[href^="/roll/"]').first();
    await firstRollLink.click();
    await page.waitForLoadState('networkidle');

    // Bouger la souris sur la nouvelle page
    await page.mouse.move(600, 400);
    await page.waitForTimeout(150);

    const cursor = page.locator('#custom-cursor');
    await expect(cursor).toHaveCSS('opacity', '1');
    const transform = await cursor.evaluate(el => el.style.transform);
    expect(transform).toContain('translate3d');
    // Le curseur doit être positionné là où est la souris (tolérance de 20px)
    const match = transform.match(/translate3d\(([^,]+)px,\s*([^,]+)px/);
    expect(match).not.toBeNull();
    if (match) {
      expect(Math.abs(parseFloat(match[1]) - 600)).toBeLessThan(20);
      expect(Math.abs(parseFloat(match[2]) - 400)).toBeLessThan(20);
    }
  });

  test('ne disparaît pas après un mouseleave puis une navigation', async ({ page }) => {
    await page.goto('/');
    await page.mouse.move(400, 300);
    await page.waitForTimeout(100);

    // Simuler un mouseleave (sortie de fenêtre)
    await page.mouse.move(-10, -10);
    await page.waitForTimeout(50);

    const cursor = page.locator('#custom-cursor');
    await expect(cursor).toHaveCSS('opacity', '0');

    // Naviguer
    const firstRollLink = page.locator('a[href^="/roll/"]').first();
    await firstRollLink.click();
    await page.waitForLoadState('networkidle');

    // Re-entrer dans la fenêtre
    await page.mouse.move(500, 500);
    await page.waitForTimeout(150);

    await expect(cursor).toHaveCSS('opacity', '1');
  });
});
