# Plan d'Amélioration Visuelle - Système d'Aura et Métadonnées

Ce plan vise à synchroniser parfaitement le système de couleurs immersives (Aura) et les métadonnées photographiques entre la source kDrive et l'interface utilisateur.

## Objectif
- S'assurer que chaque "Roll" et chaque photo disposent d'une couleur dominante extraite pour alimenter l'Aura.
- Intégrer les données EXIF (vitesse, ouverture, boîtier) de manière robuste.

## Fichiers concernés
- `src/content.config.ts` : Mise à jour du schéma Astro Content.
- `src/scripts/fetch-kdrive.ts` : Extraction et sauvegarde de la couleur dominante.
- `src/pages/roll/[slug].astro` : Transmission des données enrichies aux composants.

## Étapes d'implémentation

### 1. Mise à jour du Schéma Astro
Ajouter le champ `dominantColor` (string) au niveau du Roll et de chaque image dans `src/content.config.ts`.

### 2. Enrichissement du Script de Synchronisation
Dans `src/scripts/fetch-kdrive.ts` :
- Calculer `dominantColor` à partir de la palette (premier élément de l'array `palette`).
- Sauvegarder ce champ dans le frontmatter Markdown généré pour chaque Roll.
- S'assurer que chaque image dans l'array `images` possède aussi sa `dominantColor`.

### 3. Vérification des Composants
- Vérifier que `Roll.astro` utilise bien `img.dominantColor` si `img.palette` est absent.
- S'assurer que l'Aura réagit correctement lors du scroll entre les images.

## Vérification & Tests
- Lancer `npm run sync` (si possible) ou vérifier la structure des fichiers `.md` générés.
- Vérifier que les types TypeScript sont correctement générés par Astro (`astro dev`).
- Observer les transitions de couleurs dans la console du navigateur (via le script `colors.ts`).
