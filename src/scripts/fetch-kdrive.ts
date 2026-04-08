import fs from 'fs/promises';
import path from 'path';
import 'dotenv/config';
import exifr from 'exifr';
import slugify from 'slugify';
import matter from 'gray-matter';
import { Vibrant } from 'node-vibrant/node';

// ----------------------------------------------------------------------
// SCRIPT DE SYNCHRONISATION KDRIVE -> ASTRO (URL PUBLIQUES)
// ----------------------------------------------------------------------

const KDRIVE_API_TOKEN = process.env.KDRIVE_API_TOKEN;
const KDRIVE_DRIVE_ID = process.env.KDRIVE_DRIVE_ID;
const KDRIVE_FOLDER_ID = process.env.KDRIVE_FOLDER_ID;

const LOCAL_CONTENT_DIR = path.join(process.cwd(), 'src', 'content', 'rolls', 'synced');
const API_BASE = 'https://api.infomaniak.com/2/drive';
const MAX_RETRIES = 3;
const FETCH_TIMEOUT = 30000; // 30 seconds

interface KDriveFile {
  id: string;
  name: string;
  type: 'dir' | 'file';
  created_at: number;
}

interface KDriveResponse {
  data: KDriveFile[];
}

interface ImageMetadata {
  tags: string[];
  exif: {
    shutter?: string;
    aperture?: string;
    iso?: string;
    body?: string;
    lens?: string;
    focalLength?: string;
  };
  palette: string[];
}

interface ImageData {
  url: string;
  exif: ImageMetadata['exif'];
  poem?: string;
  palette: string[];
  dominantColor?: string;
}

interface RollToWrite {
  path: string;
  content: string;
  slug: string;
}

async function fetchWithRetry(url: string, options: RequestInit = {}, retries = MAX_RETRIES): Promise<Response> {
  const controller = new AbortController();
  const id = setTimeout(() => controller.abort(), FETCH_TIMEOUT);
  
  try {
    const response = await fetch(url, { ...options, signal: controller.signal });
    clearTimeout(id);
    
    if (!response.ok && retries > 0 && response.status >= 500) {
      console.warn(`   ⚠️ Erreur ${response.status}. Nouvelle tentative (${MAX_RETRIES - retries + 1}/${MAX_RETRIES})...`);
      return fetchWithRetry(url, options, retries - 1);
    }
    return response;
  } catch (error: any) {
    clearTimeout(id);
    if (retries > 0 && (error.name === 'AbortError' || error.name === 'FetchError' || error.code === 'ECONNRESET')) {
      console.warn(`   ⚠️ Erreur réseau (${error.message}). Nouvelle tentative (${MAX_RETRIES - retries + 1}/${MAX_RETRIES})...`);
      return fetchWithRetry(url, options, retries - 1);
    }
    throw error;
  }
}

async function fetchKDriveAPI(endpoint: string, options: RequestInit = {}): Promise<KDriveResponse> {
  const url = `${API_BASE}/${KDRIVE_DRIVE_ID}${endpoint}`;
  const response = await fetchWithRetry(url, {
    ...options,
    headers: { 
      'Authorization': `Bearer ${KDRIVE_API_TOKEN}`,
      'Content-Type': 'application/json',
      ...options.headers 
    },
  });
  if (!response.ok) {
    const errorBody = await response.text();
    throw new Error(`Erreur API kDrive: ${response.status} ${response.statusText} - ${errorBody}`);
  }
  return response.json();
}

async function fetchKDriveFileContent(fileId: string): Promise<string | null> {
  try {
    const downloadUrl = `${API_BASE}/${KDRIVE_DRIVE_ID}/files/${fileId}/download`;
    const response = await fetchWithRetry(downloadUrl, {
      headers: { 'Authorization': `Bearer ${KDRIVE_API_TOKEN}` },
    });
    if (!response.ok) return null;
    return response.text();
  } catch (e) {
    console.error(`   ❌ Erreur lors du téléchargement du fichier ${fileId}:`, e);
    return null;
  }
}

async function extractMetadataFromRemoteImage(fileId: string): Promise<ImageMetadata> {
  try {
    const downloadUrl = `${API_BASE}/${KDRIVE_DRIVE_ID}/files/${fileId}/download`;
    const response = await fetchWithRetry(downloadUrl, {
      headers: { 'Authorization': `Bearer ${KDRIVE_API_TOKEN}` },
    });
    
    if (!response.ok) {
      console.warn(`   ⚠️ Échec du téléchargement pour les métadonnées de ${fileId}`);
      return { tags: [], exif: {}, palette: [] };
    }

    const arrayBuffer = await response.arrayBuffer();
    const buffer = Buffer.from(arrayBuffer);
    
    let hexPalette: string[] = [];
    try {
      const palette = await Vibrant.from(buffer).getPalette();
      const colors = [palette.Vibrant, palette.DarkVibrant, palette.Muted, palette.LightMuted, palette.DarkMuted];
      hexPalette = colors.filter(c => c !== null).map(c => c!.hex);
    } catch (err) {
      console.warn(`   ⚠️ Impossible d'extraire la palette pour ${fileId}`);
    }
    
    let tags: string[] = [];
    let exifObj: ImageMetadata['exif'] = {};

    try {
      const metadata = await exifr.parse(buffer, { iptc: true, xmp: true, tiff: true, exif: true });
      if (metadata) {
        if (metadata.Keywords) {
          tags = Array.isArray(metadata.Keywords) ? metadata.Keywords : [metadata.Keywords];
        } else if (metadata.subject) {
          tags = Array.isArray(metadata.subject) ? metadata.subject : [metadata.subject];
        }

        if (metadata.Make && metadata.Model) {
          const make = metadata.Make.toString().replace(/Corporation/i, '').trim();
          const model = metadata.Model.toString().startsWith(make) ? metadata.Model : `${make} ${metadata.Model}`;
          exifObj.body = model;
        } else if (metadata.Model) {
          exifObj.body = metadata.Model;
        } else {
          exifObj.body = "Unknown Camera";
        }

        if (metadata.LensModel) {
          exifObj.lens = metadata.LensModel;
        } else if (metadata.Lens) {
          exifObj.lens = metadata.Lens;
        }

        if (metadata.FocalLength) exifObj.focalLength = `${metadata.FocalLength}mm`;
        if (metadata.FNumber) exifObj.aperture = `f/${metadata.FNumber}`;
        if (metadata.ISO) exifObj.iso = `ISO ${metadata.ISO}`;
        if (metadata.ExposureTime) {
          const speed = metadata.ExposureTime < 1 ? `1/${Math.round(1/metadata.ExposureTime)}s` : `${metadata.ExposureTime}s`;
          exifObj.shutter = speed;
        }
      }
    } catch (err) {
      console.warn(`   ⚠️ Impossible d'analyser l'EXIF pour ${fileId}`);
    }
    
    return { 
      tags: tags.map(tag => tag.toString().trim()), 
      exif: exifObj,
      palette: hexPalette
    };
  } catch (e) {
    console.error(`   ❌ Erreur critique lors de l'extraction des métadonnées (${fileId}):`, e);
    return { tags: [], exif: {}, palette: [] };
  }
}

async function getPublicUrl(fileId: string): Promise<string> {
  try {
    const shares = await fetchKDriveAPI(`/files/${fileId}/shares`);
    if (shares.data && shares.data.length > 0) {
      return shares.data[0].share_url;
    }

    const newShare = await fetchKDriveAPI(`/files/${fileId}/shares`, {
      method: 'POST',
      body: JSON.stringify({
        type: 'public',
        password_protected: false,
        expiration_date: 0
      })
    });
    return newShare.data.share_url;
  } catch (e) {
    console.warn(`   ⚠️ Impossible d'obtenir l'URL de partage pour ${fileId}.`);
    return '';
  }
}

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

async function sync() {
  const isDryRun = process.argv.includes('--dry-run');
  console.log(`🔄 Démarrage de la synchronisation (Mode kDrive Immersif)${isDryRun ? ' [DRY RUN]' : ''}...`);

  if (!KDRIVE_API_TOKEN || !KDRIVE_DRIVE_ID || !KDRIVE_FOLDER_ID) {
    console.error('⚠️ Variables d\'environnement manquantes. Vérifiez votre .env');
    return;
  }

  const stats = {
    rollsProcessed: 0,
    photosSynced: 0,
    errors: 0,
    skipped: 0
  };

  try {
    if (!isDryRun) {
      await fs.mkdir(LOCAL_CONTENT_DIR, { recursive: true });
    }

    const data = await fetchKDriveAPI(`/files/${KDRIVE_FOLDER_ID}/files`);
    const folders = data.data.filter((file: any) => file.type === 'dir');
    console.log(`📁 Trouvé ${folders.length} dossiers (Rolls) sur kDrive.`);

    const searchIndex: any[] = [];
    const rollsToWrite: Array<{path: string, content: string, slug: string}> = [];

    for (const folder of folders) {
      const rollTitle = folder.name;
      const rollSlug = slugify(rollTitle, { lower: true, strict: true });
      const rollDate = new Date(folder.created_at * 1000).toISOString().split('T')[0];
      
      console.log(`\n📸 Traitement du Roll: ${rollTitle}`);

      const folderData = await fetchKDriveAPI(`/files/${folder.id}/files`);
      
      const photos = folderData.data.filter((f: any) => 
        f.type === 'file' && (f.name.toLowerCase().endsWith('.jpg') || f.name.toLowerCase().endsWith('.jpeg'))
      );

      if (photos.length === 0) {
        console.warn(`   ⚠️ Aucun fichier image trouvé dans "${rollTitle}". On passe.`);
        stats.skipped++;
        continue;
      }

      // Recherche de poésie (.md) directement dans le dossier du Roll
      const poetryFile = folderData.data.find((f: any) => f.type === 'file' && f.name.toLowerCase().endsWith('.md'));
      let poetryData: { globalPoem?: string, photos: Record<string, string> } = { globalPoem: undefined, photos: {} };
      
      if (poetryFile) {
        console.log(`   📖 Poésie trouvée: ${poetryFile.name}`);
        const contentStr = await fetchKDriveFileContent(poetryFile.id);
        if (contentStr) {
          const { data: frontmatter, content } = matter(contentStr);
          poetryData = {
            globalPoem: content.trim() || undefined,
            photos: frontmatter.photos || {}
          };
        }
      }

      const audioFile = folderData.data.find((f: any) => f.type === 'file' && f.name.toLowerCase().endsWith('.mp3'));
      let audioUrl = undefined;
      if (audioFile) {
        console.log(`   🎵 Audio trouvé: ${audioFile.name}`);
        audioUrl = await getPublicUrl(audioFile.id);
      }

      let rollTags = new Set<string>();
      let allPalettes: string[][] = [];

      const imagesData = (await Promise.all(photos.map(async (photo: any) => {
        try {
          console.log(`   🔗 Traitement de: ${photo.name}...`);
          
          const publicUrl = await getPublicUrl(photo.id);
          if (!publicUrl) {
            console.warn(`   ⚠️ Pas d'URL publique pour ${photo.name}. Image ignorée.`);
            return null;
          }

          const { tags, exif, palette } = await extractMetadataFromRemoteImage(photo.id);
          
          tags.forEach(t => rollTags.add(t));
          
          if (palette.length > 0) {
            allPalettes.push(palette);
          }

          stats.photosSynced++;
          return {
            url: publicUrl,
            exif: exif,
            poem: poetryData.photos[photo.name] || undefined,
            palette: palette,
            dominantColor: palette.length > 0 ? palette[0] : undefined
          };
        } catch (err) {
          console.error(`   ❌ Erreur lors du traitement de la photo ${photo.name}:`, err);
          stats.errors++;
          return null;
        }
      }))).filter(img => img !== null);

      if (imagesData.length === 0) {
        console.warn(`   ⚠️ Aucune donnée d'image valide pour "${rollTitle}". On ne génère pas le fichier.`);
        stats.skipped++;
        continue;
      }

      const tagsArray = Array.from(rollTags);
      const rollPalette = calculateMeanPalette(allPalettes);
      
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
      rollsToWrite.push({ path: mdPath, content: mdContent, slug: rollSlug });
      stats.rollsProcessed++;
    }

    if (!isDryRun) {
      console.log('\n💾 Écriture des fichiers sur le disque...');
      for (const roll of rollsToWrite) {
        await fs.writeFile(roll.path, roll.content);
        console.log(`   ✅ Fichier généré: ${roll.slug}.md`);
      }

      // Sauvegarde de l'index de recherche global
      const indexPath = path.join(process.cwd(), 'public', 'search-index.json');
      await fs.writeFile(indexPath, JSON.stringify(searchIndex, null, 2));
      console.log(`\n🔍 Index de recherche généré: ${indexPath}`);
    } else {
      console.log('\n🧪 [DRY RUN] Aucun fichier n\'a été écrit.');
    }

    console.log('\n🎉 Synchronisation terminée !');
    console.log(`📊 Résumé: ${stats.rollsProcessed} Rolls, ${stats.photosSynced} Photos, ${stats.errors} Erreurs, ${stats.skipped} Ignorés.`);

  } catch (error) {
    console.error('\n❌ Erreur critique lors de la synchronisation:', error);
    process.exit(1);
  }
}

sync();
