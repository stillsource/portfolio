# Marvelous Journal of a Wanderer - Engineering Guide

Ce document sert de guide de bord pour le développement et la maintenance du portfolio photographique. Il documente les choix architecturaux et les spécificités techniques du projet.

## 1. Philosophie "Zéro Stockage"
Le projet est conçu pour ne stocker aucune image localement (hormis le favicon). 
- **Source** : Les photos sont hébergées sur un kDrive (Infomaniak).
- **Synchronisation** : Le script `src/scripts/fetch-kdrive.ts` récupère les URLs publiques et les métadonnées (EXIF, Tags, Palette).
- **Rendu** : Astro génère des pages statiques à partir de fichiers Markdown situés dans `src/content/rolls/`.

## 2. Système "Aura" (Ambiance Dynamique)
Le site utilise un système de couleurs immersives qui s'adapte au contenu.
- **Extraction** : Via `node-vibrant`, nous extrayons une palette de 5 couleurs par image.
- **Affichage** : Le `Layout.astro` contient 5 sphères de lumière (`.aura-blob`) dont les couleurs (`--p1` à `--p5`) sont mises à jour via JavaScript lors du scroll ou du survol.
- **Contraste** : La couleur du texte (`--text-main`, `--text-muted`) bascule automatiquement du blanc au noir si la luminance moyenne de la palette dépasse 0.38.

## 3. Conventions de Développement
- **Styles** : Utilisation de variables CSS et de l'API `@property` pour des transitions fluides. Les styles globaux de l'index sont dans `is:global` pour supporter les `ViewTransitions`.
- **Navigation** : Les transitions entre pages sont gérées par `ClientRouter`. L'audio utilise `transition:persist` pour ne pas être interrompu.
- **Images** : Utilisation de la balise `<img>` native avec `loading="lazy"` et `decoding="async"` pour les images distantes afin d'éviter les erreurs de calcul de dimensions côté serveur.

## 4. Ajout de Contenu
1. Créer un dossier sur kDrive avec le nom de la balise (ex: "Nuit à Tokyo").
2. Y placer les photos JPEG (taguées avec vos mots-clés dans Lightroom).
3. Optionnel : Créer un fichier `src/data/poetry/[slug].md` pour ajouter des textes poétiques par photo.
4. Lancer `npm run sync` pour générer les fichiers de contenu Astro.

## 5. Maintenance
- **Scripts** : `src/scripts/utils/` contient les fonctions partagées pour les calculs de couleurs.
- **Catégories** : Modifiables dans `src/data/categories.ts`.
