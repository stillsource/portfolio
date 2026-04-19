# Marvelous journal of a wanderer — Street Photography Portfolio

A minimalist, performant and artistic photo portfolio, built with **Astro 6** and a **Go** sync pipeline. The site keeps a "Zero Storage" architecture (images stay on your kDrive) and ships an immersive experience: cinematic transitions, adaptive ambient colors (Aura 2.0), shared lightbox with loupe, particle dust layer, scroll-reveal typography and optional soundscape.

## Requirements

- Node.js ≥ 22
- Go ≥ 1.22 (only for `npm run sync` / `npm run build:full`)
- An Infomaniak kDrive API token (only for real content sync)

## Quick start (local)

1. Clone this repository.
2. Install dependencies: `npm install`
3. Start the dev server: `npm run dev`

`npm run dev` serves whatever content already sits in `src/content/rolls/` — it does **not** require a `.env` file. Use it for UI work against the committed rolls.

## kDrive configuration (image source)

The site stores **no** heavy photos. The Go sync generates Markdown rolls whose images point at your kDrive public URLs.

1. Go to your Infomaniak console and generate a kDrive API token.
2. Locate your Drive ID and the ID of the root folder that contains your "Roll" sub-folders.
3. Copy `.env.example` to `.env` and fill in the three variables:
   ```env
   KDRIVE_API_TOKEN=your_infomaniak_api_token_here
   KDRIVE_DRIVE_ID=your_drive_id_here
   KDRIVE_FOLDER_ID=root_folder_id_containing_roll_subfolders
   ```

**Expected kDrive layout:**
- `Root folder (KDRIVE_FOLDER_ID)`
  - `Walk in Paris/` — folder name becomes the Roll title
    - `photo1.jpg`
    - `photo2.jpg`
    - `poem.md` — optional, see below
  - `Night neons/`
    - `DSCF1234.jpg`

## Adding poetry

Drop an optional `poem.md` file inside any kDrive Roll folder. The Go sync parses its frontmatter and per-photo map:

```markdown
---
photos:
  "DSCF1234.jpg": |
    Les néons ne pleurent jamais,
    Ils saignent de la lumière.
---
Global poem shown under the Roll title.
```

Per-photo verses appear on hover/click of each image; the global text sits under the Roll header. Poems survive kDrive resyncs — they are merged into the generated `src/content/rolls/<slug>.md` frontmatter.

## Adding ambient audio

1. Find a royalty-free audio file (city noise, rain, musical drone) in `.mp3`.
2. Name it `ambiance-rue.mp3` and drop it in `public/`.
3. The sound icon in the floating navbar toggles it on/off. Per-roll ambiance can be set via the `audioUrl` frontmatter field.

## Commands

```bash
npm install            # Install JS dependencies
npm run dev            # Astro dev server on http://localhost:4321
npm run sync           # Run the Go kdrive-sync once (requires .env)
npm run build          # Astro build only (uses committed content)
npm run build:full     # sync + astro build (production pipeline)
npm run preview        # Preview the built site
npm run test:unit      # Vitest unit tests
npm run test:unit:watch
npx playwright test    # End-to-end suite (chromium + mobile project)
```

## Go sync module

The sync binary lives in `src/scripts/go/kdrive-sync/`. It follows a `cmd → usecase → infrastructure → domain` layout and ships a dedicated `extractpalette` CLI for one-off palette extraction.

```bash
# From the module root
make test            # go test ./...
make test-race       # race detector
make test-coverage   # coverage.html + total %
make test-watch      # ginkgo watch -r
golangci-lint run    # lint (uses .golangci.yml)
```

A godog + go-rod BDD suite under `e2e/` drives the Astro frontend and is excluded from the unit CI run. Architecture details and the test strategy live in `CLAUDE.md`.

## CI & deployment

- **GitHub Actions** (`.github/workflows/ci.yml`): Go unit tests with race detector + coverage gate (≥ 70 % on `pkg/...`), `astro check`, and the Playwright e2e suite.
- **Lighthouse** (`.github/workflows/lighthouse.yml` + `lighthouserc.json`): 3-run performance budget on preview builds.
- **Vercel** (`vercel.json`): responsive image sizes (AVIF/WebP) and cache TTL for the kDrive image pipeline.
- **PWA**: `@vite-pwa/astro` registers a service worker and web manifest; icons live in `public/icons/`.
- **Sitemap**: `@astrojs/sitemap` emits a sitemap on production builds.
- **Analytics**: `@vercel/analytics` is wired into the layout for production deploys.

## Design & immersion features

- **Aura 2.0** — Dynamic color engine. CSS Houdini `@property` palette vars, five persisted `.aura-blob` layers, `setAuraLock` to pause transitions during View Transitions, and `IntersectionObserver`-driven palette swaps as you scroll through photos.
- **Shared lightbox + loupe** — Global DOM mounted once in `Layout.astro`, driven by `open-lightbox` events. Includes a hold-to-zoom magnifier.
- **DustCanvas** — Ambient particle layer behind the content.
- **Scroll-reveal typography** — Poem fragments animate in as they enter the viewport (`src/scripts/animations/poemAnimations.ts` + `src/styles/reveal.css`).
- **Gestures & haptics** — Mobile swipe/pinch support with optional vibration feedback; tested under a Playwright Pixel 7 project.
- **Chambre noire mode** — `h` / `.` shortcut dims the chrome for pure photo viewing.
- **Splash screen** — Cinematic intro on first visit, gated by `sessionStorage`.
- **Glassmorphism navbar**, **SVG film grain**, **View Transitions** via `ClientRouter` with Color-Lock, **`translate3d` / `will-change`** optimizations throughout.

## Content schema

`src/content.config.ts` declares the `rolls` collection. Each `src/content/rolls/<slug>.md` has frontmatter with `title`, `date`, `tags[]`, `poem`, `palette[]`, `dominantColor`, `audioUrl`, and a fully typed `images[]` array (url, exif, per-photo poem, palette, dominantColor).

## Project layout

```
src/
├─ components/      # Astro UI (Roll, Lightbox, Navbar, Search, Splash, …)
├─ content/rolls/   # Generated Markdown rolls (output of Go sync)
├─ data/            # Tag categories + legacy poetry seeds
├─ layouts/         # Global shell
├─ pages/           # Routes (index, about, 404, roll/[slug], tags/[tag])
├─ scripts/
│  ├─ animations/   # Scroll reveal
│  ├─ lightbox/     # Loupe magnifier
│  ├─ utils/        # audio, colors, formatExif, gestures, haptics, scroll, …
│  └─ go/kdrive-sync/  # Go sync binary (cmd, pkg, e2e)
├─ styles/          # Shared CSS (reveal.css)
└─ types/           # Shared TS types (content.ts)

tests/
├─ unit/            # Vitest specs
└─ e2e/             # Playwright specs (chromium + mobile)
```

Contribution notes, architecture details and agent-specific instructions live in `CLAUDE.md` and `GEMINI.md`.
