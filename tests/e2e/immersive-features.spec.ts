/**
 * BDD — Immersive features
 *
 * Feature: Photo develop effect
 * Feature: Film frame counter
 * Feature: 3D tilt on photos
 * Feature: Custom cursor viewfinder
 * Feature: Dust particles in chambre noire
 */

import { test, expect } from '@playwright/test';

// ─────────────────────────────────────────────────────────────────────────────
// FEATURE: Film development effect
// Photos appear progressively, like in a developer tray
// ─────────────────────────────────────────────────────────────────────────────
test.describe('Feature: Photo develop effect', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
    await page.goto('/roll/matin-brumeux');
    await page.waitForLoadState('networkidle');
  });

  test('Scenario: image visible on load acquires .developed', async ({ page }) => {
    // Given: an image in the initial viewport
    const firstImg = page.locator('.styled-image').first();
    await expect(firstImg).toBeVisible({ timeout: 5000 });

    // When: we wait for IntersectionObserver + setTimeout(120ms) to run
    await page.waitForTimeout(600);

    // Then: the .developed class must be present
    await expect(firstImg).toHaveClass(/developed/, { timeout: 3000 });
  });

  test('Scenario: developed image has a filter with no perceptible blur', async ({ page }) => {
    // Given: the first image is visible and developed
    const firstImg = page.locator('.styled-image').first();
    await expect(firstImg).toHaveClass(/developed/, { timeout: 4000 });

    // When: we poll from the browser to let the 0.7s CSS transition progress
    // (page.waitForTimeout throttles Chromium transitions in headless mode)
    await page.waitForFunction(
      () => {
        const img = document.querySelector('.styled-image');
        if (!img) return false;
        const filter = getComputedStyle(img).filter;
        const match = filter.match(/blur\(([0-9.]+)px\)/);
        return !match || parseFloat(match[1]) < 0.5;
      },
      { timeout: 5000, polling: 50 }
    );

    // Then: blur is below 0.5px
    const filterVal = await firstImg.evaluate(el => getComputedStyle(el).filter);
    const blurMatch = filterVal.match(/blur\(([0-9.]+)px\)/);
    const blurPx = blurMatch ? parseFloat(blurMatch[1]) : 0;
    expect(blurPx).toBeLessThan(0.5);
  });

  test('Scenario: off-viewport image does not develop prematurely', async ({ page }) => {
    // Given: several images on the page
    const allImgs = page.locator('.styled-image');
    const count = await allImgs.count();
    if (count < 2) test.skip();

    // When: we are at the top of the page (scroll = 0)
    await page.evaluate(() => window.scrollTo(0, 0));
    await page.waitForTimeout(300);

    // Then: the last image (off-viewport) does not yet have .developed
    const lastImg = allImgs.last();
    const inViewport = await lastImg.evaluate(el => {
      const r = el.getBoundingClientRect();
      return r.top < window.innerHeight;
    });
    if (!inViewport) {
      const hasDeveloped = await lastImg.evaluate(el => el.classList.contains('developed'));
      expect(hasDeveloped).toBe(false);
    }
  });

  test('Scenario: off-viewport image develops when it enters the viewport', async ({ page }) => {
    // Given: an image not yet visible
    const allImgs = page.locator('.styled-image');
    const count = await allImgs.count();
    if (count < 2) test.skip();

    const lastImg = allImgs.last();
    const isInitiallyVisible = await lastImg.evaluate(el => {
      const r = el.getBoundingClientRect();
      return r.top < window.innerHeight;
    });
    if (isInitiallyVisible) test.skip();

    // When: we scroll down to it
    await lastImg.scrollIntoViewIfNeeded();
    await page.waitForTimeout(600);

    // Then: it is developed
    await expect(lastImg).toHaveClass(/developed/, { timeout: 3000 });
  });

  test('Scenario: all images visible after full scroll are developed', async ({ page }) => {
    // Given / When: scroll to the bottom of the page
    await page.keyboard.press('End');
    await page.waitForTimeout(1000);

    // Then: all images in the viewport have .developed
    const result = await page.evaluate(() => {
      const imgs = Array.from(document.querySelectorAll('.styled-image'));
      return imgs
        .filter(img => {
          const r = img.getBoundingClientRect();
          return r.top < window.innerHeight && r.bottom > 0;
        })
        .every(img => img.classList.contains('developed'));
    });
    expect(result).toBe(true);
  });
});

// ─────────────────────────────────────────────────────────────────────────────
// FEATURE: Film frame counter
// ─────────────────────────────────────────────────────────────────────────────
test.describe('Feature: Film counter', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
    await page.goto('/roll/matin-brumeux');
    await page.waitForLoadState('networkidle');
  });

  test('Scenario: counter present in the DOM', async ({ page }) => {
    const counter = page.locator('#film-counter');
    await expect(counter).toBeAttached();
  });

  test('Scenario: counter shows the correct total (visual groups)', async ({ page }) => {
    // Given: the total matches scroll groups (imageGroups.length), not individual photos
    const total = await page.locator('#film-counter .film-total').textContent();
    const groupCount = await page.locator('.image-wrapper, .image-pair').count();
    // imageGroups.length = number of .image-wrapper + .image-pair in the DOM
    expect(total?.trim()).toBe(groupCount.toString().padStart(2, '0'));
  });

  test('Scenario: counter becomes visible on scroll into the roll', async ({ page }) => {
    // Given: counter initially transparent (opacity: 0)
    const counter = page.locator('#film-counter');

    // When: we scroll into the roll
    await page.locator('.roll-section').scrollIntoViewIfNeeded();
    await page.waitForTimeout(300);

    // Then: it gains the is-active class
    await expect(counter).toHaveClass(/is-active/, { timeout: 3000 });
  });

  test('Scenario: current number updates while scrolling', async ({ page }) => {
    // Given: initial number = 01
    const numEl = page.locator('#film-current');
    await expect(numEl).toHaveText('01');

    // When: we scroll to the 2nd frame and wait for the async IO
    const frames = page.locator('.image-wrapper, .image-pair');
    const frameCount = await frames.count();
    if (frameCount < 2) test.skip();

    const secondFrame = frames.nth(1);
    await secondFrame.scrollIntoViewIfNeeded();
    // IO is async + possible rendering delay — wait until the number changes
    await page.waitForFunction(
      () => {
        const el = document.getElementById('film-current');
        return el ? parseInt(el.textContent ?? '1') >= 2 : false;
      },
      { timeout: 3000 }
    );

    // Then: the number has advanced
    const newNum = await numEl.textContent();
    expect(parseInt(newNum ?? '1')).toBeGreaterThanOrEqual(2);
  });
});

// ─────────────────────────────────────────────────────────────────────────────
// FEATURE: 3D parallax on hover
// ─────────────────────────────────────────────────────────────────────────────
test.describe('Feature: 3D photo parallax', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
    await page.goto('/roll/matin-brumeux');
    await page.waitForLoadState('networkidle');
  });

  test('Scenario: image-container has a transform transition defined', async ({ page }) => {
    // Given: a clickable image
    const container = page.locator('.clickable-image').first();
    await expect(container).toBeVisible({ timeout: 5000 });

    // Then: the CSS transition includes transform
    const transition = await container.evaluate(el => getComputedStyle(el).transition);
    expect(transition).toContain('transform');
  });

  test('Scenario: hover applies a perspective transform', async ({ page }) => {
    // Given: a visible image
    const container = page.locator('.clickable-image').first();
    await expect(container).toBeVisible({ timeout: 5000 });

    // When: we hover the image
    await container.hover();
    await page.waitForTimeout(200);

    // Then: a transform with perspective is applied
    const transform = await container.evaluate(el => (el as HTMLElement).style.transform);
    expect(transform).toContain('perspective');
  });

  test('Scenario: after hover, transform resets', async ({ page }) => {
    const container = page.locator('.clickable-image').first();
    await expect(container).toBeVisible({ timeout: 5000 });

    // When: we hover then leave
    await container.hover();
    await page.waitForTimeout(100);
    await page.mouse.move(0, 0); // leave the image
    await page.waitForTimeout(600); // wait for 0.5s transition

    // Then: transform is empty or identity
    const transform = await container.evaluate(el => (el as HTMLElement).style.transform);
    expect(transform).toBe('');
  });
});

// ─────────────────────────────────────────────────────────────────────────────
// FEATURE: Custom cursor presence
// ─────────────────────────────────────────────────────────────────────────────
test.describe('Feature: Custom cursor', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
    await page.goto('/roll/matin-brumeux');
    await page.waitForLoadState('networkidle');
  });

  test('Scenario: cursor exists in the DOM on a roll page', async ({ page }) => {
    await expect(page.locator('#custom-cursor')).toBeAttached();
  });
});

// ─────────────────────────────────────────────────────────────────────────────
// FEATURE: Dust particles in chambre noire mode
// ─────────────────────────────────────────────────────────────────────────────
test.describe('Feature: Chambre noire particles', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
    await page.goto('/roll/matin-brumeux');
    await page.waitForLoadState('networkidle');
  });

  test('Scenario: dust canvas present in the DOM', async ({ page }) => {
    await expect(page.locator('#dust-canvas')).toBeAttached();
  });

  test('Scenario: canvas invisible by default (outside chambre noire)', async ({ page }) => {
    const canvas = page.locator('#dust-canvas');
    const opacity = await canvas.evaluate(el => getComputedStyle(el).opacity);
    expect(parseFloat(opacity)).toBe(0);
  });

  test('Scenario: canvas visible in chambre noire mode', async ({ page }) => {
    // When: we enable chambre noire
    await page.keyboard.press('h');
    await expect(page.locator('body')).toHaveClass(/chambre-noire/, { timeout: 1000 });

    // Chromium throttles transitions when Playwright is idle — poll from the browser
    // to keep JS active and let the CSS transitions progress
    await page.waitForFunction(
      () => parseFloat(getComputedStyle(document.getElementById('dust-canvas')!).opacity) > 0.9,
      { timeout: 3000, polling: 50 }
    );

    const opacity = await page.locator('#dust-canvas').evaluate(el => parseFloat(getComputedStyle(el).opacity));
    expect(opacity).toBeGreaterThan(0.9);
  });

  test('Scenario: canvas becomes invisible again when chambre noire is disabled', async ({ page }) => {
    // Enable then disable — opacity must return to 0
    await page.keyboard.press('h');
    await page.waitForFunction(
      () => parseFloat(getComputedStyle(document.getElementById('dust-canvas')!).opacity) > 0.9,
      { timeout: 3000, polling: 50 }
    );

    await page.keyboard.press('h'); // disable
    await page.waitForFunction(
      () => parseFloat(getComputedStyle(document.getElementById('dust-canvas')!).opacity) < 0.1,
      { timeout: 3000, polling: 50 }
    );

    const opacity = await page.locator('#dust-canvas').evaluate(el => parseFloat(getComputedStyle(el).opacity));
    expect(opacity).toBeLessThan(0.1);
  });
});
