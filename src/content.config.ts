import { defineCollection } from 'astro:content';
import { glob } from 'astro/loaders';
import { rollSchema } from './types/content';

const rollCollection = defineCollection({
  loader: glob({ pattern: "**/*.md", base: "./src/content/rolls" }),
  schema: rollSchema,
});

export const collections = {
  'rolls': rollCollection,
};
