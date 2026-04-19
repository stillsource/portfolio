/**
 * Global utility for handling scroll animations and visibility-based interactions.
 */

/**
 * Default visibility threshold for scroll-based triggers.
 */
export const REVEAL_THRESHOLD = 0.15;

/**
 * Initializes the reveal observer for `.reveal-on-scroll` elements.
 */
export function initScrollReveal() {
  const observerOptions = {
    root: null,
    rootMargin: '0px',
    threshold: REVEAL_THRESHOLD
  };

  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('is-visible');
        // Stop observing once revealed, except for special triggers (e.g. auto-nav)
        if (entry.target.id !== 'next-roll-trigger') {
          observer.unobserve(entry.target);
        }
      } else if (entry.target.id === 'next-roll-trigger') {
        // Special case for the auto-navigation footer: reset if we scroll back up
        entry.target.classList.remove('is-visible', 'is-loading');
        document.body.classList.remove('is-transitioning');
      }
    });
  }, observerOptions);

  document.querySelectorAll('.reveal-on-scroll').forEach(el => observer.observe(el));
  
  return observer;
}

/**
 * Handles the parallax effect on poetic thought fragments.
 */
export function initThoughtParallax() {
  const thoughts = document.querySelectorAll('.thought-fragment');
  if (thoughts.length === 0) return null;

  const handleScroll = () => {
    const vh = window.innerHeight;
    
    thoughts.forEach(thought => {
      const rect = thought.getBoundingClientRect();
      const elementCenter = rect.top + rect.height / 2;
      const distanceFromCenter = (elementCenter - vh / 2) / (vh / 2);
      
      if (rect.top < vh && rect.bottom > 0) {
        const yOffset = distanceFromCenter * -80; 
        const opacity = 1 - Math.abs(distanceFromCenter) * 1.8;
        
        (thought as HTMLElement).style.transform = `translate3d(0, ${yOffset}px, 0)`;
        (thought as HTMLElement).style.opacity = Math.max(0, Math.min(0.6, opacity)).toString();
      }
    });
  };

  window.addEventListener('scroll', handleScroll, { passive: true });
  handleScroll(); // Initial check

  return () => window.removeEventListener('scroll', handleScroll);
}
