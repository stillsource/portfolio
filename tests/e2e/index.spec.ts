import { test, expect } from '@playwright/test';

test.describe('Index page', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
    await page.goto('/');
  });

  test('renders the header and the roll list', async ({ page }) => {
    const header = page.locator('.index-header');
    await expect(header).toBeVisible();
    await expect(header.locator('h2')).toHaveText('Recueil');

    const rolls = page.locator('.roll-item');
    await expect(rolls).toHaveCount(5);
  });

  test('rolls become visible after reveal', async ({ page }) => {
    const firstRoll = page.locator('.roll-item').first();
    await expect(firstRoll).toHaveClass(/is-visible/, { timeout: 3000 });
  });

  test('each roll shows title and date', async ({ page }) => {
    const firstLink = page.locator('.roll-link').first();
    await expect(firstLink.locator('.title')).not.toBeEmpty();
    await expect(firstLink.locator('.date')).not.toBeEmpty();
  });

  test('roll links point to /roll/[slug]', async ({ page }) => {
    const hrefs = await page.locator('a.roll-link').evaluateAll(
      (els) => els.map(el => el.getAttribute('href') ?? '')
    );
    expect(hrefs.length).toBe(5);
    hrefs.forEach(href => expect(href).toMatch(/^\/roll\/.+/));
  });

  test('roll page loadable directly', async ({ page }) => {
    // The hrefs are validated by "roll links point to /roll/[slug]"
    // Here we check that /roll/:slug is accessible and correct
    await page.goto('/roll/nuit-a-tokyo');
    await expect(page).toHaveURL('/roll/nuit-a-tokyo');
    await expect(page.locator('h1, .roll-title').first()).toBeVisible();
  });

  test('filter disclosure reveals a panel with "Tout" active by default', async ({ page }) => {
    const disclosureToggle = page.locator('#filter-disclosure-toggle');
    await expect(disclosureToggle).toBeVisible({ timeout: 3000 });
    await disclosureToggle.click();

    const allBtn = page.locator('.filter-btn[data-filter="all"]');
    await expect(allBtn).toBeVisible();
    await expect(allBtn).toHaveClass(/active/);
    await expect(allBtn).toHaveAttribute('aria-pressed', 'true');
  });

  test('filtering by tag hides non-matching rolls', async ({ page }) => {
    await expect(page.locator('.roll-item').first()).toHaveClass(/is-visible/, { timeout: 3000 });

    // Open the filter disclosure, then click the first non-"Tout" tag button.
    await page.locator('#filter-disclosure-toggle').click();
    const tagBtn = page.locator('.filter-btn:not([data-filter="all"])').first();
    const tagName = await tagBtn.getAttribute('data-filter');
    await tagBtn.click();

    await expect(tagBtn).toHaveClass(/active/);
    await expect(tagBtn).toHaveAttribute('aria-pressed', 'true');

    // Visible rolls should actually carry the tag
    const visibleRolls = page.locator('.roll-item:not(.hidden)');
    await expect(visibleRolls.first()).toBeVisible();
    const visibleCount = await visibleRolls.count();
    for (let i = 0; i < visibleCount; i++) {
      const tags = await visibleRolls.nth(i).getAttribute('data-tags');
      expect(JSON.parse(tags ?? '[]')).toContain(tagName);
    }
  });

  test('"Tout" shows all rolls again', async ({ page }) => {
    await expect(page.locator('.roll-item').first()).toHaveClass(/is-visible/, { timeout: 3000 });

    await page.locator('#filter-disclosure-toggle').click();
    await page.locator('.filter-btn:not([data-filter="all"])').first().click();
    await page.locator('.filter-btn[data-filter="all"]').click();

    const hiddenRolls = page.locator('.roll-item.hidden');
    await expect(hiddenRolls).toHaveCount(0);
  });
});
