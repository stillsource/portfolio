// Small Pointer Events-based gesture helpers.
//
// Each helper attaches listeners to an element and returns a cleanup function.
// They assume the caller already applies `touch-action: none` (or similar)
// when capturing multi-touch gestures such as two-finger rotation.

export type SwipeDirection = 'up' | 'down' | 'left' | 'right';

export interface SwipeOptions {
  direction: SwipeDirection;
  threshold?: number; // minimum main-axis delta in px (default 80)
  maxOrthogonal?: number; // maximum off-axis delta in px (default 60)
  onSwipe(): void;
}

export function attachSwipe(el: HTMLElement, opts: SwipeOptions): () => void {
  const threshold = opts.threshold ?? 80;
  const maxOrthogonal = opts.maxOrthogonal ?? 60;
  let startX = 0;
  let startY = 0;
  let tracking = false;

  const onDown = (e: PointerEvent) => {
    if (!e.isPrimary) return;
    startX = e.clientX;
    startY = e.clientY;
    tracking = true;
  };

  const onUp = (e: PointerEvent) => {
    if (!tracking) return;
    tracking = false;
    const dx = e.clientX - startX;
    const dy = e.clientY - startY;

    switch (opts.direction) {
      case 'down':
        if (dy > threshold && Math.abs(dx) < maxOrthogonal) opts.onSwipe();
        break;
      case 'up':
        if (-dy > threshold && Math.abs(dx) < maxOrthogonal) opts.onSwipe();
        break;
      case 'right':
        if (dx > threshold && Math.abs(dy) < maxOrthogonal) opts.onSwipe();
        break;
      case 'left':
        if (-dx > threshold && Math.abs(dy) < maxOrthogonal) opts.onSwipe();
        break;
    }
  };

  el.addEventListener('pointerdown', onDown);
  el.addEventListener('pointerup', onUp);
  el.addEventListener('pointercancel', onUp);
  return () => {
    el.removeEventListener('pointerdown', onDown);
    el.removeEventListener('pointerup', onUp);
    el.removeEventListener('pointercancel', onUp);
  };
}

export interface RotateOptions {
  onRotate(angleDeg: number): void;
  onEnd?(): void;
}

// attachTwoFingerRotate tracks the angle delta between two pointers. It calls
// `onRotate` with the current rotation in degrees relative to the initial
// touchdown, and `onEnd` when the gesture completes.
export function attachTwoFingerRotate(el: HTMLElement, opts: RotateOptions): () => void {
  const pointers = new Map<number, { x: number; y: number }>();
  let startAngle = 0;

  const currentAngle = (): number => {
    const [a, b] = Array.from(pointers.values());
    return Math.atan2(b.y - a.y, b.x - a.x);
  };

  const onDown = (e: PointerEvent) => {
    pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
    if (pointers.size === 2) startAngle = currentAngle();
  };

  const onMove = (e: PointerEvent) => {
    if (!pointers.has(e.pointerId)) return;
    pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
    if (pointers.size === 2) {
      const delta = currentAngle() - startAngle;
      opts.onRotate(delta * (180 / Math.PI));
    }
  };

  const onUp = (e: PointerEvent) => {
    pointers.delete(e.pointerId);
    if (pointers.size < 2) opts.onEnd?.();
  };

  el.addEventListener('pointerdown', onDown);
  el.addEventListener('pointermove', onMove);
  el.addEventListener('pointerup', onUp);
  el.addEventListener('pointercancel', onUp);
  return () => {
    el.removeEventListener('pointerdown', onDown);
    el.removeEventListener('pointermove', onMove);
    el.removeEventListener('pointerup', onUp);
    el.removeEventListener('pointercancel', onUp);
  };
}
