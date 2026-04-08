# Marvelous journal of a wanderer - Portfolio de Photographie de Rue

Un portfolio photographique minimaliste, performant et artistique, conçu avec **Astro**.
Il repose sur une architecture "Zéro Stockage" (les images restent sur votre kDrive) et offre une expérience utilisateur premium (animations cinématiques, fond coloré adaptatif, paysage sonore).

## 🚀 Démarrage rapide (Local)

1. Clonez ce dépôt.
2. Installez les dépendances : `npm install`
3. Créez un fichier `.env` à la racine (voir *Configuration kDrive* ci-dessous).
4. Lancez le serveur de développement : `npm run dev` (Astro affichera les photos locales de test si l'API n'est pas configurée).

## ☁️ Configuration kDrive (La source des images)

Ce site ne stocke **aucune** photo lourde. Il génère des URLs vers vos fichiers kDrive.

1. Allez sur votre console **Infomaniak** et générez un Token API kDrive.
2. Repérez l'ID de votre Drive et l'ID du dossier racine contenant vos "Rolls" (balades).
3. Renommez le fichier `.env.example` en `.env` et remplissez vos identifiants :
   ```env
   KDRIVE_API_TOKEN=votre_cle_api_ici
   KDRIVE_DRIVE_ID=votre_drive_id_ici
   KDRIVE_FOLDER_ID=votre_folder_id_racine_ici
   ```

**Structure attendue sur kDrive :**
- `Dossier Racine (KDRIVE_FOLDER_ID)`
  - `Balade à Paris` (Ce sera le titre du Roll)
    - `photo1.jpg`
    - `photo2.jpg`
  - `Néons de Nuit`
    - `DSCF1234.jpg`

## 🖋️ Ajouter de la Poésie

Vous pouvez associer des textes à vos sessions (Rolls) ou à des photos spécifiques.
Allez dans le dossier `src/data/poetry/` et créez un fichier `.md` portant le nom "slugifié" de votre dossier kDrive (ex: si le dossier s'appelle "Balade à Paris", créez `balade-a-paris.md`).

**Syntaxe du fichier poétique :**
```markdown
---
photos:
  "DSCF1234.jpg": |
    Les néons ne pleurent jamais,
    Ils saignent de la lumière.
---
Ceci est le poème global qui s'affichera sous le titre de la balade.
```
*Ces textes survivront aux synchronisations kDrive et s'afficheront au "clic/survol" des images sur le site.*

## 🎵 Ajouter l'ambiance sonore

1. Trouvez un fichier audio libre de droits (ex: bruit de ville, pluie, drone musical) au format `.mp3`.
2. Nommez-le **`ambiance-rue.mp3`**.
3. Placez-le dans le dossier `public/` à la racine du projet.
4. L'icône de son dans le menu flottant permettra aux visiteurs de l'activer/désactiver.

## 🛠️ Le Processus de Build (Déploiement)

Lorsque vous (ou votre hébergeur comme Vercel/Netlify) lancez la commande de génération du site :
```bash
npm run build
```
Voici ce qu'il se passe automatiquement :
1. Le script `fetch-kdrive.ts` se connecte à Infomaniak.
2. Il scanne vos dossiers, extrait les **mots-clés (tags)** et les **données de l'appareil (EXIF)** de chaque photo.
3. Il lit vos poèmes locaux (`src/data/poetry/`).
4. Il génère les données pour Astro et Astro compile un site HTML/CSS ultra-rapide.

## 🎨 Fonctionnalités de Design & Immersion
- **Système Aura 2.0 :** Un moteur d'ambiance dynamique qui extrait les couleurs des photos pour créer un éclairage immersif fluide et persistant.
- **Performance GPU :** Animations optimisées via `translate3d` et `will-change` pour un rendu à 60fps constant.
- **Transition:Persist :** L'interface (Aura, Navigation, Audio) persiste entre les pages pour une expérience sans coupure.
- **Splash Screen :** Animation cinématique d'introduction à la première visite.
- **Glassmorphism :** Menu flottant et overlays de métadonnées en verre dépoli (20px blur).
- **View Transitions :** Navigation fluide via Astro ClientRouter avec Color-Lock matériel.
- **Grain global :** Texture SVG dynamique imitant le grain de pellicule cinématographique.

## 🛠️ Optimisations Techniques (Pickle Rick Edition)
- **Zéro Fuite Mémoire :** Délégation d'événements et nettoyage automatique des listeners sur tout le projet.
- **Cache de Palette :** Évite les manipulations DOM redondantes pour un gain de CPU immédiat.
- **Race Condition Zero :** Synchronisation robuste des timers de navigation et des transitions d'état.

