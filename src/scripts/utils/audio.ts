// Shared Web Audio helpers used by the global lightbox, the film-advance
// effect on navigation, and the per-image sonic design triggers.
//
// All players early-return when `prefers-reduced-motion: reduce` is active
// or when the user has muted synth sounds through the AudioPlayer toggle.

const MUTE_STORAGE_KEY = 'audio-muted-synth';

let soundCtx: AudioContext | null = null;

export function getSoundCtx(): AudioContext | null {
  if (typeof window === 'undefined') return null;
  if (!soundCtx) {
    try {
      soundCtx = new AudioContext();
    } catch {
      return null;
    }
  }
  return soundCtx;
}

export function isReducedMotion(): boolean {
  if (typeof window === 'undefined') return false;
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

export function isSynthMuted(): boolean {
  if (typeof window === 'undefined') return false;
  try {
    return sessionStorage.getItem(MUTE_STORAGE_KEY) === '1';
  } catch {
    return false;
  }
}

export function setSynthMuted(muted: boolean): void {
  try {
    if (muted) sessionStorage.setItem(MUTE_STORAGE_KEY, '1');
    else sessionStorage.removeItem(MUTE_STORAGE_KEY);
  } catch {
    /* ignore */
  }
}

function shouldSkip(): boolean {
  return isReducedMotion() || isSynthMuted();
}

// playShutter emits a short noise burst shaped like a film-camera shutter.
// `pitch` scales the high-pass cutoff (default 1 = 600 Hz); `volume`
// overrides the peak gain (default 0.15).
export function playShutter(opts?: { pitch?: number; volume?: number }): void {
  if (shouldSkip()) return;
  const ac = getSoundCtx();
  if (!ac || ac.state === 'suspended') return;

  try {
    const bufLen = Math.floor(ac.sampleRate * 0.055);
    const buf = ac.createBuffer(1, bufLen, ac.sampleRate);
    const d = buf.getChannelData(0);
    for (let i = 0; i < bufLen; i++) d[i] = (Math.random() * 2 - 1) * ((1 - i / bufLen) ** 3);
    const src = ac.createBufferSource();
    src.buffer = buf;
    const g = ac.createGain();
    g.gain.setValueAtTime(opts?.volume ?? 0.15, ac.currentTime);
    g.gain.exponentialRampToValueAtTime(0.001, ac.currentTime + 0.055);
    const f = ac.createBiquadFilter();
    f.type = 'highpass';
    f.frequency.value = 600 * (opts?.pitch ?? 1);
    src.connect(f); f.connect(g); g.connect(ac.destination);
    src.start();
  } catch {
    /* ignore */
  }
}

// playFilmAdvance emits three quick transients, like film being advanced in a
// manual camera.
export function playFilmAdvance(): void {
  if (shouldSkip()) return;
  const ac = getSoundCtx();
  if (!ac || ac.state === 'suspended') return;

  try {
    [0, 0.045, 0.09].forEach((delay) => {
      const bufLen = Math.floor(ac.sampleRate * 0.02);
      const buf = ac.createBuffer(1, bufLen, ac.sampleRate);
      const d = buf.getChannelData(0);
      for (let i = 0; i < bufLen; i++) d[i] = (Math.random() * 2 - 1) * ((1 - i / bufLen) ** 4);
      const src = ac.createBufferSource();
      src.buffer = buf;
      const g = ac.createGain();
      g.gain.setValueAtTime(0.06, ac.currentTime + delay);
      g.gain.exponentialRampToValueAtTime(0.001, ac.currentTime + delay + 0.02);
      src.connect(g); g.connect(ac.destination);
      src.start(ac.currentTime + delay);
    });
  } catch {
    /* ignore */
  }
}

// playViewportTone emits a soft sine bell tied to an image's position in the
// roll. Pitch climbs progressively (≈ one semitone per image) so scrolling
// through a series feels like a quiet melodic line.
export function playViewportTone(index: number, _total: number): void {
  if (shouldSkip()) return;
  const ac = getSoundCtx();
  if (!ac || ac.state === 'suspended') return;

  try {
    const baseFreq = 180;
    const freq = baseFreq * Math.pow(1.06, index);

    const osc = ac.createOscillator();
    osc.type = 'sine';
    osc.frequency.value = freq;

    const g = ac.createGain();
    g.gain.setValueAtTime(0, ac.currentTime);
    g.gain.linearRampToValueAtTime(0.04, ac.currentTime + 0.2);
    g.gain.exponentialRampToValueAtTime(0.001, ac.currentTime + 1.2);

    const f = ac.createBiquadFilter();
    f.type = 'lowpass';
    f.frequency.value = 800;

    osc.connect(f); f.connect(g); g.connect(ac.destination);
    osc.start();
    osc.stop(ac.currentTime + 1.3);
  } catch {
    /* ignore */
  }
}
