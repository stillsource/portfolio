#!/usr/bin/env node
/**
 * Generates public/search-index.json from local roll markdown files.
 * Used by E2E tests and dev mode (no kDrive credentials needed).
 */
import { readdir, readFile, writeFile } from 'node:fs/promises';
import { join, basename } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = fileURLToPath(new URL('..', import.meta.url));
const rollDirs = [
  join(root, 'src/content/rolls'),
  join(root, 'src/content/rolls/synced'),
];
const outFile = join(root, 'public/search-index.json');

function parseFrontmatter(content) {
  const match = content.match(/^---\n([\s\S]*?)\n---/);
  if (!match) return null;
  const fm = match[1];

  const get = (key) => {
    const m = fm.match(new RegExp(`^${key}:\\s*(.+)$`, 'm'));
    return m ? m[1].trim() : null;
  };

  const title = get('title')?.replace(/^["']|["']$/g, '') ?? '';
  const date = get('date') ?? '';
  const dominantColor = get('dominantColor')?.replace(/^["']|["']$/g, '') ?? '';
  const poem = get('poem')?.replace(/^["']|["']$/g, '') ?? undefined;

  // tags: ["a", "b", ...]
  const tagsMatch = fm.match(/^tags:\s*\[([^\]]*)\]/m);
  const tags = tagsMatch
    ? tagsMatch[1].split(',').map(t => t.trim().replace(/^["']|["']$/g, '')).filter(Boolean)
    : [];

  // palette: ["#hex", ...]
  const paletteMatch = fm.match(/^palette:\s*\[([^\]]*)\]/m);
  const palette = paletteMatch
    ? paletteMatch[1].split(',').map(t => t.trim().replace(/^["']|["']$/g, '')).filter(Boolean)
    : [];

  // First image url
  const imgMatch = fm.match(/"url":\s*"([^"]+)"/);
  const cover = imgMatch ? imgMatch[1] : undefined;

  return { title, date, tags, poem, palette, cover, dominantColor };
}

async function getMdFiles(dir) {
  try {
    const entries = await readdir(dir, { withFileTypes: true });
    return entries
      .filter(e => e.isFile() && e.name.endsWith('.md'))
      .map(e => join(dir, e.name));
  } catch {
    return [];
  }
}

const seen = new Set();
const index = [];

for (const dir of rollDirs) {
  const files = await getMdFiles(dir);
  for (const file of files) {
    const id = basename(file, '.md');
    if (seen.has(id)) continue;
    seen.add(id);
    const content = await readFile(file, 'utf8');
    const fm = parseFrontmatter(content);
    if (!fm || !fm.title) continue;
    index.push({ id, ...fm });
  }
}

await writeFile(outFile, JSON.stringify(index, null, 2));
console.log(`search-index.json: ${index.length} rolls indexed`);
