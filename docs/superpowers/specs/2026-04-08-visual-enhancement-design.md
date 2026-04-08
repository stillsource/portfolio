# Design Spec: Visual Enhancement - Aura Sync & Metadata

This document defines the technical design for synchronizing the photographic metadata (EXIF) and the immersive color system (Aura) between the kDrive source and the Astro frontend.

## 1. Goal
- Implement a robust structured EXIF metadata system (Option A).
- Calculate a "Real Mean Palette" for each Roll to capture the global atmosphere.
- Synchronize the Aura system to use these enriched data points for a more immersive experience.

## 2. Data Architecture (Astro Content Schema)

The `rolls` collection in `src/content.config.ts` will be updated to enforce the following structure:

### EXIF Object
```typescript
{
  shutter: string;      // e.g., "1/500s"
  aperture: string;     // e.g., "f/2.8"
  iso: string;          // e.g., "ISO 400"
  body: string;         // e.g., "Fujifilm X-T4"
  lens: string;         // e.g., "XF23mmF2 R WR"
  focalLength: string;  // e.g., "23mm"
}
```

### Color Metadata
- `palette`: An array of 5 hex strings.
- `dominantColor`: A single hex string representing the primary tone for the Aura.

## 3. Synchronization Logic (`fetch-kdrive.ts`)

### Image Processing
For each image:
1. **EXIF Extraction**: Use `exifr` to parse raw metadata and map it to the structured object.
2. **Color Extraction**: Use `node-vibrant` to get the 5 core swatches.
3. **Dominant Color**: Select the most "vibrant" or "muted" swatch based on a consistent logic (e.g., population-weighted) to serve as the `dominantColor`.

### Roll Processing (The Atmospheric Mean)
To calculate the `palette` for the entire Roll:
1. **Collection**: Gather all hex palettes from every image in the folder.
2. **CIELAB Conversion**: Convert all hex values to LAB color space for perceptual accuracy.
3. **K-Means Clustering / Weighted Averaging**: Aggregate these values to find the 5 "mean" colors that represent the overall mood of the collection.
4. **Dominant Color**: The first color of this mean palette becomes the Roll's `dominantColor`.

## 4. UI Implementation

### Roll Component (`Roll.astro`)
- **Metadata Display**: Format the EXIF object into a clean, minimalist line (e.g., `Fujifilm X-T4 • 23mm • f/2.0 • 1/500s • ISO 400`).
- **Aura Trigger**: Each image wrapper will continue to use `IntersectionObserver` to update the global Aura, but will now pass the high-fidelity structured palette.

### Theme Logic (`colors.ts`)
- **Luminance Guard**: Keep the existing logic that flips text color (black/white) based on the palette's average luminance (> 0.38) to ensure readability against the Aura.

## 5. Success Criteria
- [ ] `npm run sync` generates Markdown files with structured EXIF objects and calculated mean palettes.
- [ ] The Aura starts with the Roll's global mood and transitions smoothly between photo-specific palettes.
- [ ] Metadata is displayed consistently and elegantly across all images.
- [ ] TypeScript types are fully respected in all components.
