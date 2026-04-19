// Thin wrapper around the Vibration API. All helpers short-circuit when the
// device does not support vibration, when the user has no touch capability,
// or when prefers-reduced-motion is active — so call sites do not need to
// repeat those guards.

const hasVibrate =
  typeof navigator !== 'undefined' && typeof navigator.vibrate === 'function';

const isTouch =
  typeof navigator !== 'undefined' && (navigator.maxTouchPoints ?? 0) > 0;

function isReducedMotion(): boolean {
  if (typeof window === 'undefined') return false;
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

export function vibrate(pattern: number | number[]): void {
  if (!hasVibrate || !isTouch || isReducedMotion()) return;
  try {
    navigator.vibrate(pattern);
  } catch {
    /* ignore */
  }
}

// Named semantic presets. Prefer these over raw numbers at call sites.
export const haptics = {
  tap:     () => vibrate(8),              // photo click
  swipe:   () => vibrate([10, 30, 10]),   // confirmed swipe
  success: () => vibrate([5, 10, 20]),    // validated action
};
