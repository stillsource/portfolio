/**
 * Global state used to lock aura transitions during navigation.
 */
let isAuraLocked = false;
let lastPalette: string[] = [];
let lastVisibility = false;
let pendingPalette: { palette: string[], isVisible: boolean } | null = null;

export const setAuraLock = (lock: boolean) => {
  isAuraLocked = lock;
  if (lock) {
    document.documentElement.setAttribute('data-aura-locked', 'true');
  } else {
    document.documentElement.removeAttribute('data-aura-locked');
    // If there is a pending update, apply it now
    if (pendingPalette) {
      applyThemeColors(pendingPalette.palette, pendingPalette.isVisible);
      pendingPalette = null;
    }
  }
};

/**
 * Resets the palette cache to force an update on the next call.
 */
export const resetColorCache = () => {
  lastPalette = [];
  lastVisibility = false;
  pendingPalette = null;
};

/**
 * Updates the global CSS variables for the aura and text contrast.
 */
export function applyThemeColors(palette: string[], isVisible: boolean = true, force: boolean = false) {
  if (isAuraLocked && !force) {
    pendingPalette = { palette, isVisible };
    return;
  }

  // Quick comparison to avoid unnecessary work, unless forced
  if (!force && 
      isVisible === lastVisibility && 
      palette.length === lastPalette.length && 
      palette.every((c, i) => c === lastPalette[i])) {
    return;
  }
  
  lastPalette = [...palette];
  lastVisibility = isVisible;

  const root = document.documentElement;
  const isLight = root.getAttribute('data-theme') === 'light';
  const hasPalette = isVisible && palette?.length > 0;

  if (!hasPalette) {
    const defaultBg = isLight ? '#f2f0eb' : '#050505';
    for (let i = 1; i <= 5; i++) root.style.setProperty(`--p${i}`, defaultBg);
    root.style.setProperty('--text-main', isLight ? '#1a1a1a' : '#f0f0f0');
    root.style.setProperty('--text-muted', isLight ? '#666666' : '#aaaaaa');
    root.style.setProperty('--accent', isLight ? '#000000' : '#ffffff');
    root.style.setProperty('--aura-opacity', isLight ? '0.12' : '0.25');
    root.style.setProperty('--bg-base', defaultBg);
    return;
  }

  const baseColor = palette[0] || (isLight ? '#f2f0eb' : '#050505');
  // Update blob colors (max 5)
  for (let i = 0; i < 5; i++) {
    root.style.setProperty(`--p${i + 1}`, palette[i] || baseColor);
  }

  // Text contrast follows the explicit theme preference, not the palette
  // luminance — the user's toggle is the source of truth. The palette still
  // drives the blob colors and a subtle background tint.
  if (isLight) {
    root.style.setProperty('--text-main', '#1a1a1a');
    root.style.setProperty('--text-muted', '#555555');
    root.style.setProperty('--accent', '#000000');
    root.style.setProperty('--aura-opacity', '0.2');
    root.style.setProperty('--bg-base', `color-mix(in srgb, ${baseColor} 10%, #f2f0eb)`);
  } else {
    root.style.setProperty('--text-main', '#ffffff');
    root.style.setProperty('--text-muted', '#bbbbbb');
    root.style.setProperty('--accent', '#ffffff');
    root.style.setProperty('--aura-opacity', '0.4');
    root.style.setProperty('--bg-base', `color-mix(in srgb, ${baseColor} 12%, #020202)`);
  }
}

// Re-apply theme colors whenever <html data-theme> changes so the toggle
// is reflected immediately on pages where an aura palette is already active
// (the inline :root style.setProperty calls above override the CSS cascade).
if (typeof window !== 'undefined' && typeof MutationObserver !== 'undefined') {
  const html = document.documentElement;
  let currentTheme = html.getAttribute('data-theme');
  new MutationObserver(() => {
    const next = html.getAttribute('data-theme');
    if (next === currentTheme) return;
    currentTheme = next;
    applyThemeColors(lastPalette, lastVisibility, true);
  }).observe(html, { attributes: true, attributeFilter: ['data-theme'] });
}
