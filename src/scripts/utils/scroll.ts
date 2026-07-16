/**
 * Global utility for handling scroll animations and visibility-based interactions.
 */

/**
 * Default visibility threshold for scroll-based triggers.
 */
export const REVEAL_THRESHOLD = 0.08;

/**
 * Pre-trigger margin so elements reveal before they are fully in view —
 * critical when the user scrolls fast, otherwise the transition plays below
 * the fold and looks like the text is always missing.
 */
export const REVEAL_ROOT_MARGIN = '0px 0px 12% 0px';

// Kept at module scope so repeated `astro:page-load` calls reuse a single
// observer instead of stacking a new one on the same elements every navigation.
let revealObserver: IntersectionObserver | null = null;

/**
 * Initializes the reveal observer for `.reveal-on-scroll` elements.
 */
export function initScrollReveal() {
  // Drop the previous observer before creating a new one (soft navigations).
  revealObserver?.disconnect();

  const observerOptions = {
    root: null,
    rootMargin: REVEAL_ROOT_MARGIN,
    threshold: REVEAL_THRESHOLD
  };

  revealObserver = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('is-visible');
        // Stop observing once revealed, except for special triggers (e.g. auto-nav)
        if (entry.target.id !== 'next-roll-trigger') {
          revealObserver?.unobserve(entry.target);
        }
      } else if (entry.target.id === 'next-roll-trigger') {
        // Special case for the auto-navigation footer: reset if we scroll back up
        entry.target.classList.remove('is-visible', 'is-loading');
        document.body.classList.remove('is-transitioning');
      }
    });
  }, observerOptions);

  document.querySelectorAll('.reveal-on-scroll').forEach(el => revealObserver!.observe(el));

  return revealObserver;
}
