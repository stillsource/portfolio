# Marvelous Journal of a Wanderer — Engineering Guide

Onboarding guide for developing and maintaining this photo portfolio. Captures the architectural choices and technical specifics.

## 1. "Zero Storage" philosophy

The project stores no images locally (apart from the favicon and PWA icons).

- **Source**: Photos are hosted on Infomaniak kDrive.
- **Sync**: The Go program in `src/scripts/go/kdrive-sync/` retrieves public URLs and metadata (EXIF, IPTC/XMP tags, CIELAB k-means palette, dominant color) and generates Markdown rolls under `src/content/rolls/synced/`.
- **Rendering**: Astro generates static pages from the `rolls` collection (`src/content/rolls/**/*.md`), driven by a Zod schema in `src/content.config.ts`.

## 2. "Aura 2.0" system (Dynamic ambience)

The site uses an immersive color system that adapts to the content.

- **Extraction**: k-means clustering in CIELAB space produces a 5-color palette per image, aggregated per roll by the Go sync.
- **Display**: `Layout.astro` hosts five `.aura-blob` layers whose CSS variables `--p1` through `--p5` are updated by `src/scripts/utils/colors.ts` on scroll, hover or route change.
- **Contrast**: `--text-main` and `--text-muted` auto-switch between white and black based on average palette luminance (threshold 0.38).
- **Color-Lock**: `setAuraLock(true)` during `astro:before-preparation` prevents palette flicker between View Transitions; the lock is released on `astro:page-load`.

## 3. Key UI surfaces

- **Shared lightbox + loupe** (`LightboxGlobal.astro` + `src/scripts/lightbox/loupe.ts`) — mounted once in `Layout.astro`, driven by `open-lightbox` custom events.
- **DustCanvas** (`DustCanvas.astro`) — ambient particle layer.
- **Scroll-reveal typography** (`src/scripts/animations/poemAnimations.ts` + `src/styles/reveal.css`).
- **Gestures & haptics** (`src/scripts/utils/{gestures,haptics,once}.ts`) — mobile swipe/pinch with optional vibration.
- **Chambre noire** — `h` / `.` shortcut dims the chrome for pure photo viewing.
- **Splash screen** — cinematic intro, gated by `sessionStorage.splashSeen`.

## 4. Development conventions

- **Styles**: CSS custom properties with `@property` declarations (CSS Houdini) for smooth transitions. Global layout styles use `is:global` so they survive View Transitions.
- **Navigation**: `ClientRouter` drives transitions. The aura container and audio player use `transition:persist`. Init flags (`window.__aura_initialized`, `__lightbox_initialized`, `__reveal_initialized`) guard against duplicate listener binding.
- **Images**: Native `<img>` with `loading="lazy"` and `decoding="async"` for remote kDrive images, which avoids Astro's server-side dimension computation errors.
- **Types**: Shared shapes live in `src/types/content.ts`. Ambient module declarations (e.g. `@fontsource/*` side-effect imports) live in `src/env.d.ts`.

## 5. Adding content

1. Create a folder on kDrive with the roll title (e.g. "Nuit à Tokyo").
2. Drop JPEG photos inside (tag them with your keywords in Lightroom — IPTC/XMP).
3. Optional: drop a `poem.md` inside the same kDrive folder with frontmatter mapping filenames to per-photo verses plus a global body (consumed by `pkg/infrastructure/poetryparser`).
4. Run `npm run sync` to regenerate `src/content/rolls/synced/*.md` and `public/search-index.json`.

## 6. Testing

- **Vitest** (`npm run test:unit`) — unit specs under `tests/unit/`.
- **Playwright** (`npx playwright test`) — e2e specs under `tests/e2e/` (chromium + Pixel 7 mobile project for gestures).
- **Go unit tests** (`go test -C src/scripts/go/kdrive-sync ./...`) — table-driven for pure packages, Ginkgo v2 + Gomega BDD for orchestration.
- **godog + go-rod e2e** under `src/scripts/go/kdrive-sync/e2e/` — excluded from the unit CI run.

## 7. CI & deployment

- `.github/workflows/ci.yml` — Go race + coverage (≥ 70 %), `astro check`, Playwright.
- `.github/workflows/lighthouse.yml` — Lighthouse budget on preview builds.
- `vercel.json` — Vercel image pipeline (AVIF/WebP, cache TTL).
- `@vite-pwa/astro` + `@astrojs/sitemap` + `@vercel/analytics` wired through `astro.config.mjs`.

## 8. Maintenance

- **Shared helpers**: `src/scripts/utils/` (color, scroll, audio, exif, gestures, haptics, once).
- **Categories**: editable in `src/data/categories.ts` (kept in French to match editorial voice).
- **Search index**: `scripts/generate-search-index.mjs` bootstraps `public/search-index.json` for local dev and Playwright without requiring a full kDrive sync.
