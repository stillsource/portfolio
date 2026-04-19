# Marvelous journal of a wanderer - Street Photography Portfolio

A minimalist, performant and artistic photo portfolio, built with **Astro**.
It is based on a "Zero Storage" architecture (images stay on your kDrive) and delivers a premium user experience (cinematic animations, adaptive colored background, soundscape).

## Quick start (Local)

1. Clone this repository.
2. Install dependencies: `npm install`
3. Create a `.env` file at the root (see *kDrive configuration* below).
4. Start the dev server: `npm run dev` (Astro will display local test photos if the API is not configured).

## kDrive configuration (The image source)

This site stores **no** heavy photos. It generates URLs pointing to your kDrive files.

1. Go to your **Infomaniak** console and generate a kDrive API token.
2. Locate your Drive ID and the ID of the root folder containing your "Rolls" (walks).
3. Rename the `.env.example` file to `.env` and fill in your credentials:
   ```env
   KDRIVE_API_TOKEN=your_api_key_here
   KDRIVE_DRIVE_ID=your_drive_id_here
   KDRIVE_FOLDER_ID=your_root_folder_id_here
   ```

**Expected kDrive structure:**
- `Root folder (KDRIVE_FOLDER_ID)`
  - `Walk in Paris` (This will be the Roll title)
    - `photo1.jpg`
    - `photo2.jpg`
  - `Night neons`
    - `DSCF1234.jpg`

## Adding poetry

You can associate texts with your sessions (Rolls) or with specific photos.
Go to the `src/data/poetry/` folder and create a `.md` file with the "slugified" name of your kDrive folder (e.g. if the folder is called "Walk in Paris", create `walk-in-paris.md`).

**Poetry file syntax:**
```markdown
---
photos:
  "DSCF1234.jpg": |
    Les néons ne pleurent jamais,
    Ils saignent de la lumière.
---
This is the global poem that will be displayed under the walk title.
```
*These texts will survive kDrive synchronizations and will be shown on image click/hover on the site.*

## Adding ambient audio

1. Find a royalty-free audio file (e.g. city noise, rain, musical drone) in `.mp3` format.
2. Name it **`ambiance-rue.mp3`**.
3. Place it in the `public/` folder at the project root.
4. The sound icon in the floating menu lets visitors toggle it on/off.

## The build process (Deployment)

When you (or your host such as Vercel/Netlify) run the site generation command:
```bash
npm run build
```
Here's what happens automatically:
1. The Go program `kdrive-sync` connects to Infomaniak.
2. It scans your folders, extracts **keywords (tags)** and **camera data (EXIF)** from each photo.
3. It reads your local poems (`src/data/poetry/`).
4. It generates the data for Astro and Astro compiles an ultra-fast HTML/CSS site.

## Design & immersion features
- **Aura 2.0 system:** A dynamic ambience engine that extracts colors from photos to create smooth, persistent immersive lighting.
- **GPU performance:** Animations optimized via `translate3d` and `will-change` for a consistent 60fps rendering.
- **Transition:Persist:** The interface (Aura, Navigation, Audio) persists across pages for a seamless experience.
- **Splash Screen:** Cinematic intro animation on the first visit.
- **Glassmorphism:** Floating menu and metadata overlays in frosted glass (20px blur).
- **View Transitions:** Smooth navigation via Astro ClientRouter with hardware Color-Lock.
- **Global grain:** Dynamic SVG texture mimicking cinematic film grain.

## Technical optimizations (Pickle Rick Edition)
- **Zero memory leaks:** Event delegation and automatic listener cleanup across the whole project.
- **Palette cache:** Avoids redundant DOM manipulations for an immediate CPU gain.
- **Race condition zero:** Robust synchronization of navigation timers and state transitions.
