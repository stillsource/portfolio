import { test, expect } from '@playwright/test';

// Mobile-only smoke tests. See playwright.config.ts — this spec runs under
// the `mobile` project (iPhone 14 viewport + touch events).

test.describe('Touch gestures on mobile', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
    await page.goto('/roll/matin-brumeux');
    await page.locator('.clickable-image').first().waitFor({ state: 'visible' });
  });

  test('tapping a photo opens the lightbox', async ({ page }) => {
    await page.locator('.clickable-image').first().tap();
    await expect(page.locator('#global-lightbox')).toHaveClass(/is-open/);
  });

  test('swipe-down closes the lightbox', async ({ page }) => {
    await page.locator('.clickable-image').first().tap();
    await expect(page.locator('#global-lightbox')).toHaveClass(/is-open/);

    const img = page.locator('#lightbox-img');
    const box = await img.boundingBox();
    test.skip(!box, 'lightbox image not laid out');

    const startX = box!.x + box!.width / 2;
    const startY = box!.y + 40;
    // Playwright touch emulation: use CDP-powered touchscreen.
    await page.touchscreen.tap(startX, startY);
    await page.evaluate(
      ([sx, sy]) => {
        const el = document.getElementById('global-lightbox')!;
        const make = (type: string, y: number) => {
          const t = new Touch({
            identifier: 1,
            target: el,
            clientX: sx,
            clientY: y,
            radiusX: 1,
            radiusY: 1,
          });
          el.dispatchEvent(new TouchEvent(type, {
            cancelable: true,
            bubbles: true,
            touches: type === 'touchend' ? [] : [t],
            changedTouches: [t],
          }));
        };
        make('touchstart', sy);
        make('touchmove', sy + 200);
        make('touchend', sy + 200);
      },
      [startX, startY]
    );

    await expect(page.locator('#global-lightbox')).not.toHaveClass(/is-open/, { timeout: 2000 });
  });
});
