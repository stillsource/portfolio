// Loupe: hold-to-magnify interaction for the lightbox image.
// Desktop-only (skipped for touch pointers). Mobile still relies on the
// existing pinch-zoom handler in LightboxGlobal.astro.

export interface LoupeStyle {
  inside: boolean;
  left?: number;
  top?: number;
  bgSize?: string;
  bgPosition?: string;
}

export interface ComputeLoupeStyleParams {
  rect: { left: number; top: number; right: number; bottom: number; width: number; height: number };
  clientX: number;
  clientY: number;
  zoom: number;
  diameter: number;
}

// Pure geometry: given the image rect, pointer position, zoom factor and
// loupe diameter, produce the CSS values for the lens. Kept side-effect-free
// so it can be unit-tested without a DOM.
export function computeLoupeStyle(params: ComputeLoupeStyleParams): LoupeStyle {
  const { rect, clientX, clientY, zoom, diameter } = params;
  if (
    clientX < rect.left || clientX > rect.right ||
    clientY < rect.top || clientY > rect.bottom ||
    rect.width === 0 || rect.height === 0
  ) {
    return { inside: false };
  }
  const rx = clientX - rect.left;
  const ry = clientY - rect.top;
  return {
    inside: true,
    left: clientX,
    top: clientY,
    bgSize: `${rect.width * zoom}px ${rect.height * zoom}px`,
    bgPosition: `${-(rx * zoom - diameter / 2)}px ${-(ry * zoom - diameter / 2)}px`,
  };
}

export interface AttachLoupeConfig {
  image: HTMLImageElement;
  loupe: HTMLElement;
  lightbox: HTMLElement;
  zoom?: number;
  onHoldEnd?: () => void;
}

export interface LoupeController {
  detach: () => void;
}

// Wires hold-to-magnify on an image element. Returns a controller so callers
// can detach listeners on teardown.
export function attachLoupe(config: AttachLoupeConfig): LoupeController {
  const { image, loupe, lightbox, zoom = 2.8, onHoldEnd } = config;
  let holding = false;

  const applyPosition = (cx: number, cy: number) => {
    const res = computeLoupeStyle({
      rect: image.getBoundingClientRect(),
      clientX: cx,
      clientY: cy,
      zoom,
      diameter: loupe.offsetWidth,
    });
    if (!res.inside) {
      loupe.classList.remove('is-active');
      return;
    }
    loupe.classList.add('is-active');
    loupe.style.left = `${res.left}px`;
    loupe.style.top = `${res.top}px`;
    loupe.style.backgroundSize = res.bgSize!;
    loupe.style.backgroundPosition = res.bgPosition!;
  };

  const onMove = (e: PointerEvent) => applyPosition(e.clientX, e.clientY);
  const onUp = () => stop();

  const stop = () => {
    if (!holding) return;
    holding = false;
    loupe.classList.remove('is-active');
    document.body.classList.remove('loupe-active');
    lightbox.classList.remove('loupe-active');
    window.removeEventListener('pointermove', onMove);
    window.removeEventListener('pointerup', onUp, true);
    window.removeEventListener('pointercancel', onUp, true);
    onHoldEnd?.();
  };

  const onDown = (e: PointerEvent) => {
    if (e.pointerType === 'touch' || e.button !== 0) return;
    holding = true;
    loupe.style.backgroundImage = `url("${image.src}")`;
    document.body.classList.add('loupe-active');
    lightbox.classList.add('loupe-active');
    applyPosition(e.clientX, e.clientY);
    window.addEventListener('pointermove', onMove, { passive: true });
    window.addEventListener('pointerup', onUp, true);
    window.addEventListener('pointercancel', onUp, true);
    e.preventDefault();
  };

  const onImageLoad = () => {
    if (!holding) loupe.style.backgroundImage = '';
  };

  image.addEventListener('pointerdown', onDown);
  image.addEventListener('load', onImageLoad);

  return {
    detach: () => {
      image.removeEventListener('pointerdown', onDown);
      image.removeEventListener('load', onImageLoad);
      stop();
    },
  };
}
