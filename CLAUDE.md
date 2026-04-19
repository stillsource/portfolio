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

## Frontend tests

Playwright e2e suite at `tests/e2e/` (cursor, immersive-features, index, roll, search, tags). Run with `npx playwright test`. `playwright.config.ts` auto-starts `npm run dev` after generating the search index.

## Go Sync Script

The kdrive-sync binary lives in `src/scripts/go/kdrive-sync/`. It uses `go run` (no compilation step needed).

```bash
# From repo root:
go test -C src/scripts/go/kdrive-sync ./...          # Run Go tests
golangci-lint run -C src/scripts/go/kdrive-sync      # Lint (uses .golangci.yml)
# Or from src/scripts/go/kdrive-sync/ :
make test-race        # race detector
make test-coverage    # coverage.html + total %
make test-watch       # ginkgo watch -r
```

Architecture: `cmd/` → `pkg/usecase/` → `pkg/infrastructure/*` → `pkg/domain/`

Tests: table-driven (`t.Run`) for pure packages (`pkg/domain/slug`, `paletteaggregator`, `poetryparser`, `rollwriter`, `searchindexwriter`, `imageanalyzer`); Ginkgo v2 + Gomega BDD for orchestration / HTTP (`pkg/usecase`, `pkg/infrastructure/kdriveapi`, `filelister`, `filedownloader`, `sharepublisher`). Shared fakes live in `pkg/service/servicefakes/` (thread-safe, Stub + Results + Calls pattern). godog + go-rod e2e suite at `src/scripts/go/kdrive-sync/e2e/` drives the Astro frontend and is excluded from the unit CI run (`go list ./... | grep -v /e2e`).

CI gate: `go test -race -coverprofile -coverpkg=./pkg/...` with a ≥ 70% total threshold (`.github/workflows/ci.yml`).

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
