import { test, expect } from '@playwright/test';

test.describe('About page', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
  });

  test('renders the editorial structure', async ({ page }) => {
    await page.goto('/about');

    await expect(page.locator('.masthead')).toBeVisible();
    await expect(page.locator('.masthead .issue')).toContainText('I');
    await expect(page.locator('.kicker')).toContainText('Manifeste');
    await expect(page.locator('h1.title')).toBeVisible();
    await expect(page.locator('.standfirst')).toContainText('Photographie de rue');
    await expect(page.locator('.pull-quote')).toBeVisible();
    await expect(page.locator('.colophon .signature')).toBeVisible();
  });

  test('reveals the body paragraph quickly even after a fast scroll', async ({ page }) => {
    await page.goto('/about');

    // Jump past the hero without waiting for the cascade so the body enters
    // the viewport before its reveal transition naturally completes.
    await page.evaluate(() => window.scrollTo({ top: document.body.scrollHeight * 0.4, behavior: 'instant' as ScrollBehavior }));

    const lede = page.locator('.body .lede');
    await expect(lede).toBeVisible();
    // Reveal transition is 0.65s; allow a bit of buffer for the observer
    // callback + transitionend.
    await expect(lede).toHaveCSS('opacity', '1', { timeout: 1500 });
  });

  test('drop cap on the first paragraph', async ({ page }) => {
    await page.goto('/about');
    const ledeFirstLetter = await page.locator('.body .lede').evaluate((el) => {
      const style = window.getComputedStyle(el, '::first-letter');
      return {
        fontSize: parseFloat(style.fontSize),
        fontFamily: style.fontFamily,
      };
    });
    // Drop cap should be meaningfully bigger than body copy
    expect(ledeFirstLetter.fontSize).toBeGreaterThan(40);
    expect(ledeFirstLetter.fontFamily.toLowerCase()).toContain('bodoni');
  });

  test('colophon links are reachable and labeled', async ({ page }) => {
    await page.goto('/about');
    const meta = page.locator('.colophon .meta');
    await expect(meta.getByRole('link', { name: /retour au recueil/i })).toBeVisible();
    await expect(meta.getByRole('link', { name: /instagram/i })).toHaveAttribute('rel', /noopener/);
  });

  test('byline portrait falls back gracefully while /portrait.jpg is absent', async ({ page }) => {
    await page.goto('/about');
    const figure = page.locator('.byline-portrait');
    await expect(figure).toBeAttached();

    // Wait for the onerror handler to fire when the placeholder path 404s.
    await expect(figure).toHaveClass(/no-image/, { timeout: 3000 });
    await expect(figure.locator('.portrait-monogram')).toHaveText('HC');
    // The original <img> is removed from the DOM by the handler so it cannot
    // be read by screen readers.
    await expect(figure.locator('img')).toHaveCount(0);
  });

  test('masthead and meta use the "Hors-Champ" brand', async ({ page }) => {
    await page.goto('/about');
    await expect(page.locator('.masthead .edition')).toHaveText('Hors-Champ');
    await expect(page).toHaveTitle(/Hors-Champ/);
  });
});
