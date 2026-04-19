// One-time per-tab initialisation helper.
// Astro's ClientRouter fires `astro:page-load` on every soft navigation, so any
// code that should only run once for the tab (window-level listeners, shared
// audio context, etc.) must gate itself on a persistent window flag.
// This helper centralises that pattern so the flags live behind a single API.

export function oncePerTab(key: string, fn: () => void): void {
  const w = window as unknown as Record<string, unknown>;
  if (w[key]) return;
  w[key] = true;
  fn();
}
