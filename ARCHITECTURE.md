# Architecture

Deep technical reference for this portfolio. README, CLAUDE.md and GEMINI.md all point here for the substantive content — keep the canonical architecture here to avoid drift.

## High-level shape

```
Infomaniak kDrive (image source)
        │
        ▼
src/scripts/go/kdrive-sync/          (npm run sync — Go binary)
  ├─ cmd/                            (wiring: main, extractpalette CLI)
  ├─ pkg/usecase/                    (orchestration, sync_rolls)
  ├─ pkg/infrastructure/             (kdriveapi, filelister, filedownloader,
  │                                   imageanalyzer, paletteaggregator,
  │                                   poetryparser, rollwriter,
  │                                   searchindexwriter, sharepublisher)
  ├─ pkg/domain/                     (pure types: exif, image, roll, slug)
  └─ pkg/service/                    (service interfaces + thread-safe fakes)
        │
        ▼
src/content/rolls/synced/*.md        (one Markdown file per Roll)
public/search-index.json             (full-text search index)
        │
        ▼
Astro build (static site)
  ├─ src/content.config.ts           (Zod schema for the `rolls` collection)
  ├─ src/layouts/Layout.astro        (composer)
  ├─ src/components/                 (Roll, Lightbox, Navbar, Search, Splash, …)
  └─ src/pages/                      (index, about, 404, roll/[slug], tags/[tag])
```

## Content schema

`src/content.config.ts` declares the `rolls` collection via Astro's `glob` loader on `src/content/rolls/**/*.md`. The Zod schema per roll:

| Field | Type | Notes |
|---|---|---|
| `title` | string | Display name, also slugified for the URL. |
| `date` | date | Shoot date, parsed from EXIF. |
| `tags[]` | string[] | IPTC/XMP keywords; mapped to categories by `src/data/categories.ts`. |
| `poem` | string? | Global poem shown under the Roll header. |
| `palette[]` | string[] | 5 hex colors (k-means in CIELAB). |
| `dominantColor` | string | Single hex. |
| `audioUrl` | string? | Optional per-roll ambient audio override. |
| `images[]` | object[] | Per-image record: `url`, `exif`, optional `poem`, `palette`, `dominantColor`. |

## Go sync module

- Architecture: `cmd/` → `pkg/usecase/` → `pkg/infrastructure/*` → `pkg/domain/`.
- A secondary `cmd/extractpalette` CLI exposes the palette engine for one-off runs.
- Test strategy
  - **Table-driven (`t.Run`)** for pure packages: `pkg/domain/slug`, `paletteaggregator`, `poetryparser`, `rollwriter`, `searchindexwriter`, `imageanalyzer`.
  - **Ginkgo v2 + Gomega BDD** for orchestration and HTTP-adjacent packages: `pkg/usecase/sync_rolls`, `pkg/infrastructure/{kdriveapi, filelister, filedownloader, sharepublisher}`.
  - **Shared fakes** under `pkg/service/servicefakes/` (thread-safe, Stub + Results + Calls pattern).
  - **godog + go-rod e2e** under `e2e/` drives the Astro frontend; excluded from the unit CI run (`go list ./... | grep -v /e2e`).
- CI gate: `go test -race -coverprofile -coverpkg=./pkg/...` with a ≥ 70 % threshold (`.github/workflows/ci.yml`).

## Aura 2.0 — ambient color engine

Drives the entire visual identity; works in five moving parts:

1. **Design tokens** in `src/styles/global.css`: CSS custom properties (`--p1`..`--p5`, `--bg-base`, `--text-main`, `--text-muted`, `--accent`, `--aura-opacity`) with `@property` declarations (CSS Houdini) so they animate smoothly.
2. **Five `.aura-blob` divs** inside `NavigationChrome.astro`, wrapped in a `<div class="bg-aura" transition:persist="aura-container">` so they survive View Transitions.
3. **`src/scripts/utils/colors.ts`** — the client-side engine:
   - `applyThemeColors(palette, isVisible)` updates the CSS vars and the `data-theme` attribute based on average luminance (threshold 0.38 switches light/dark text).
   - `setAuraLock(bool)` pauses transitions on the blobs during View Transitions (the **Color-Lock** mechanism); triggered on `astro:before-preparation`, released on `astro:page-load`.
   - `resetColorCache()` clears the memoised palette lookup on navigation.
4. **Scroll triggers** — in `Roll.astro`, each `image-wrapper` carries a `data-palette` and an `IntersectionObserver` swaps the aura as each photo enters view.
5. **Hover triggers** — on the index, hovering a roll link previews its palette.

A Chromium workaround in `HeadMeta.astro` deletes `Element.prototype.moveBefore` before the router boots — Astro 6.1.1 otherwise throws `invalid hierarchy` on the first roll-to-roll navigation.

## Layout composition

`Layout.astro` is a thin shell (~60 lines). It wires four focused pieces:

- **`src/components/layout/HeadMeta.astro`** — all `<head>` content: charset, viewport, favicon, title, SEO/OG/Twitter meta, JSON-LD, preload hint, anti-FOUC script, Chromium workaround, `ClientRouter`.
- **`src/components/layout/NavigationChrome.astro`** — aura blobs, noise SVG, photo-flash div, and the Color-Lock script (`astro:before-preparation` / `astro:after-swap` handlers, palette hand-off between rolls).
- **`src/components/layout/ReadProgress.astro`** — reading progress bar and its scroll handler, only active on roll pages.
- **`src/styles/global.css`** — design tokens, aura CSS, reduced-motion and forced-colors fallbacks, chambre noire overrides, photo flash keyframes.

Page-level components mounted globally by Layout:

- **`LightboxGlobal.astro`** + **`src/scripts/lightbox/loupe.ts`** — shared lightbox DOM; driven by `open-lightbox` custom events. Hold-to-zoom loupe attached on demand.
- **`DustCanvas.astro`** — ambient particle layer.
- **`CustomCursor.astro`** — custom dot cursor; hidden via CSS when `prefers-reduced-motion`.

## Page-level components

- **`Roll.astro`** — three-column layout (left poem | center image | right poem), parallax on `.thought-fragment` elements, click → `open-lightbox` custom event, `data-palette` for scroll-driven aura updates.
- **`Navbar.astro`** — floating glassmorphism nav bar, AudioPlayer integration, search toggle.
- **`Search.astro`** — client-side search against `public/search-index.json` (fetched on first open).
- **`SplashScreen.astro`** — cinematic intro on first visit (`sessionStorage.splashSeen`). Fires a `splashFinished` window event so the index can reveal its content.
- **`AudioPlayer.astro`** — ambient sound player for `public/ambiance-rue.mp3` or per-roll `audioUrl`.

## Interaction utilities

- **`src/scripts/animations/poemAnimations.ts`** + **`src/styles/reveal.css`** — scroll-reveal engine for thought fragments and poem rhythm.
- **`src/scripts/utils/{gestures,haptics,once}.ts`** — mobile swipe/pinch with optional vibration; `once.ts` guards window-level listeners across View Transitions.
- **`src/scripts/utils/audio.ts`** — shared audio helpers for AudioPlayer and per-roll ambiance (and the film-advance cue played on navigation).
- **`src/scripts/utils/formatExif.ts`** — EXIF formatting used in lightbox and roll metadata overlays.
- **`src/scripts/utils/colors.ts`** — aura engine + Color-Lock mechanism.
- **`src/scripts/utils/scroll.ts`** — centralised scroll-reveal utility.

## View Transitions & persistence

Astro's `ClientRouter` is active. The aura container has `transition:persist` so it never re-mounts between navigations. Global event listeners (lightbox, keyboard, aura-lock, reveal) are guarded by `window.__aura_initialized`, `window.__lightbox_initialized` and `window.__reveal_initialized` flags to prevent duplicate binding across navigations. Page-specific JS re-runs on `astro:page-load`.

## Tag categories

Tags come from IPTC/XMP metadata on photos. `src/data/categories.ts` maps known tags to display categories (`Ambiance`, `Lieu`, `Technique`, `Sujet`). Unknown tags fall back to `Sujet`. Category keys and tag values are kept in French to match the site's editorial voice.

## CI & deployment

- **`.github/workflows/ci.yml`** — Go race + coverage (≥ 70 %), `astro check`, Playwright (chromium + Pixel 7 mobile project).
- **`.github/workflows/lighthouse.yml`** + **`lighthouserc.json`** — 3-run Lighthouse budget on preview builds.
- **`vercel.json`** — responsive image pipeline (AVIF/WebP, cache TTL).
- **`@vite-pwa/astro`** + **`@astrojs/sitemap`** + **`@vercel/analytics`** — wired through `astro.config.mjs` and `Layout.astro`.
- **`scripts/generate-search-index.mjs`** — bootstraps `public/search-index.json` before the Playwright dev server boots, so tests don't require a full kDrive sync.

## Testing matrix

| Layer | Runner | Location |
|---|---|---|
| Frontend unit | Vitest | `tests/unit/` |
| Frontend e2e (chromium) | Playwright | `tests/e2e/` |
| Frontend e2e (mobile gestures) | Playwright Pixel 7 project | `tests/e2e/gestures.spec.ts` |
| Go pure packages | `go test` (table-driven) | `src/scripts/go/kdrive-sync/pkg/**` |
| Go orchestration | Ginkgo v2 + Gomega | `src/scripts/go/kdrive-sync/pkg/**` |
| Go full-stack e2e | godog + go-rod | `src/scripts/go/kdrive-sync/e2e/` |
