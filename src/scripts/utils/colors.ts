/**
 * Calcule la luminance moyenne d'une palette de couleurs hexadécimales.
 */
export function getAverageLuminance(palette: string[]): number {
  if (!palette || palette.length === 0) return 0;
  const lums = palette.map(hex => {
    // Nettoyage au cas où le hex contient des caractères invisibles ou n'a pas le #
    const cleanHex = hex.startsWith('#') ? hex.slice(1) : hex;
    const r = parseInt(cleanHex.slice(0, 2), 16);
    const g = parseInt(cleanHex.slice(2, 4), 16);
    const b = parseInt(cleanHex.slice(4, 6), 16);
    return (0.299 * r + 0.587 * g + 0.114 * b) / 255;
  });
  return lums.reduce((a, b) => a + b, 0) / lums.length;
}

/**
 * Met à jour les variables CSS globales pour l'aura et le contraste du texte.
 */
export function applyThemeColors(palette: string[], isVisible: boolean = true) {
  const root = document.documentElement;
  
  if (!isVisible || !palette || palette.length === 0) {
    // Reset par défaut (Mode Sombre)
    for (let i = 1; i <= 5; i++) root.style.setProperty(`--p${i}`, "#050505");
    root.style.setProperty('--bg-base', '#050505');
    root.style.setProperty('--text-main', '#f0f0f0');
    root.style.setProperty('--text-muted', '#aaaaaa');
    root.style.setProperty('--accent', '#ffffff');
    root.style.setProperty('--aura-opacity', '0.25');
    return;
  }

  // Appliquer la palette aux blobs
  const baseColor = palette[0] || "#050505";
  for (let i = 0; i < 5; i++) {
    root.style.setProperty(`--p${i+1}`, palette[i] || baseColor);
  }

  // Calcul du contraste
  const avgLum = getAverageLuminance(palette);
  
  if (avgLum > 0.38) {
    // Mode Clair Dynamique
    root.style.setProperty('--text-main', '#000000');
    root.style.setProperty('--text-muted', '#1a1a1a');
    root.style.setProperty('--accent', '#000000');
    root.style.setProperty('--aura-opacity', '0.5');
  } else {
    // Mode Sombre Dynamique
    root.style.setProperty('--text-main', '#ffffff');
    root.style.setProperty('--text-muted', '#bbbbbb');
    root.style.setProperty('--accent', '#ffffff');
    root.style.setProperty('--aura-opacity', '0.35');
  }
  
  root.style.setProperty('--bg-base', '#050505');
}
