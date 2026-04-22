// Build-time helpers that pick a readable, neutrally-toned text color for
// UI layered over the aura (subtitle, EXIF captions, index excerpts).
//
// The solver returns an achromatic gray that is:
//   - as close to the visible background as possible (stays "neutral",
//     avoids harsh pure-black or pure-white against a colored aura),
//   - while still reaching ~4.5:1 contrast (WCAG AA) when achievable.
//
// Pure — no DOM, no browser APIs. Safe to import from Astro frontmatter.

export interface NeutralTextColors {
  light: string;
  dark: string;
}

function hexToRgb(hex: string): { r: number; g: number; b: number } {
  const h = hex.replace('#', '').trim();
  const full = h.length === 3 ? h.split('').map(c => c + c).join('') : h;
  const n = parseInt(full, 16);
  return { r: (n >> 16) & 255, g: (n >> 8) & 255, b: n & 255 };
}

function mixRgb(a: string, b: string, pctA: number): { r: number; g: number; b: number } {
  const ra = hexToRgb(a);
  const rb = hexToRgb(b);
  const t = pctA / 100;
  return {
    r: Math.round(ra.r * t + rb.r * (1 - t)),
    g: Math.round(ra.g * t + rb.g * (1 - t)),
    b: Math.round(ra.b * t + rb.b * (1 - t)),
  };
}

function srgbToLinear(c: number): number {
  const s = c / 255;
  return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
}

function linearToSrgbChannel(L: number): number {
  const clamped = Math.max(0, Math.min(1, L));
  const s = clamped <= 0.0031308 ? clamped * 12.92 : 1.055 * Math.pow(clamped, 1 / 2.4) - 0.055;
  return Math.round(s * 255);
}

function relativeLuminance(rgb: { r: number; g: number; b: number }): number {
  return 0.2126 * srgbToLinear(rgb.r) + 0.7152 * srgbToLinear(rgb.g) + 0.0722 * srgbToLinear(rgb.b);
}

// Approximate the composite background the user actually sees: the tinted
// bg-base plus the dominant aura blob at its rendering opacity. Mirrors the
// `color-mix` + blob-opacity used in src/scripts/utils/colors.ts.
function visibleBgLuminance(baseColor: string, theme: 'light' | 'dark'): number {
  const tintPct = theme === 'light' ? 10 : 12;
  const anchor = theme === 'light' ? '#f2f0eb' : '#020202';
  const blobOpacity = theme === 'light' ? 0.2 : 0.4;

  const bgBase = mixRgb(baseColor, anchor, tintPct);
  const blob = hexToRgb(baseColor);
  const composite = {
    r: Math.round(blob.r * blobOpacity + bgBase.r * (1 - blobOpacity)),
    g: Math.round(blob.g * blobOpacity + bgBase.g * (1 - blobOpacity)),
    b: Math.round(blob.b * blobOpacity + bgBase.b * (1 - blobOpacity)),
  };
  return relativeLuminance(composite);
}

// Luminance equivalents of the "editorial extremes" we never want to exceed —
// pure black/white reads too harsh on a photographic aura.
const MIN_TEXT_LUM = srgbToLinear(0x1a); // ≈ 0.0131 — "#1a1a1a"
const MAX_TEXT_LUM = srgbToLinear(0xf0); // ≈ 0.8713 — "#f0f0f0"
const TARGET_CONTRAST = 4.5;

/**
 * Solve for an achromatic gray whose contrast against `bgLum` is ≥ target,
 * and that stays as close to the bg as possible (maximally neutral).
 * Falls back to the extreme candidate in the better direction when the
 * bg is mid-gray and AA is unreachable.
 */
function neutralGrayForBgLum(bgLum: number): string {
  // Maximum achievable contrast in each direction.
  const darkMaxC = (bgLum + 0.05) / (MIN_TEXT_LUM + 0.05);
  const lightMaxC = (MAX_TEXT_LUM + 0.05) / (bgLum + 0.05);

  const pickDark = darkMaxC >= lightMaxC;

  let Lt: number;
  if (pickDark) {
    // Target contrast → Lt = (bgLum + 0.05) / C − 0.05. Clamp to [MIN, bg).
    const ideal = (bgLum + 0.05) / TARGET_CONTRAST - 0.05;
    Lt = Math.max(MIN_TEXT_LUM, Math.min(bgLum, ideal));
  } else {
    // Target contrast → Lt = C * (bgLum + 0.05) − 0.05. Clamp to (bg, MAX].
    const ideal = TARGET_CONTRAST * (bgLum + 0.05) - 0.05;
    Lt = Math.min(MAX_TEXT_LUM, Math.max(bgLum, ideal));
  }

  const c = linearToSrgbChannel(Lt);
  const hex = c.toString(16).padStart(2, '0');
  return `#${hex}${hex}${hex}`;
}

/**
 * Given a palette's dominant (first) color, return the neutral text gray to
 * use in each theme. Embed both values inline on the element — CSS selects
 * between them via `:root[data-theme]`.
 */
export function neutralTextForPalette(baseColor: string | undefined): NeutralTextColors {
  if (!baseColor) {
    return { light: '#3a3a3a', dark: '#c8c8c8' };
  }
  return {
    light: neutralGrayForBgLum(visibleBgLuminance(baseColor, 'light')),
    dark: neutralGrayForBgLum(visibleBgLuminance(baseColor, 'dark')),
  };
}
