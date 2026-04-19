# Marvelous Journal of a Wanderer — Engineering Guide for Gemini CLI

Onboarding brief for Gemini CLI when working in this repository.

> Read [**ARCHITECTURE.md**](./ARCHITECTURE.md) first — it is the canonical reference for the Go sync pipeline, Aura 2.0 engine, layout composition, view transitions, tag categories, CI and the testing matrix. This file only carries what's specific to operating here and the "Zero Storage" philosophy at the top.

## Zero Storage philosophy

The project stores no images locally (apart from the favicon and PWA icons).

- **Source** — Photos live on Infomaniak kDrive.
- **Sync** — The Go program in `src/scripts/go/kdrive-sync/` retrieves public URLs and metadata (EXIF, IPTC/XMP tags, CIELAB k-means palette, dominant color) and emits Markdown rolls under `src/content/rolls/synced/`.
- **Rendering** — Astro generates static pages from the `rolls` collection (`src/content/rolls/**/*.md`), typed via a Zod schema in `src/content.config.ts`.

## Adding content

1. Create a folder on kDrive with the roll title (e.g. `Nuit à Tokyo`).
2. Drop JPEGs inside (tagged with IPTC/XMP keywords in Lightroom).
3. Optional: drop a `poem.md` file inside the same folder with frontmatter mapping filenames to per-photo verses plus a global body (consumed by `pkg/infrastructure/poetryparser`).
4. Run `npm run sync` to regenerate `src/content/rolls/synced/*.md` and `public/search-index.json`.

## Development conventions

- **Site copy stays in French** (UI labels, poems, meta descriptions). Technical files, docs, commit messages and code comments stay in English. Do **not** add `Co-Authored-By` trailers to commits.
- **Images**: native `<img>` with `loading="lazy"` and `decoding="async"` for remote kDrive images — avoids Astro's server-side dimension computation errors.
- **Styles**: CSS custom properties with `@property` declarations (Houdini) live in `src/styles/global.css`. Global styles sit in plain `.css` files, not `<style is:global>` blocks inside components.
- **Navigation**: `ClientRouter` drives transitions. The aura container and audio player use `transition:persist`. Init flags (`window.__aura_initialized`, `__lightbox_initialized`, `__reveal_initialized`) guard against duplicate listener binding.
- **Types**: shared shapes live in `src/types/content.ts`. Ambient module declarations (e.g. `@fontsource/*` side-effect imports) live in `src/env.d.ts`.

## Commands

See the cheat-sheet in [`CLAUDE.md`](./CLAUDE.md#commands) — the command list is identical regardless of which agent is driving.

## Maintenance pointers

- **Shared helpers** — `src/scripts/utils/` (color, scroll, audio, exif, gestures, haptics, once).
- **Categories** — editable in `src/data/categories.ts` (kept in French to match editorial voice).
- **Search index bootstrap** — `scripts/generate-search-index.mjs` primes `public/search-index.json` for local dev and Playwright without requiring a full kDrive sync.
