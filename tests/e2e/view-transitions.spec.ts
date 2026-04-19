import { test, expect } from '@playwright/test';

/**
 * Persisted DOM nodes across Astro ClientRouter view transitions.
 *
 * Layout.astro marks a handful of nodes with `transition:persist` — they must
 * keep the same JS object identity across navigations so that effects tied
 * to them (cursor lerp, dust particles, aura colors, lightbox state) do not
 * reset. We tag each candidate node with a custom property before navigating
 * and verify the tag is still attached to the node in the new document.
 */
test.describe('View transitions — persist nodes', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
  });

  test('#custom-cursor, .bg-aura, #dust-canvas and #global-lightbox survive navigation', async ({ page }) => {
    await page.goto('/roll/matin-brumeux');
    await page.waitForLoadState('networkidle');

    // Tag each persisted node with a marker property.
    await page.evaluate(() => {
      const ids = ['custom-cursor', 'dust-canvas', 'global-lightbox'];
      for (const id of ids) {
        const el = document.getElementById(id);
        if (el) {
          (el as unknown as Record<string, string>).__persistTag = id + '-marker';
        }
      }
      const aura = document.querySelector('.bg-aura');
      if (aura) {
        (aura as unknown as Record<string, string>).__persistTag = 'bg-aura-marker';
      }
    });

    // Navigate via the next-roll link which uses the ClientRouter.
    const nextLink = page.locator('.next-roll-footer a.next-title');
    await expect(nextLink).toBeVisible();
    await nextLink.click();
    await page.waitForLoadState('networkidle');

    // Verify each tagged node is still present AND carries its tag. If Astro
    // re-created the node, the marker would have been lost.
    const results = await page.evaluate(() => {
      const check = (sel: string, expectedTag: string) => {
        const el = document.querySelector(sel);
        if (!el) return { found: false, preserved: false };
        const tag = (el as unknown as Record<string, string>).__persistTag;
        return { found: true, preserved: tag === expectedTag };
      };
      return {
        cursor: check('#custom-cursor', 'custom-cursor-marker'),
        aura: check('.bg-aura', 'bg-aura-marker'),
        dust: check('#dust-canvas', 'dust-canvas-marker'),
        lightbox: check('#global-lightbox', 'global-lightbox-marker'),
      };
    });

    expect(results.cursor.found).toBe(true);
    expect(results.cursor.preserved).toBe(true);
    expect(results.aura.found).toBe(true);
    expect(results.aura.preserved).toBe(true);
    expect(results.dust.found).toBe(true);
    expect(results.dust.preserved).toBe(true);
    expect(results.lightbox.found).toBe(true);
    expect(results.lightbox.preserved).toBe(true);
  });
});
