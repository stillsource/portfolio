# Finalize Visual Enhancement Plan (Tasks 3, 4, 5)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement CIELAB mean palette calculation in the sync script, update the Roll component with the scroll-to-top fix and initial theme application, and validate the build.

**Architecture:** Use CIELAB color space for accurate color averaging. CIELAB is better than RGB for mean color calculation because it's perceptually uniform. The Roll component will now set its own initial theme based on the aggregated roll palette.

**Tech Stack:** TypeScript, Astro, CIELAB color conversion.

---

### Task 3: Implement CIELAB Mean Palette in `src/scripts/fetch-kdrive.ts`

**Files:**
- Modify: `src/scripts/fetch-kdrive.ts`

- [ ] **Step 1: Add color conversion utilities and `calculateMeanPalette`**

Add these functions before the `sync()` function.

```typescript
// --- Utilitaires de conversion de couleur (CIELAB) ---

function hexToRgb(hex: string) {
  const r = parseInt(hex.slice(1, 3), 16);
  const g = parseInt(hex.slice(3, 5), 16);
  const b = parseInt(hex.slice(5, 7), 16);
  return { r, g, b };
}

function rgbToLab(r: number, g: number, b: number) {
  let r_ = r / 255, g_ = g / 255, b_ = b / 255;
  r_ = (r_ > 0.04045) ? Math.pow((r_ + 0.055) / 1.055, 2.4) : r_ / 12.92;
  g_ = (g_ > 0.04045) ? Math.pow((g_ + 0.055) / 1.055, 2.4) : g_ / 12.92;
  b_ = (b_ > 0.04045) ? Math.pow((b_ + 0.055) / 1.055, 2.4) : b_ / 12.92;
  
  let x = (r_ * 0.4124 + g_ * 0.3576 + b_ * 0.1805) * 100;
  let y = (r_ * 0.2126 + g_ * 0.7152 + b_ * 0.0722) * 100;
  let z = (r_ * 0.0193 + g_ * 0.1192 + b_ * 0.9505) * 100;
  
  x /= 95.047; y /= 100.000; z /= 108.883;
  x = (x > 0.008856) ? Math.pow(x, 1/3) : (7.787 * x) + (16 / 116);
  y = (y > 0.008856) ? Math.pow(y, 1/3) : (7.787 * y) + (16 / 116);
  z = (z > 0.008856) ? Math.pow(z, 1/3) : (7.787 * z) + (16 / 116);
  
  return { l: (116 * y) - 16, a: 500 * (x - y), b: 200 * (y - z) };
}

function labToRgb(l: number, a: number, b: number) {
  let y = (l + 16) / 116;
  let x = a / 500 + y;
  let z = y - b / 200;
  
  x = (Math.pow(x, 3) > 0.008856) ? Math.pow(x, 3) : (x - 16 / 116) / 7.787;
  y = (Math.pow(y, 3) > 0.008856) ? Math.pow(y, 3) : (y - 16 / 116) / 7.787;
  z = (Math.pow(z, 3) > 0.008856) ? Math.pow(z, 3) : (z - 16 / 116) / 7.787;
  
  x *= 95.047; y *= 100.000; z *= 108.883;
  
  let r = x * 3.2406 + y * -1.5372 + z * -0.4986;
  let g = x * -0.9689 + y * 1.8758 + z * 0.0415;
  let b_ = x * 0.0557 + y * -0.2040 + z * 1.0570;
  
  r /= 100; g /= 100; b_ /= 100;
  r = (r > 0.0031308) ? (1.055 * Math.pow(r, 1 / 2.4) - 0.055) : 12.92 * r;
  g = (g > 0.0031308) ? (1.055 * Math.pow(g, 1 / 2.4) - 0.055) : 12.92 * g;
  b_ = (b_ > 0.0031308) ? (1.055 * Math.pow(b_, 1 / 2.4) - 0.055) : 12.92 * b_;
  
  return {
    r: Math.max(0, Math.min(255, Math.round(r * 255))),
    g: Math.max(0, Math.min(255, Math.round(g * 255))),
    b: Math.max(0, Math.min(255, Math.round(b_ * 255)))
  };
}

function rgbToHex(r: number, g: number, b: number) {
  const toHex = (n: number) => n.toString(16).padStart(2, '0');
  return `#${toHex(r)}${toHex(g)}${toHex(b)}`;
}

/**
 * Calcule la palette moyenne d'un Roll en utilisant l'espace CIELAB.
 * Fusionne les palettes de N photos en 5 couleurs représentatives.
 */
function calculateMeanPalette(palettes: string[][]): string[] {
  if (palettes.length === 0) return [];
  
  // On s'assure que toutes les palettes ont 5 couleurs (ou on prend le min)
  const numColors = 5;
  const labSums = Array.from({ length: numColors }, () => ({ l: 0, a: 0, b: 0, count: 0 }));
  
  palettes.forEach(palette => {
    palette.forEach((hex, i) => {
      if (i < numColors) {
        const rgb = hexToRgb(hex);
        const lab = rgbToLab(rgb.r, rgb.g, rgb.b);
        labSums[i].l += lab.l;
        labSums[i].a += lab.a;
        labSums[i].b += lab.b;
        labSums[i].count++;
      }
    });
  });
  
  return labSums.map(sum => {
    if (sum.count === 0) return '#000000';
    const avgLab = { l: sum.l / sum.count, a: sum.a / sum.count, b: sum.b / sum.count };
    const rgb = labToRgb(avgLab.l, avgLab.a, avgLab.b);
    return rgbToHex(rgb.r, rgb.g, rgb.b);
  });
}
```

- [ ] **Step 2: Update `sync()` function to use mean palette**

In `sync()`, initialize `allPalettes` at the start of the `folder` loop. Collect palettes in `imagesData` and calculate the mean after processing all images.

```typescript
// Remplacer la boucle for (const folder of folders) par :

    for (const folder of folders) {
      const rollTitle = folder.name;
      const rollSlug = slugify(rollTitle, { lower: true, strict: true });
      const rollDate = new Date(folder.created_at * 1000).toISOString().split('T')[0];
      
      console.log(`\n📸 Traitement du Roll: ${rollTitle}`);

      const folderData = await fetchKDriveAPI(`/files/${folder.id}/files`);
      
      // ... (poetry and audio logic stays the same)

      const photos = folderData.data.filter((f: any) => 
        f.type === 'file' && (f.name.toLowerCase().endsWith('.jpg') || f.name.toLowerCase().endsWith('.jpeg'))
      );

      let rollTags = new Set<string>();
      let allPalettes: string[][] = []; // <--- Nouvelle collection

      const imagesData = await Promise.all(photos.map(async (photo: any) => {
        console.log(`   🔗 Traitement de: ${photo.name}...`);
        
        const publicUrl = await getPublicUrl(photo.id);
        const { tags, exif, palette } = await extractMetadataFromRemoteImage(photo.id);
        
        tags.forEach(t => rollTags.add(t));
        if (palette.length > 0) allPalettes.push(palette); // <--- Accumulation

        return {
          url: publicUrl,
          exif: exif,
          poem: poetryData.photos[photo.name] || undefined,
          palette: palette,
          dominantColor: palette.length > 0 ? palette[0] : undefined
        };
      }));

      const tagsArray = Array.from(rollTags);
      const rollPalette = calculateMeanPalette(allPalettes); // <--- Calcul final
      
      // Ajout à l'index de recherche global
      searchIndex.push({
        id: rollSlug,
        title: rollTitle,
        date: rollDate,
        tags: tagsArray,
        poem: poetryData.globalPoem,
        cover: imagesData.length > 0 ? imagesData[0].url : undefined,
        palette: rollPalette
      });

      const rollPoemAttr = poetryData.globalPoem ? `poem: ${JSON.stringify(poetryData.globalPoem)}\n` : '';
      const rollPaletteAttr = rollPalette.length > 0 ? `palette: ${JSON.stringify(rollPalette)}\n` : '';
      const rollDominantColorAttr = rollPalette.length > 0 ? `dominantColor: "${rollPalette[0]}"\n` : '';
      const rollAudioAttr = audioUrl ? `audioUrl: "${audioUrl}"\n` : '';
      
      const mdContent = `---
title: "${rollTitle}"
date: ${rollDate}
tags: ${JSON.stringify(tagsArray)}
${rollPoemAttr}${rollPaletteAttr}${rollDominantColorAttr}${rollAudioAttr}images: ${JSON.stringify(imagesData)}
---
`;
      const mdPath = path.join(LOCAL_CONTENT_DIR, `${rollSlug}.md`);
      await fs.writeFile(mdPath, mdContent);
      console.log(`   ✅ Fichier généré: ${rollSlug}.md`);
    }
```

---

### Task 4: Update `src/components/Roll.astro`

**Files:**
- Modify: `src/components/Roll.astro`

- [ ] **Step 1: Update `Props` and HTML**

Include `palette` in `Props` and add `data-roll-palette`.

```typescript
export interface Props {
  title: string;
  poem?: string;
  dominantColor?: string;
  palette?: string[]; // <--- Ajouté
  images: {
    // ...
  }[];
}
const { title, poem, dominantColor = '#050505', palette, images } = Astro.props; // <--- Ajouté palette
```

And in the HTML:

```html
<div class="roll-images" data-roll-palette={JSON.stringify(palette || [dominantColor])}>
```

- [ ] **Step 2: Update `initRoll()` script**

Add `window.scrollTo(0, 0);` and initial `applyThemeColors()`.

```javascript
  function initRoll() {
    // S'assurer que l'utilisateur commence en haut de la page (Fix Scroll demandé)
    window.scrollTo(0, 0);

    // Nettoyage de l'instance précédente si elle existe
    if (currentCleanup) {
      currentCleanup();
      currentCleanup = null;
    }

    // Appliquer l'ambiance initiale du Roll
    const rollContainer = document.querySelector('.roll-images');
    if (rollContainer) {
      const rollPaletteStr = (rollContainer as HTMLElement).dataset.rollPalette;
      const rollPalette = rollPaletteStr ? JSON.parse(rollPaletteStr) : [];
      if (rollPalette && rollPalette.length > 0) {
        applyThemeColors(rollPalette, true);
      }
    }

    const colorTriggers = document.querySelectorAll('.color-trigger');
    // ... reste de la fonction ...
```

---

### Task 4.5: Update Transition Trigger in `src/pages/roll/[slug].astro`

**Files:**
- Modify: `src/pages/roll/[slug].astro`

- [ ] **Step 1: Decouple reveal animation and transition trigger**

Modify the script to use a separate observer or target for the transition logic, ensuring it only starts when the title is visible.

```javascript
    // Dans le <script> de src/pages/roll/[slug].astro

    const observer = new IntersectionObserver((entries, observer) => {
      entries.forEach(entry => {
        const target = entry.target as HTMLElement;

        if (entry.isIntersecting) {
          target.classList.add('is-visible');
          
          // On ne lance plus le timer ici pour le footer entier
          if (target.id === 'next-roll-trigger') {
             // On laisse juste l'animation de reveal se faire via is-visible
          } else {
             observer.unobserve(target);
          }
        } else {
          if (target.id === 'next-roll-trigger') {
            target.classList.remove('is-visible', 'is-loading');
            document.body.classList.remove('is-transitioning');
            if (navigationTimer) {
              clearTimeout(navigationTimer);
              navigationTimer = null;
            }
          }
        }
      });
    }, observerOptions);

    // Nouvel observateur spécifique pour le titre (déclencheur réel de la transition)
    const transitionObserver = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        const footer = document.getElementById('next-roll-trigger');
        if (!footer) return;

        if (entry.isIntersecting) {
          const nextUrl = footer.dataset.nextUrl;
          if (nextUrl) {
            footer.classList.add('is-loading');
            
            setTimeout(() => {
              if (footer.classList.contains('is-visible')) {
                document.body.classList.add('is-transitioning');
              }
            }, 1500);
            
            if (navigationTimer) clearTimeout(navigationTimer);
            
            navigationTimer = setTimeout(() => {
              if (footer.classList.contains('is-visible')) {
                navigate(nextUrl);
              }
            }, 2500);
          }
        } else {
          // Si le titre n'est plus visible, on annule le chargement
          footer.classList.remove('is-loading');
          document.body.classList.remove('is-transitioning');
          if (navigationTimer) {
            clearTimeout(navigationTimer);
            navigationTimer = null;
          }
        }
      });
    }, { threshold: 0.5 }); // On attend que 50% du titre soit visible

    document.addEventListener('astro:page-load', () => {
      const elements = document.querySelectorAll('.reveal-on-scroll');
      elements.forEach(el => observer.observe(el));

      const nextTitle = document.querySelector('.next-title');
      if (nextTitle) transitionObserver.observe(nextTitle);
    });
```

---

### Task 5: Final Verification

- [ ] **Step 1: Sync and Build**

Run commands to verify everything works.

```bash
npm run sync
npx astro check
npm run build
```

- [ ] **Step 2: Commit**

```bash
git add src/scripts/fetch-kdrive.ts src/components/Roll.astro
git commit -m "feat: finalize visual enhancement with mean palette and scroll fix"
```
