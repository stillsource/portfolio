/**
 * Calcule la luminance moyenne d'une palette de couleurs hexadécimales.
 * Optimisé avec une approche linéaire.
 */
export function getAverageLuminance(palette: string[]): number {
  if (!palette?.length) return 0;
  
  let totalLuminance = 0;
  for (const hex of palette) {
    const cleanHex = hex.startsWith('#') ? hex.slice(1) : hex;
    const r = parseInt(cleanHex.slice(0, 2), 16);
    const g = parseInt(cleanHex.slice(2, 4), 16);
    const b = parseInt(cleanHex.slice(4, 6), 16);
    // Standard relative luminance formula
    totalLuminance += (0.299 * r + 0.587 * g + 0.114 * b);
  }
  
  return (totalLuminance / palette.length) / 255;
}

/**
 * État global pour verrouiller les transitions de l'aura pendant la navigation.
 */
let isAuraLocked = false;
let lastPalette: string[] = [];
let lastVisibility = false;

export const setAuraLock = (lock: boolean) => {
  isAuraLocked = lock;
  if (lock) {
    document.documentElement.setAttribute('data-aura-locked', 'true');
  } else {
    document.documentElement.removeAttribute('data-aura-locked');
  }
};

/**
 * Met à jour les variables CSS globales pour l'aura et le contraste du texte.
 */
export function applyThemeColors(palette: string[], isVisible: boolean = true) {
  if (isAuraLocked) return;

  // Comparaison rapide pour éviter le travail inutile
  if (isVisible === lastVisibility && 
      palette.length === lastPalette.length && 
      palette.every((c, i) => c === lastPalette[i])) {
    return;
  }
  
  lastPalette = [...palette];
  lastVisibility = isVisible;

  const root = document.documentElement;
  const hasPalette = isVisible && palette?.length > 0;
  
  if (!hasPalette) {
    const defaultColor = "#050505";
    for (let i = 1; i <= 5; i++) root.style.setProperty(`--p${i}`, defaultColor);
    root.style.setProperty('--text-main', '#f0f0f0');
    root.style.setProperty('--text-muted', '#aaaaaa');
    root.style.setProperty('--accent', '#ffffff');
    root.style.setProperty('--aura-opacity', '0.25');
    root.style.setProperty('--bg-base', defaultColor);
    return;
  }

  const baseColor = palette[0] || "#050505";
  // Mise à jour des couleurs des blobs (max 5)
  for (let i = 0; i < 5; i++) {
    root.style.setProperty(`--p${i + 1}`, palette[i] || baseColor);
  }

  const avgLum = getAverageLuminance(palette);
  
  // Bascule dynamique du contraste
  if (avgLum > 0.38) {
    root.style.setProperty('--text-main', '#000000');
    root.style.setProperty('--text-muted', '#1a1a1a');
    root.style.setProperty('--accent', '#000000');
    root.style.setProperty('--aura-opacity', '0.5');
  } else {
    root.style.setProperty('--text-main', '#ffffff');
    root.style.setProperty('--text-muted', '#bbbbbb');
    root.style.setProperty('--accent', '#ffffff');
    root.style.setProperty('--aura-opacity', '0.35');
  }
  
  root.style.setProperty('--bg-base', '#050505');
}
