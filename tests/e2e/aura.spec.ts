import { test, expect } from '@playwright/test';

/**
 * Ambient Color Engine ("Aura 2.0") — end-to-end coverage.
 *
 * These tests verify the three user-visible contracts of the aura system:
 *   1. Roll pages load their palette into the --p1..--p5 CSS custom properties
 *      which propagate to the .aura-blob divs' background-color.
 *   2. Navigation between rolls sets data-aura-locked on <html> during the
 *      `astro:before-preparation` phase to prevent color flicker (Color-Lock).
 *   3. The aura container is marked transition:persist so its DOM nodes survive
 *      Astro's view transitions — same object identity before and after nav.
 *   4. On the index, hovering a roll link previews that roll's palette via
 *      an updated --p1 custom property.
 */
test.describe('Aura 2.0', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => sessionStorage.setItem('splashSeen', 'true'));
  });

  test('blob .b1 resolves to the roll palette on a roll page', async ({ page }) => {
    await page.goto('/roll/matin-brumeux');
    await page.waitForLoadState('networkidle');

    // --p1 is updated by Roll.astro's init script (applyThemeColors(..., true, true)).
    // Poll until the custom property diverges from the default #050505 rather
    // than relying on a fixed sleep — Houdini transitions can run up to 3s.
    await page.waitForFunction(
      () => {
        const p1 = getComputedStyle(document.documentElement).getPropertyValue('--p1').trim();
        // Default initial value and dark-mode fallback both normalize to #050505.
        return p1 !== '' && p1.toLowerCase() !== '#050505' && p1 !== 'rgb(5, 5, 5)';
      },
      null,
      { timeout: 5000 },
    );

    // Custom property --p1 must have shifted away from the default black.
    const p1 = await page.evaluate(() =>
      getComputedStyle(document.documentElement).getPropertyValue('--p1').trim(),
    );
    expect(p1).not.toBe('rgb(5, 5, 5)');
    expect(p1).not.toBe('#050505');
    expect(p1).not.toBe('');

    // The .b1 blob applies `background-color: var(--p1)` with a 3s CSS
    // transition. Wait for the transition to complete, then assert the
    // resolved color also left the default.
    await page.waitForFunction(
      () => {
        const bg = getComputedStyle(document.querySelector('.aura-blob.b1')!).backgroundColor;
        return bg !== 'rgb(5, 5, 5)';
      },
      null,
      { timeout: 6000 },
    );
    const bgColor = await page.locator('.aura-blob.b1').evaluate(
      el => getComputedStyle(el).backgroundColor,
    );
    expect(bgColor).toMatch(/rgba?\(/);
  });

  test('<html data-aura-locked> is set during astro:before-preparation', async ({ page }) => {
    await page.goto('/roll/matin-brumeux');
    await page.waitForLoadState('networkidle');

    // Subscribe to the event and record whether the attribute was set at
    // the moment the handler fires. We expose the capture via window.
    await page.evaluate(() => {
      (window as unknown as { __auraCapture: { locked: boolean | null } }).__auraCapture = {
        locked: null,
      };
      document.addEventListener('astro:before-preparation', () => {
        const html = document.documentElement;
        (window as unknown as { __auraCapture: { locked: boolean | null } }).__auraCapture.locked =
          html.hasAttribute('data-aura-locked');
      });
    });

    // Trigger a navigation by following the next-roll link.
    const nextLink = page.locator('.next-roll-footer a.next-title');
    await expect(nextLink).toBeVisible();
    await nextLink.click();
    await page.waitForLoadState('networkidle');

    const captured = await page.evaluate(
      () => (window as unknown as { __auraCapture: { locked: boolean | null } }).__auraCapture.locked,
    );
    expect(captured).toBe(true);
  });

  test('.bg-aura survives a navigation (transition:persist)', async ({ page }) => {
    await page.goto('/roll/matin-brumeux');
    await page.waitForLoadState('networkidle');

    // Tag the persistent DOM node before navigating.
    await page.evaluate(() => {
      const aura = document.querySelector('.bg-aura');
      if (aura) {
        (aura as unknown as { __tag: string }).__tag = 'persisted-aura';
      }
    });

    const nextLink = page.locator('.next-roll-footer a.next-title');
    await nextLink.click();
    await page.waitForLoadState('networkidle');

    // After navigation the same DOM node must still be present and tagged.
    const tagStillPresent = await page.evaluate(() => {
      const aura = document.querySelector('.bg-aura');
      return aura ? (aura as unknown as { __tag?: string }).__tag === 'persisted-aura' : false;
    });
    expect(tagStillPresent).toBe(true);
  });

  test('first ClientRouter navigation refreshes --p1 (regression)', async ({ page }) => {
    // Regression test for the bug where Astro 6.1.1's ClientRouter crashed
    // on the first cross-roll navigation because `Element.moveBefore` rejects
    // divs as direct children of <html>. The fix (inline script removing
    // Element.prototype.moveBefore) forces Astro onto its appendChild fallback
    // so `astro:after-swap` fires and the destination palette is applied.
    await page.goto('/roll/matin-brumeux');
    await page.waitForLoadState('networkidle');
    await page.waitForFunction(
      () => {
        const p1 = getComputedStyle(document.documentElement).getPropertyValue('--p1').trim();
        return p1 !== '' && p1.toLowerCase() !== '#050505' && p1 !== 'rgb(5, 5, 5)';
      },
      null,
      { timeout: 5000 },
    );
    const before = await page.evaluate(() =>
      getComputedStyle(document.documentElement).getPropertyValue('--p1').trim(),
    );

    const nextLink = page.locator('.next-roll-footer a.next-title');
    await nextLink.scrollIntoViewIfNeeded();
    await page.waitForTimeout(200);
    await nextLink.click({ force: true });
    await page.waitForURL('**/roll/nuit-a-tokyo', { timeout: 10000 });
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(400);

    const after = await page.evaluate(() =>
      getComputedStyle(document.documentElement).getPropertyValue('--p1').trim(),
    );
    expect(after).not.toBe(before);
    // Nuit-à-Tokyo's first palette color is #1a1c2c → rgb(26, 28, 44).
    expect(after).toBe('rgb(26, 28, 44)');
  });

  test('hovering a roll link on the index updates --p1', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    // Allow the staggered reveal animation to finish.
    await page.waitForTimeout(600);

    const before = await page.evaluate(() =>
      getComputedStyle(document.documentElement).getPropertyValue('--p1').trim(),
    );

    const firstRollLink = page.locator('a.roll-link').first();
    await firstRollLink.hover();
    // Houdini color transitions run for up to 3s; poll for a change.
    await page.waitForFunction(
      (prev) =>
        getComputedStyle(document.documentElement).getPropertyValue('--p1').trim() !== prev,
      before,
      { timeout: 4000 },
    );

    const after = await page.evaluate(() =>
      getComputedStyle(document.documentElement).getPropertyValue('--p1').trim(),
    );
    expect(after).not.toBe(before);
  });
});
