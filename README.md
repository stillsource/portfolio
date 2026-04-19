# Marvelous journal of a wanderer — Street Photography Portfolio

A minimalist, performant and artistic photo portfolio, built with **Astro 6** and a **Go** sync pipeline. The site keeps a "Zero Storage" architecture (images stay on your kDrive) and ships an immersive experience: cinematic transitions, adaptive ambient colors (Aura 2.0), shared lightbox with loupe, particle dust layer, scroll-reveal typography and optional soundscape.

> For a deep tour of the code (Go sync internals, Aura engine, layout composition, testing matrix), read [**ARCHITECTURE.md**](./ARCHITECTURE.md).

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

The sync binary lives in `src/scripts/go/kdrive-sync/`. From the module root:

```bash
make test            # go test ./...
make test-race       # race detector
make test-coverage   # coverage.html + total %
make test-watch      # ginkgo watch -r
golangci-lint run    # lint (uses .golangci.yml)
```

Architecture, test strategy and CI gate are detailed in [ARCHITECTURE.md](./ARCHITECTURE.md#go-sync-module).

## Project layout

```
src/
├─ components/      # Astro UI (Roll, Lightbox, Navbar, Search, Splash, …)
│  └─ layout/       # Layout sub-components (HeadMeta, NavigationChrome, ReadProgress)
├─ content/rolls/   # Generated Markdown rolls (output of Go sync)
├─ data/            # Tag categories
├─ layouts/         # Global shell (Layout.astro — thin composer)
├─ pages/           # Routes (index, about, 404, roll/[slug], tags/[tag])
├─ scripts/
│  ├─ animations/   # Scroll reveal
│  ├─ lightbox/     # Loupe magnifier
│  ├─ utils/        # audio, colors, formatExif, gestures, haptics, scroll, …
│  └─ go/kdrive-sync/  # Go sync binary (cmd, pkg, e2e)
├─ styles/          # Global CSS + reveal animations
└─ types/           # Shared TS types

tests/
├─ unit/            # Vitest specs
└─ e2e/             # Playwright specs (chromium + mobile)
```

See [ARCHITECTURE.md](./ARCHITECTURE.md) for the Aura engine internals, component responsibilities, interaction utilities, view transitions, CI & deployment and the full testing matrix.

Agent-specific briefs: [`CLAUDE.md`](./CLAUDE.md) (Claude Code) and [`GEMINI.md`](./GEMINI.md) (Gemini CLI).
