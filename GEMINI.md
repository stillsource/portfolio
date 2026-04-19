# Marvelous Journal of a Wanderer - Engineering Guide

This document serves as an onboard guide for the development and maintenance of the photo portfolio. It documents the architectural choices and technical specifics of the project.

## 1. "Zero Storage" philosophy
The project is designed to store no image locally (apart from the favicon).
- **Source**: Photos are hosted on a kDrive (Infomaniak).
- **Sync**: The Go program `src/scripts/go/kdrive-sync/` retrieves public URLs and metadata (EXIF, tags, palette via k-means CIELAB).
- **Rendering**: Astro generates static pages from Markdown files located in `src/content/rolls/`.

## 2. "Aura" system (Dynamic ambience)
The site uses an immersive color system that adapts to the content.
- **Extraction**: Via k-means clustering in CIELAB space, we extract a palette of 5 colors per image (Go sync).
- **Display**: `Layout.astro` contains 5 spheres of light (`.aura-blob`) whose colors (`--p1` through `--p5`) are updated via JavaScript on scroll or hover.
- **Contrast**: The text color (`--text-main`, `--text-muted`) automatically switches from white to black if the average palette luminance exceeds 0.38.

## 3. Development conventions
- **Styles**: Use of CSS variables and the `@property` API for smooth transitions. The global index styles are in `is:global` to support `ViewTransitions`.
- **Navigation**: Transitions between pages are handled by `ClientRouter`. The audio uses `transition:persist` so it is not interrupted.
- **Images**: Use of the native `<img>` tag with `loading="lazy"` and `decoding="async"` for remote images to avoid server-side dimension computation errors.

## 4. Adding content
1. Create a folder on kDrive with the tag name (e.g. "Nuit à Tokyo").
2. Drop JPEG photos inside (tagged with your keywords in Lightroom).
3. Optional: Create a `src/data/poetry/[slug].md` file to add poetic texts per photo.
4. Run `npm run sync` to generate the Astro content files.

## 5. Maintenance
- **Scripts**: `src/scripts/utils/` contains the shared helpers for color computations.
- **Categories**: Editable in `src/data/categories.ts`.
