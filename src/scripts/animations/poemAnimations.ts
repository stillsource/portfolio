// Registry of reveal animations applied to `.thought-fragment` poem elements
// inside Roll.astro. Each strategy controls its own DOM preparation, the
// transition fired when the element enters the viewport, and optional cleanup
// when it leaves. Keeping them in one file makes it easy to add a new strategy
// without touching the observer wiring.

export type PoemAnimationName = 'typewriter' | 'fade' | 'slide' | 'word' | 'blur';

export interface PoemAnimation {
  name: PoemAnimationName;
  // Called once per element before the observer starts.
  prepare(el: HTMLElement, origText: string): void;
  // Called whenever the element enters the viewport. May return a cleanup
  // function (e.g. to cancel a running setInterval).
  onEnter(el: HTMLElement, origText: string): (() => void) | void;
  // Called when the element leaves the viewport. Should undo onEnter side
  // effects so re-entering plays the animation again.
  onLeave(el: HTMLElement, origText: string): void;
}

const simpleEnter: PoemAnimation['onEnter'] = (el) => {
  el.classList.add('is-visible');
};

const simpleLeave: PoemAnimation['onLeave'] = (el) => {
  el.classList.remove('is-visible');
};

const fade: PoemAnimation = {
  name: 'fade',
  prepare(el) {
    el.classList.add('anim-fade');
  },
  onEnter: simpleEnter,
  onLeave: simpleLeave,
};

const slide: PoemAnimation = {
  name: 'slide',
  prepare(el) {
    el.classList.add('anim-slide');
  },
  onEnter: simpleEnter,
  onLeave: simpleLeave,
};

const blur: PoemAnimation = {
  name: 'blur',
  prepare(el) {
    el.classList.add('anim-blur');
  },
  onEnter: simpleEnter,
  onLeave: simpleLeave,
};

const word: PoemAnimation = {
  name: 'word',
  prepare(el, origText) {
    el.classList.add('anim-word');
    // Split text into word <span>s so CSS can stagger their opacity via --word-i.
    const words = origText.split(' ');
    el.textContent = '';
    words.forEach((w, i) => {
      if (i > 0) el.appendChild(document.createTextNode(' '));
      const span = document.createElement('span');
      span.className = 'poem-word';
      span.style.setProperty('--word-i', String(i));
      span.textContent = w;
      el.appendChild(span);
    });
  },
  onEnter: simpleEnter,
  onLeave: simpleLeave,
};

const typewriter: PoemAnimation = {
  name: 'typewriter',
  prepare(el) {
    el.classList.add('anim-typewriter');
    el.textContent = '';
    el.style.opacity = '0.75';
  },
  onEnter(el, origText) {
    el.textContent = '';
    let i = 0;
    const id = setInterval(() => {
      el.textContent = origText.slice(0, ++i);
      if (i >= origText.length) clearInterval(id);
    }, 40);
    return () => clearInterval(id);
  },
  onLeave(el) {
    el.classList.remove('is-visible');
    el.textContent = '';
  },
};

export const poemAnimations: Record<PoemAnimationName, PoemAnimation> = {
  fade, slide, blur, word, typewriter,
};

export const POEM_ANIMATION_POOL: readonly PoemAnimationName[] = [
  'typewriter', 'fade', 'slide', 'word', 'blur',
];

// pickAnimation resolves the requested mode to a concrete strategy. "random"
// picks from the pool; unknown modes fall back to `fade`.
export function pickAnimation(mode: string): PoemAnimation {
  if (mode === 'random') {
    const name = POEM_ANIMATION_POOL[Math.floor(Math.random() * POEM_ANIMATION_POOL.length)];
    return poemAnimations[name];
  }
  return poemAnimations[mode as PoemAnimationName] ?? poemAnimations.fade;
}
