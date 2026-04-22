import { describe, it, expect } from 'vitest';
import { neutralTextForPalette } from '../../src/data/poemColor';

// Local re-implementations of the visible-bg model so tests can assert
// contrast without coupling to the internal helpers.
function hexToRgb(hex: string) {
  const h = hex.replace('#', '');
  const full = h.length === 3 ? h.split('').map(c => c + c).join('') : h;
  const n = parseInt(full, 16);
  return { r: (n >> 16) & 255, g: (n >> 8) & 255, b: n & 255 };
}
function srgbToLinear(c: number) {
  const s = c / 255;
  return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
}
function lum(hex: string) {
  const { r, g, b } = hexToRgb(hex);
  return 0.2126 * srgbToLinear(r) + 0.7152 * srgbToLinear(g) + 0.0722 * srgbToLinear(b);
}
function contrast(a: string, b: string) {
  const la = lum(a), lb = lum(b);
  const [hi, lo] = la > lb ? [la, lb] : [lb, la];
  return (hi + 0.05) / (lo + 0.05);
}
function mixRgb(a: string, b: string, pctA: number) {
  const ra = hexToRgb(a), rb = hexToRgb(b);
  const t = pctA / 100;
  return {
    r: Math.round(ra.r * t + rb.r * (1 - t)),
    g: Math.round(ra.g * t + rb.g * (1 - t)),
    b: Math.round(ra.b * t + rb.b * (1 - t)),
  };
}
function toHex(rgb: { r: number; g: number; b: number }) {
  const pad = (v: number) => v.toString(16).padStart(2, '0');
  return `#${pad(rgb.r)}${pad(rgb.g)}${pad(rgb.b)}`;
}
function visibleBg(base: string, theme: 'light' | 'dark') {
  const tint = theme === 'light' ? 10 : 12;
  const anchor = theme === 'light' ? '#f2f0eb' : '#020202';
  const blobOp = theme === 'light' ? 0.2 : 0.4;
  const bgBase = mixRgb(base, anchor, tint);
  const blob = hexToRgb(base);
  return toHex({
    r: Math.round(blob.r * blobOp + bgBase.r * (1 - blobOp)),
    g: Math.round(blob.g * blobOp + bgBase.g * (1 - blobOp)),
    b: Math.round(blob.b * blobOp + bgBase.b * (1 - blobOp)),
  });
}

const isGray = (hex: string) => {
  const { r, g, b } = hexToRgb(hex);
  return r === g && g === b;
};

describe('neutralTextForPalette', () => {
  it('returns sensible defaults when no base color is given', () => {
    const n = neutralTextForPalette(undefined);
    expect(isGray(n.light)).toBe(true);
    expect(isGray(n.dark)).toBe(true);
  });

  it('always returns achromatic grays (r === g === b)', () => {
    const palettes = ['#ececec', '#2f3640', '#80a0c0', '#ffaa00', '#000000', '#ffffff'];
    for (const p of palettes) {
      const n = neutralTextForPalette(p);
      expect(isGray(n.light), `light for ${p}: ${n.light}`).toBe(true);
      expect(isGray(n.dark), `dark for ${p}: ${n.dark}`).toBe(true);
    }
  });

  it('stays inside the editorial luminance band [#1a1a1a..#f0f0f0]', () => {
    const palettes = ['#ececec', '#2f3640', '#ffffff', '#000000', '#808080'];
    for (const p of palettes) {
      const { light, dark } = neutralTextForPalette(p);
      for (const hex of [light, dark]) {
        const v = parseInt(hex.slice(1, 3), 16);
        expect(v, `${hex} for ${p}`).toBeGreaterThanOrEqual(0x1a);
        expect(v, `${hex} for ${p}`).toBeLessThanOrEqual(0xf0);
      }
    }
  });

  it('meets or approaches WCAG AA (4.5:1) against the visible bg', () => {
    const palettes = ['#ececec', '#2f3640', '#80a0c0', '#ffaa00'];
    for (const p of palettes) {
      const n = neutralTextForPalette(p);
      const cLight = contrast(n.light, visibleBg(p, 'light'));
      const cDark = contrast(n.dark, visibleBg(p, 'dark'));
      // Allow a small shortfall when bg is mid-gray (AA unreachable);
      // best-effort is still > 4:1 for realistic palettes.
      expect(cLight, `light contrast for ${p}: ${cLight.toFixed(2)}`).toBeGreaterThanOrEqual(4.0);
      expect(cDark, `dark contrast for ${p}: ${cDark.toFixed(2)}`).toBeGreaterThanOrEqual(4.0);
    }
  });

  it('picks a darker gray when the visible bg is light', () => {
    const { light } = neutralTextForPalette('#ececec');
    const bg = visibleBg('#ececec', 'light');
    expect(lum(light)).toBeLessThan(lum(bg));
  });

  it('picks a lighter gray when the visible bg is dark', () => {
    const { dark } = neutralTextForPalette('#2f3640');
    const bg = visibleBg('#2f3640', 'dark');
    expect(lum(dark)).toBeGreaterThan(lum(bg));
  });

  it('accepts 3-digit hex shorthand', () => {
    expect(neutralTextForPalette('#fff')).toEqual(neutralTextForPalette('#ffffff'));
    expect(neutralTextForPalette('#000')).toEqual(neutralTextForPalette('#000000'));
  });

  it('produces continuous output — similar palettes map to similar grays', () => {
    const a = neutralTextForPalette('#ececec');
    const b = neutralTextForPalette('#e8e8e8');
    const delta = (h1: string, h2: string) =>
      Math.abs(parseInt(h1.slice(1, 3), 16) - parseInt(h2.slice(1, 3), 16));
    expect(delta(a.light, b.light)).toBeLessThanOrEqual(6);
    expect(delta(a.dark, b.dark)).toBeLessThanOrEqual(6);
  });
});
