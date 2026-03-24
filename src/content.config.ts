import { z, defineCollection } from 'astro:content';
import { glob } from 'astro/loaders';

const rollCollection = defineCollection({
  loader: glob({ pattern: "**/*.md", base: "./src/content/rolls" }),
  schema: z.object({
    title: z.string(),
    date: z.date(),
    tags: z.array(z.string()).optional(),
    poem: z.string().optional(), // Poème global pour le Roll
    palette: z.array(z.string()).optional(), // Palette de couleurs du Roll
    audioUrl: z.string().optional(), // URL du fichier audio contextuel
    images: z.array(z.object({
      url: z.string(),
      metadata: z.string().optional(),
      poem: z.string().optional(),
      palette: z.array(z.string()).optional()
    })), 
  }),
});

export const collections = {
  'rolls': rollCollection,
};
