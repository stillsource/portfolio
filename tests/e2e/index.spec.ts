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

  test('filter section visible with "Tout" button active', async ({ page }) => {
    const filterSection = page.locator('.filter-section');
    await expect(filterSection).toHaveClass(/is-visible/, { timeout: 3000 });

    const allBtn = page.locator('.filter-btn[data-filter="all"]');
    await expect(allBtn).toHaveClass(/active/);
    await expect(allBtn).toHaveAttribute('aria-pressed', 'true');
  });

  test('filtering by tag hides non-matching rolls', async ({ page }) => {
    // Wait for the rolls to be visible
    await expect(page.locator('.roll-item').first()).toHaveClass(/is-visible/, { timeout: 3000 });

    // Click on the first available tag (not "Tout")
    const tagBtn = page.locator('.filter-btn:not([data-filter="all"])').first();
    const tagName = await tagBtn.getAttribute('data-filter');
    await tagBtn.click();

    await expect(tagBtn).toHaveClass(/active/);
    await expect(tagBtn).toHaveAttribute('aria-pressed', 'true');

    // Wait for transitions to finish (600ms delay + margin)
    await page.waitForTimeout(800);

    // Verify that rolls without this tag are hidden
    const visibleRolls = page.locator('.roll-item:not(.hidden)');

    // At least one roll visible
    await expect(visibleRolls.first()).toBeVisible();

    // Visible rolls should actually carry the tag
    const visibleCount = await visibleRolls.count();
    for (let i = 0; i < visibleCount; i++) {
      const tags = await visibleRolls.nth(i).getAttribute('data-tags');
      expect(JSON.parse(tags ?? '[]')).toContain(tagName);
    }
  });

  test('"Tout" shows all rolls again', async ({ page }) => {
    await expect(page.locator('.roll-item').first()).toHaveClass(/is-visible/, { timeout: 3000 });

    // Filter, wait for the setTimeout(600ms) to finish, then "Tout"
    await page.locator('.filter-btn:not([data-filter="all"])').first().click();
    await page.waitForTimeout(800);
    await page.locator('.filter-btn[data-filter="all"]').click();
    await page.waitForTimeout(300);

    const hiddenRolls = page.locator('.roll-item.hidden');
    await expect(hiddenRolls).toHaveCount(0);
  });
});
