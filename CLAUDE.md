# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working in this repository.

> The canonical architecture reference is [**ARCHITECTURE.md**](./ARCHITECTURE.md). Read it before non-trivial edits — it covers the Go sync internals, Aura 2.0 engine, layout composition, content schema, view transitions, CI and the testing matrix. This file only carries the command cheat-sheet and Claude-specific guidance that doesn't live there.

## Commands

```bash
npm install               # Install dependencies
npm run dev               # Dev server (committed content, no kDrive sync required)
npm run sync              # Fetch from kDrive and regenerate content files
npm run build             # Astro build only (uses existing content)
npm run build:full        # sync + astro build (full production build)
npm run preview           # Preview production build
npm run test:unit         # Vitest unit tests
npm run test:unit:watch   # Vitest watch mode
npx playwright test       # End-to-end suite (chromium + Pixel 7 mobile project)

# Go module (src/scripts/go/kdrive-sync/)
go test -C src/scripts/go/kdrive-sync ./...          # Unit tests
golangci-lint run -C src/scripts/go/kdrive-sync      # Lint
make -C src/scripts/go/kdrive-sync test-race         # Race detector
make -C src/scripts/go/kdrive-sync test-coverage     # coverage.html + total %
```

`npm run dev` works without `.env` — it serves whatever content is already committed under `src/content/rolls/`. Only `npm run sync` and `npm run build:full` require kDrive credentials.

## Required environment variables

Create a `.env` at the repo root:

```
KDRIVE_API_TOKEN=
KDRIVE_DRIVE_ID=
KDRIVE_FOLDER_ID=   # ID of the root folder containing Roll sub-folders
```

## Working conventions

- **Site copy stays in French** (UI labels, poems, meta descriptions). Technical files, docs, commit messages and code comments stay in English.
- **Commit messages** follow Conventional Commits. Do **not** add `Co-Authored-By` trailers.
- **Layout.astro is a thin composer** — new global chrome belongs in `src/components/layout/*.astro`, not inline. Global CSS goes in `src/styles/global.css`.
- **View Transitions are live** (`ClientRouter`). Global listeners must be guarded by `window.__*_initialized` flags to avoid double-binding across navigations.
- **Go tests**: follow the existing split — table-driven (`t.Run`) for pure packages, Ginkgo v2 + Gomega for orchestration, shared fakes in `pkg/service/servicefakes/` (Stub + Results + Calls pattern). CI gate is ≥ 70 % coverage on `pkg/...`.

## Tag categories

Tags come from IPTC/XMP metadata on photos. `src/data/categories.ts` maps known tags to display categories (`Ambiance`, `Lieu`, `Technique`, `Sujet`). Unknown tags default to `Sujet`. Keep keys and tag values in French — this is data, not UI copy.
