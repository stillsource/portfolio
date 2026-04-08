# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
npm install          # Install dependencies
npm run dev          # Start dev server (uses local content only, no kDrive sync)
npm run sync         # Fetch from kDrive and generate content files
npm run build        # sync + astro build (full production build)
npm run preview      # Preview production build
```

> `npm run dev` works without `.env` — it serves whatever content is already in `src/content/rolls/synced/`. Only `npm run sync` and `npm run build` require kDrive credentials.

No test suite exists in this project.

## Required Environment Variables

Create a `.env` file at the root:
```
KDRIVE_API_TOKEN=
KDRIVE_DRIVE_ID=
KDRIVE_FOLDER_ID=   # ID of the root folder containing Roll sub-folders
```

## Architecture

### Data Flow

```
kDrive API
  └─ src/scripts/fetch-kdrive.ts  (npm run sync)
       ├─ fetches folder list, downloads images for EXIF + palette extraction
       ├─ reads .md poetry files inside kDrive folders (or src/data/poetry/)
       └─ writes → src/content/rolls/synced/*.md  (one file per Roll)
                 → public/search-index.json

Astro build
  └─ src/content.config.ts  (defines "rolls" collection schema)
       └─ src/pages/index.astro          (lists all Rolls)
       └─ src/pages/roll/[slug].astro    (individual Roll page)
       └─ src/pages/tags/[tag].astro     (tag filter pages)
```

Each synced `.md` file has frontmatter with: `title`, `date`, `tags[]`, `poem`, `palette[]` (hex colors), `dominantColor`, `audioUrl`, and an `images[]` array (url, metadata/EXIF string, poem, palette, dominantColor per photo).

### Aura System (Ambient Color Engine)

The "Aura 2.0" system drives the entire visual identity. It works through:

1. **CSS custom properties** (`--p1` through `--p5`, `--bg-base`, `--text-main`, `--text-muted`, `--accent`) defined globally in `src/layouts/Layout.astro` with `@property` declarations (CSS Houdini) to enable smooth color transitions.
2. **Five `.aura-blob` divs** with `transition:persist="aura-container"` — they survive Astro page navigations and are colored by the palette vars.
3. **`src/scripts/utils/colors.ts`** — the client-side engine. `applyThemeColors(palette, isVisible)` updates the CSS vars; `setAuraLock(bool)` pauses transitions during navigation (the "Color-Lock" mechanism) to prevent flickering. The lock is set on `astro:before-preparation` and released on `astro:page-load`.
4. **Scroll-based triggers** — in `Roll.astro`, each `image-wrapper` has `data-palette` and uses an `IntersectionObserver` so the aura updates as you scroll through photos.
5. **Hover triggers** — on the index, hovering a roll link previews its palette.

### Component Responsibilities

- **`Layout.astro`** — global shell: `<head>`, aura blobs, noise overlay, custom cursor, global lightbox DOM, global keyboard shortcuts (`h`/`.` for "chambre noire" mode, `Escape` to close lightbox), Astro `ClientRouter`.
- **`Roll.astro`** — renders a single photo series: three-column layout (left poem | center image | right poem), scroll parallax on `.thought-fragment` elements, click → `open-lightbox` custom event.
- **`Navbar.astro`** — floating glassmorphism nav bar, `AudioPlayer` integration, search toggle.
- **`Search.astro`** — client-side search against `public/search-index.json` (fetched on first open).
- **`SplashScreen.astro`** — cinematic intro on first visit (uses `sessionStorage.splashSeen`). Fires `splashFinished` window event when done so index reveals its content.
- **`AudioPlayer.astro`** — ambient sound player for `public/ambiance-rue.mp3` or per-roll `audioUrl`.
- **`CustomCursor.astro`** — custom dot cursor, hidden via CSS when `prefers-reduced-motion`.

### View Transitions & Persistence

Astro's `ClientRouter` is active. The aura container has `transition:persist` so it never re-mounts between navigations. Global event listeners (lightbox, keyboard, aura-lock) are guarded by `window.__aura_initialized` and `window.__reveal_initialized` flags to prevent duplicate binding across navigations. All page-specific JS re-runs on `astro:page-load`.

### Content Schema

`src/content.config.ts` — the `rolls` collection uses Astro's `glob` loader on `src/content/rolls/**/*.md`. The schema is defined with Zod. The `images` array is fully typed with optional per-photo `poem`, `palette`, and `dominantColor`.

### Tag Categories

Tags come from IPTC/XMP metadata on photos. `src/data/categories.ts` maps known tags to display categories (Ambiance, Lieu, Technique, Sujet). Unknown tags default to "Sujet".
