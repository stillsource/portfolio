# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
npm install               # Install dependencies
npm run dev               # Start dev server (uses committed content, no kDrive sync)
npm run sync              # Fetch from kDrive and regenerate content files
npm run build             # Astro build only (uses existing content)
npm run build:full        # sync + astro build (full production build)
npm run preview           # Preview production build
npm run test:unit         # Vitest unit tests
npm run test:unit:watch   # Vitest watch mode
npx playwright test       # End-to-end suite (chromium + Pixel 7 mobile project)
```

> `npm run dev` works without `.env` — it serves whatever content is already in `src/content/rolls/`. Only `npm run sync` and `npm run build:full` require kDrive credentials.

## Frontend tests

- **Vitest** (`vitest.config.ts`): unit specs under `tests/unit/` — currently `formatExif.test.ts` and `loupe.test.ts`.
- **Playwright** (`playwright.config.ts`): e2e specs under `tests/e2e/` — `aura`, `chambre-noire`, `cursor`, `gestures` (mobile-only), `immersive-features`, `index`, `lightbox`, `performance`, `roll`, `search`, `sonic`, `tags`, `theme`, `view-transitions`. The config's `webServer` runs `node scripts/generate-search-index.mjs && npm run dev` so the search index is pre-populated without a full kDrive sync.

## Go Sync Script

The kdrive-sync binary lives in `src/scripts/go/kdrive-sync/`. It uses `go run` (no compilation step needed). A second `cmd/extractpalette` CLI exposes the palette engine on its own for one-off runs.

```bash
# From repo root:
go test -C src/scripts/go/kdrive-sync ./...          # Run Go tests
golangci-lint run -C src/scripts/go/kdrive-sync      # Lint (uses .golangci.yml)
# Or from src/scripts/go/kdrive-sync/ :
make test              # go test ./...
make test-race         # race detector
make test-coverage     # coverage.html + total %
make test-watch        # ginkgo watch -r
```

Architecture: `cmd/` → `pkg/usecase/` → `pkg/infrastructure/*` → `pkg/domain/`

Tests: table-driven (`t.Run`) for pure packages (`pkg/domain/slug`, `paletteaggregator`, `poetryparser`, `rollwriter`, `searchindexwriter`, `imageanalyzer`); Ginkgo v2 + Gomega BDD for orchestration / HTTP (`pkg/usecase`, `pkg/infrastructure/kdriveapi`, `filelister`, `filedownloader`, `sharepublisher`). Shared fakes live in `pkg/service/servicefakes/` (thread-safe, Stub + Results + Calls pattern). godog + go-rod e2e suite at `src/scripts/go/kdrive-sync/e2e/` drives the Astro frontend and is excluded from the unit CI run (`go list ./... | grep -v /e2e`).

CI gate: `go test -race -coverprofile -coverpkg=./pkg/...` with a ≥ 70% total threshold (`.github/workflows/ci.yml`).

## CI & deployment

- `.github/workflows/ci.yml` — Go unit tests (race + coverage ≥ 70 %), `astro check`, Playwright e2e.
- `.github/workflows/lighthouse.yml` + `lighthouserc.json` — 3-run Lighthouse budget on preview builds.
- `vercel.json` — responsive image pipeline config (AVIF/WebP, cache TTL).
- `@vite-pwa/astro` + `@astrojs/sitemap` + `@vercel/analytics` wired through `astro.config.mjs` and `Layout.astro`.

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
  └─ src/scripts/go/kdrive-sync/  (npm run sync — Go binary)
       ├─ fetches folder list, downloads images for EXIF + palette extraction
       ├─ reads .md poetry files inside kDrive folders
       └─ writes → src/content/rolls/synced/*.md  (one file per Roll)
                 → public/search-index.json

Astro build
  └─ src/content.config.ts  (defines "rolls" collection schema)
       └─ src/pages/index.astro          (lists all Rolls)
       └─ src/pages/roll/[slug].astro    (individual Roll page)
       └─ src/pages/tags/[tag].astro     (tag filter pages)
       └─ src/pages/404.astro            (themed not-found page)
```

Each synced `.md` file has frontmatter with: `title`, `date`, `tags[]`, `poem`, `palette[]` (hex colors), `dominantColor`, `audioUrl`, and an `images[]` array (url, exif object, poem, palette, dominantColor per photo).

### Aura System (Ambient Color Engine)

The "Aura 2.0" system drives the entire visual identity. It works through:

1. **CSS custom properties** (`--p1` through `--p5`, `--bg-base`, `--text-main`, `--text-muted`, `--accent`) defined globally in `src/layouts/Layout.astro` with `@property` declarations (CSS Houdini) to enable smooth color transitions.
2. **Five `.aura-blob` divs** with `transition:persist="aura-container"` — they survive Astro page navigations and are colored by the palette vars.
3. **`src/scripts/utils/colors.ts`** — the client-side engine. `applyThemeColors(palette, isVisible)` updates the CSS vars; `setAuraLock(bool)` pauses transitions during navigation (the "Color-Lock" mechanism) to prevent flickering. The lock is set on `astro:before-preparation` and released on `astro:page-load`.
4. **Scroll-based triggers** — in `Roll.astro`, each `image-wrapper` has `data-palette` and uses an `IntersectionObserver` so the aura updates as you scroll through photos.
5. **Hover triggers** — on the index, hovering a roll link previews its palette.

### Component Responsibilities

- **`Layout.astro`** — global shell: `<head>`, aura blobs, noise overlay, custom cursor, global lightbox DOM, dust canvas, global keyboard shortcuts (`h`/`.` for "chambre noire" mode, `Escape` to close lightbox), Astro `ClientRouter`.
- **`Roll.astro`** — renders a single photo series: three-column layout (left poem | center image | right poem), scroll parallax on `.thought-fragment` elements, click → `open-lightbox` custom event.
- **`LightboxGlobal.astro`** — shared lightbox DOM mounted once; listens for `open-lightbox` events. `src/scripts/lightbox/loupe.ts` attaches a hold-to-zoom magnifier.
- **`DustCanvas.astro`** — ambient particle layer behind the content, mounted globally.
- **`Navbar.astro`** — floating glassmorphism nav bar, `AudioPlayer` integration, search toggle.
- **`Search.astro`** — client-side search against `public/search-index.json` (fetched on first open).
- **`SplashScreen.astro`** — cinematic intro on first visit (uses `sessionStorage.splashSeen`). Fires `splashFinished` window event when done so index reveals its content.
- **`AudioPlayer.astro`** — ambient sound player for `public/ambiance-rue.mp3` or per-roll `audioUrl`.
- **`CustomCursor.astro`** — custom dot cursor, hidden via CSS when `prefers-reduced-motion`.

### Interaction utilities

- **`src/scripts/animations/poemAnimations.ts`** + **`src/styles/reveal.css`** — scroll-reveal engine for thought fragments and poem rhythm.
- **`src/scripts/utils/{gestures,haptics,once}.ts`** — mobile swipe/pinch support with optional vibration feedback. `once.ts` guards window-level listeners across View Transitions.
- **`src/scripts/utils/audio.ts`** — shared audio helpers for AudioPlayer and per-roll ambiance.
- **`src/scripts/utils/formatExif.ts`** — EXIF formatting used in lightbox and roll metadata overlays.
- **`src/scripts/utils/colors.ts`** — aura engine + Color-Lock mechanism.
- **`src/scripts/utils/scroll.ts`** — centralised scroll-reveal utility.

### View Transitions & Persistence

Astro's `ClientRouter` is active. The aura container has `transition:persist` so it never re-mounts between navigations. Global event listeners (lightbox, keyboard, aura-lock) are guarded by `window.__aura_initialized`, `window.__lightbox_initialized` and `window.__reveal_initialized` flags to prevent duplicate binding across navigations. All page-specific JS re-runs on `astro:page-load`.

A Chromium workaround in `Layout.astro` deletes `Element.prototype.moveBefore` before the router boots — Astro 6.1.1 otherwise throws `invalid hierarchy` on the first roll-to-roll navigation.

### Content Schema

`src/content.config.ts` — the `rolls` collection uses Astro's `glob` loader on `src/content/rolls/**/*.md`. The schema is defined with Zod. The `images` array is fully typed with optional per-photo `poem`, `palette`, and `dominantColor`.

### Tag Categories

Tags come from IPTC/XMP metadata on photos. `src/data/categories.ts` maps known tags to display categories (Ambiance, Lieu, Technique, Sujet). Unknown tags default to "Sujet". Category keys and tag values are kept in French to match the site's editorial voice.
