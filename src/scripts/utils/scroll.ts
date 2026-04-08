/**
 * Utilitaire global pour gérer les animations au scroll et les interactions basées sur la visibilité.
 */

/**
 * Initialise l'observateur de révélation pour les éléments .reveal-on-scroll.
 */
export function initScrollReveal() {
  const observerOptions = {
    root: null,
    rootMargin: '0px',
    threshold: 0.15
  };

  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('is-visible');
        // On arrête d'observer une fois révélé, sauf si c'est un trigger spécial (ex: auto-nav)
        if (entry.target.id !== 'next-roll-trigger') {
          observer.unobserve(entry.target);
        }
      } else if (entry.target.id === 'next-roll-trigger') {
        // Cas spécial pour le footer de navigation automatique : on reset si on remonte
        entry.target.classList.remove('is-visible', 'is-loading');
        document.body.classList.remove('is-transitioning');
      }
    });
  }, observerOptions);

  document.querySelectorAll('.reveal-on-scroll').forEach(el => observer.observe(el));
  
  return observer;
}

/**
 * Gère l'effet de parallaxe sur les fragments de pensées poétiques.
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
