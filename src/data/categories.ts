/**
 * Dictionnaire de classification des tags pour le filtrage sur l'index.
 */
export const CATEGORIES = {
  "Ambiance": ["Nocturne", "Matin", "Brume", "Golden Hour", "Froid", "Bleu", "Mélancolie", "Chaud", "Doux"],
  "Lieu": ["Tokyo", "Paris", "Urbain", "Rue", "Ville", "Nature"],
  "Technique": ["Reflets", "Contraste", "Silhouette", "Pluie", "Flou", "Argentique", "Grain", "Long Exposure"],
  "Sujet": ["Architecture", "Néon", "Minimalisme", "Humain", "Géométrie", "Ombres", "Lumière"]
} as const;

/**
 * Retourne la catégorie correspondante à un tag donné.
 */
export function getTagCategory(tag: string): string {
  for (const [category, tags] of Object.entries(CATEGORIES)) {
    if ((tags as readonly string[]).includes(tag)) return category;
  }
  return "Sujet"; // Catégorie par défaut
}
