import { z } from 'zod';

export const exifSchema = z.object({
  shutter: z.string().optional(),
  aperture: z.string().optional(),
  iso: z.string().optional(),
  body: z.string().optional(),
  lens: z.string().optional(),
  focalLength: z.string().optional(),
});

export const rollSchema = z.object({
  title: z.string(),
  date: z.date(),
  tags: z.array(z.string()).optional(),
  poem: z.string().optional(), // Global poem for the whole Roll
  palette: z.array(z.string()).optional(), // Roll-level color palette
  dominantColor: z.string().optional(), // Roll dominant color (produced by the Go sync)
  audioUrl: z.string().optional(), // URL of the contextual ambient audio file
  videoUrl: z.string().optional(), // URL of the ambient video (played on hover on the index)
  poemAnimation: z.enum(['scroll', 'typewriter', 'fade', 'slide', 'word', 'blur', 'random']).optional(), // Per-photo poem animation mode
  images: z.array(z.object({
    url: z.string(),
    alt: z.string().optional(),
    exif: exifSchema.optional(),
    metadata: z.string().optional(), // Pre-formatted EXIF caption string (produced by the Go sync)
    poem: z.string().optional(),
    palette: z.array(z.string()).optional(),
    dominantColor: z.string().optional(),
    blurDataUrl: z.string().optional(), // Base64 LQIP placeholder produced by the Go sync
    width: z.number().int().positive().optional(), // Natural pixel width (reduces CLS; produced by the Go sync)
    height: z.number().int().positive().optional(),
    orientation: z.enum(['portrait', 'landscape']).optional(), // Pre-computed so the lightbox doesn't re-probe on open
    size: z.enum(['full', 'large', 'medium', 'small']).optional(), // Display width
    layout: z.enum(['single', 'pair-left', 'pair-right']).optional(), // Row composition
  })),
});

export type ExifData = z.infer<typeof exifSchema>;
export type RollData = z.infer<typeof rollSchema>;
export type ImageItemData = RollData['images'][number];

export interface SearchIndexItem {
  id: string;
  title: string;
  date: string;
  tags: string[];
  poem?: string;
  cover?: string;
  palette: string[];
}
