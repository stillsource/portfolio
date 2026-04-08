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
  poem: z.string().optional(), // Poème global pour le Roll
  palette: z.array(z.string()).optional(), // Palette de couleurs du Roll
  dominantColor: z.string().optional(), // Couleur dominante du Roll (générée par sync)
  audioUrl: z.string().optional(), // URL du fichier audio contextuel
  images: z.array(z.object({
    url: z.string(),
    exif: exifSchema.optional(),
    metadata: z.string().optional(),
    poem: z.string().optional(),
    palette: z.array(z.string()).optional(),
    dominantColor: z.string().optional()
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
