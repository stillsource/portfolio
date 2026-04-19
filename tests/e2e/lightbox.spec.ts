import { test, expect } from '@playwright/test';

test.describe('LightboxGlobal (extracted)', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
    await page.goto('/roll/matin-brumeux');
    // Wait until at least one image is rendered so clicking it opens the lightbox.
    await page.locator('.clickable-image').first().waitFor({ state: 'visible' });
  });

  test('clicking a photo opens the lightbox', async ({ page }) => {
    await page.locator('.clickable-image').first().click();
    await expect(page.locator('#global-lightbox')).toHaveClass(/is-open/);
    await expect(page.locator('#global-lightbox')).toHaveAttribute('aria-hidden', 'false');
  });

  test('Escape closes the lightbox', async ({ page }) => {
    await page.locator('.clickable-image').first().click();
    await expect(page.locator('#global-lightbox')).toHaveClass(/is-open/);
    await page.keyboard.press('Escape');
    await expect(page.locator('#global-lightbox')).not.toHaveClass(/is-open/);
  });

  test('ArrowRight navigates to the next image', async ({ page }) => {
    const clickable = page.locator('.clickable-image');
    const count = await clickable.count();
    test.skip(count < 2, 'roll has fewer than 2 images');

    await clickable.first().click();
    const counter = page.locator('.lightbox-counter');
    await expect(counter).toHaveText(/1 \/ \d+/);
    await page.keyboard.press('ArrowRight');
    await expect(counter).toHaveText(/2 \/ \d+/);
  });

  test('lightbox DOM is present on every page (global component)', async ({ page }) => {
    await expect(page.locator('#global-lightbox')).toBeAttached();
    await page.goto('/');
    await expect(page.locator('#global-lightbox')).toBeAttached();
  });

  test('clicking outside the prev/next buttons closes the lightbox', async ({ page }) => {
    await page.locator('.clickable-image').first().click();
    await expect(page.locator('#global-lightbox')).toHaveClass(/is-open/);
    // Click the backdrop corner — away from any control.
    await page.mouse.click(5, 5);
    await expect(page.locator('#global-lightbox')).not.toHaveClass(/is-open/);
  });

  test('clicking prev/next does not close the lightbox', async ({ page }) => {
    const count = await page.locator('.clickable-image').count();
    test.skip(count < 2, 'roll has fewer than 2 images');
    await page.locator('.clickable-image').first().click();
    await page.locator('.lightbox-next').click();
    await expect(page.locator('#global-lightbox')).toHaveClass(/is-open/);
    await expect(page.locator('.lightbox-counter')).toHaveText(/2 \/ \d+/);
  });

  test('counter is positioned at the top (away from exif at bottom)', async ({ page }) => {
    await page.locator('.clickable-image').first().click();
    const counter = page.locator('.lightbox-counter');
    const box = await counter.boundingBox();
    expect(box).not.toBeNull();
    // Counter should be in the top third of the viewport so it never overlaps exif.
    const viewportHeight = page.viewportSize()?.height ?? 0;
    expect(box!.y).toBeLessThan(viewportHeight / 3);
  });
});

test.describe('Lightbox scroll-to-photo on close', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
    await page.goto('/roll/matin-brumeux');
    await page.locator('.clickable-image').first().waitFor({ state: 'visible' });
  });

  test('closing after navigating scrolls the page to the last-viewed photo', async ({ page }) => {
    const images = page.locator('.clickable-image');
    const count = await images.count();
    test.skip(count < 3, 'need at least 3 images');

    // Open first image, navigate to the third, then close.
    await images.first().click();
    await expect(page.locator('#global-lightbox')).toHaveClass(/is-open/);
    await page.keyboard.press('ArrowRight');
    await page.keyboard.press('ArrowRight');
    await expect(page.locator('.lightbox-counter')).toHaveText(/3 \/ \d+/);
    await page.keyboard.press('Escape');
    await expect(page.locator('#global-lightbox')).not.toHaveClass(/is-open/);

    // Wait for the smooth scroll to settle around the third image.
    const vh = page.viewportSize()?.height ?? 0;
    await expect.poll(async () => {
      const box = await images.nth(2).boundingBox();
      if (!box) return Number.POSITIVE_INFINITY;
      return Math.abs(box.y + box.height / 2 - vh / 2);
    }, { timeout: 5000 }).toBeLessThan(vh / 2);
  });
});

test.describe('Lightbox long-poem layout', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
    await page.goto('/roll/matin-brumeux');
    await page.locator('.clickable-image').first().waitFor({ state: 'visible' });
  });

  test('short poem keeps the default (landscape/portrait) layout — no long-poem class', async ({ page }) => {
    await page.locator('.clickable-image').first().click();
    await expect(page.locator('#global-lightbox')).toHaveClass(/is-open/);
    await expect(page.locator('#global-lightbox')).not.toHaveClass(/long-poem/);
  });

  test('long multi-stanza poem triggers the side layout', async ({ page }) => {
    // Dispatch open-lightbox directly with a synthetic long poem.
    await page.evaluate(() => {
      const longPoem = [
        'Au matin, la brume s\'étire',
        'Sur les cimes endormies',
        'Un souffle blanc, un soupir',
        'Et le jour qui se dénie',
        '',
        'La lumière, hésitante encore,',
        'Effleure les bourgeons d\'argent',
      ].join('\n');
      window.dispatchEvent(new CustomEvent('open-lightbox', {
        detail: {
          images: [{ url: 'https://picsum.photos/seed/long/800/600', poem: longPoem, exif: 'TEST' }],
          index: 0,
        },
      }));
    });
    await expect(page.locator('#global-lightbox')).toHaveClass(/is-open/);
    await expect(page.locator('#global-lightbox')).toHaveClass(/long-poem/);
  });

  test('poem preserves line breaks (white-space: pre-line)', async ({ page }) => {
    await page.evaluate(() => {
      window.dispatchEvent(new CustomEvent('open-lightbox', {
        detail: {
          images: [{ url: 'https://picsum.photos/seed/v/800/600', poem: 'Ligne une\nLigne deux\nLigne trois', exif: '' }],
          index: 0,
        },
      }));
    });
    const ws = await page.locator('#lightbox-poem').evaluate(el => getComputedStyle(el).whiteSpace);
    expect(ws).toBe('pre-line');
  });
});

test.describe('Lightbox loupe (hold-to-magnify)', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
    await page.goto('/roll/matin-brumeux');
    await page.locator('.clickable-image').first().waitFor({ state: 'visible' });
    await page.locator('.clickable-image').first().click();
    await expect(page.locator('#global-lightbox')).toHaveClass(/is-open/);
    // Wait until the lightbox image is loaded so getBoundingClientRect is valid.
    await page.locator('#lightbox-img').waitFor({ state: 'visible' });
    // Wait for the image to be loaded AND for the landscape/portrait class to
    // be applied — the final element rect is only settled after that.
    await page.waitForFunction(() => {
      const img = document.getElementById('lightbox-img') as HTMLImageElement | null;
      const lb = document.getElementById('global-lightbox');
      return !!img && img.complete && img.naturalWidth > 0 && !!lb && (lb.classList.contains('landscape') || lb.classList.contains('portrait'));
    });
  });

  const imageCenter = async (page: import('@playwright/test').Page) => {
    const box = await page.locator('#lightbox-img').boundingBox();
    if (!box) throw new Error('lightbox image has no bounding box');
    return { x: box.x + box.width / 2, y: box.y + box.height / 2 };
  };

  test('pointerdown on the image activates the loupe', async ({ page }) => {
    const { x, y } = await imageCenter(page);
    await page.mouse.move(x, y);
    await page.mouse.down();
    await expect(page.locator('#lightbox-loupe')).toHaveClass(/is-active/);
    await expect(page.locator('body')).toHaveClass(/loupe-active/);
    const bg = await page.locator('#lightbox-loupe').evaluate(el => (el as HTMLElement).style.backgroundImage);
    expect(bg).toContain('url');
    await page.mouse.up();
  });

  test('pointerup deactivates the loupe', async ({ page }) => {
    const { x, y } = await imageCenter(page);
    await page.mouse.move(x, y);
    await page.mouse.down();
    await expect(page.locator('#lightbox-loupe')).toHaveClass(/is-active/);
    await page.mouse.up();
    await expect(page.locator('#lightbox-loupe')).not.toHaveClass(/is-active/);
    await expect(page.locator('body')).not.toHaveClass(/loupe-active/);
  });

  test('press-release on the image does not close the lightbox', async ({ page }) => {
    const { x, y } = await imageCenter(page);
    await page.mouse.move(x, y);
    await page.mouse.down();
    await page.mouse.up();
    await expect(page.locator('#global-lightbox')).toHaveClass(/is-open/);
  });

  test('loupe updates its background-position when the pointer moves', async ({ page }) => {
    const box = await page.locator('#lightbox-img').boundingBox();
    if (!box) throw new Error('no bbox');
    const x1 = box.x + box.width * 0.25;
    const y1 = box.y + box.height * 0.25;
    const x2 = box.x + box.width * 0.75;
    const y2 = box.y + box.height * 0.75;
    await page.mouse.move(x1, y1);
    await page.mouse.down();
    await expect(page.locator('#lightbox-loupe')).toHaveClass(/is-active/);
    const posStart = await page.locator('#lightbox-loupe').evaluate(el => (el as HTMLElement).style.backgroundPosition);
    await page.mouse.move(x2, y2);
    const posEnd = await page.locator('#lightbox-loupe').evaluate(el => (el as HTMLElement).style.backgroundPosition);
    expect(posStart).not.toBe('');
    expect(posEnd).not.toBe('');
    expect(posStart).not.toBe(posEnd);
    await page.mouse.up();
  });

  test('computeLoupeStyle geometry', async ({ page }) => {
    // Evaluate the pure function inside the page so we exercise the same module
    // that ships to production.
    const result = await page.evaluate(async () => {
      // Vite serves TS modules via absolute path in dev.
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const mod: any = await import(/* @vite-ignore */ ('/src/scripts/lightbox/loupe.ts' as string));
      const rect = { left: 0, top: 0, right: 800, bottom: 600, width: 800, height: 600 };
      const inside = mod.computeLoupeStyle({ rect, clientX: 400, clientY: 300, zoom: 2, diameter: 200 });
      const outside = mod.computeLoupeStyle({ rect, clientX: -10, clientY: 300, zoom: 2, diameter: 200 });
      return { inside, outside };
    });
    expect(result.outside.inside).toBe(false);
    expect(result.inside.inside).toBe(true);
    expect(result.inside.bgSize).toBe('1600px 1200px');
    // cursor at image center → rx=400, ry=300 → -(400*2 - 100)=-700, -(300*2 - 100)=-500
    expect(result.inside.bgPosition).toBe('-700px -500px');
    expect(result.inside.left).toBe(400);
    expect(result.inside.top).toBe(300);
  });
});
