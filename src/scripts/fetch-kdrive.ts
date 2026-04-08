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

async function fetchKDriveAPI(endpoint: string, options: any = {}) {
  const url = `${API_BASE}/${KDRIVE_DRIVE_ID}${endpoint}`;
  const response = await fetch(url, {
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

async function fetchKDriveFileContent(fileId: string): Promise<string> {
  const downloadUrl = `${API_BASE}/${KDRIVE_DRIVE_ID}/files/${fileId}/download`;
  const response = await fetch(downloadUrl, {
    headers: { 'Authorization': `Bearer ${KDRIVE_API_TOKEN}` },
  });
  if (!response.ok) return '';
  return response.text();
}

async function extractMetadataFromRemoteImage(fileId: string): Promise<{tags: string[], exif: string, palette: string[]}> {
  try {
    const downloadUrl = `${API_BASE}/${KDRIVE_DRIVE_ID}/files/${fileId}/download`;
    const response = await fetch(downloadUrl, {
      headers: { 'Authorization': `Bearer ${KDRIVE_API_TOKEN}` },
    });
    
    if (!response.ok) return { tags: [], exif: '', palette: [] };

    const arrayBuffer = await response.arrayBuffer();
    const buffer = Buffer.from(arrayBuffer);
    
    let hexPalette: string[] = [];
    try {
      const palette = await Vibrant.from(buffer).getPalette();
      const colors = [palette.Muted, palette.DarkMuted, palette.Vibrant, palette.LightMuted, palette.DarkVibrant];
      hexPalette = colors.filter(c => c !== null).map(c => c!.hex);
    } catch (err) { }
    
    const metadata = await exifr.parse(buffer, { iptc: true, xmp: true, tiff: true, exif: true });
    
    let tags: string[] = [];
    let exifString = '';

    if (metadata) {
      if (metadata.Keywords) {
        tags = Array.isArray(metadata.Keywords) ? metadata.Keywords : [metadata.Keywords];
      } else if (metadata.subject) {
        tags = Array.isArray(metadata.subject) ? metadata.subject : [metadata.subject];
      }

      const parts = [];
      if (metadata.Make && metadata.Model) {
        const make = metadata.Make.toString().replace(/Corporation/i, '').trim();
        const model = metadata.Model.toString().startsWith(make) ? metadata.Model : `${make} ${metadata.Model}`;
        parts.push(model);
      } else if (metadata.Model) {
        parts.push(metadata.Model);
      }

      if (metadata.FocalLength) parts.push(`${metadata.FocalLength}mm`);
      if (metadata.FNumber) parts.push(`f/${metadata.FNumber}`);
      if (metadata.ExposureTime) {
        const speed = metadata.ExposureTime < 1 ? `1/${Math.round(1/metadata.ExposureTime)}s` : `${metadata.ExposureTime}s`;
        parts.push(speed);
      }
      
      exifString = parts.join(' • ');
    }
    
    return { 
      tags: tags.map(tag => tag.toString().trim()), 
      exif: exifString,
      palette: hexPalette
    };
  } catch (e) {
    console.log("   ⚠️ Impossible d'extraire les métadonnées de cette image.");
    return { tags: [], exif: '', palette: [] };
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
    console.log(`   ⚠️ Impossible d'obtenir l'URL de partage pour ${fileId}.`);
    return '';
  }
}

async function sync() {
  console.log('🔄 Démarrage de la synchronisation (Mode kDrive Immersif)...');

  if (!KDRIVE_API_TOKEN || !KDRIVE_DRIVE_ID || !KDRIVE_FOLDER_ID) {
    console.error('⚠️ Variables d\'environnement manquantes. Vérifiez votre .env');
    return;
  }

  try {
    await fs.mkdir(LOCAL_CONTENT_DIR, { recursive: true });

    const data = await fetchKDriveAPI(`/files/${KDRIVE_FOLDER_ID}/files`);
    const folders = data.data.filter((file: any) => file.type === 'dir');
    console.log(`📁 Trouvé ${folders.length} dossiers (Rolls) sur kDrive.`);

    const searchIndex: any[] = [];

    for (const folder of folders) {
      const rollTitle = folder.name;
      const rollSlug = slugify(rollTitle, { lower: true, strict: true });
      const rollDate = new Date(folder.created_at * 1000).toISOString().split('T')[0];
      
      console.log(`\n📸 Traitement du Roll: ${rollTitle}`);

      const folderData = await fetchKDriveAPI(`/files/${folder.id}/files`);
      
      // Recherche de poésie (.md) directement dans le dossier du Roll
      const poetryFile = folderData.data.find((f: any) => f.type === 'file' && f.name.toLowerCase().endsWith('.md'));
      let poetryData = { globalPoem: undefined, photos: {} };
      
      if (poetryFile) {
        console.log(`   📖 Poésie trouvée: ${poetryFile.name}`);
        const contentStr = await fetchKDriveFileContent(poetryFile.id);
        const { data: frontmatter, content } = matter(contentStr);
        poetryData = {
          globalPoem: content.trim() || undefined,
          photos: frontmatter.photos || {}
        };
      }

      const audioFile = folderData.data.find((f: any) => f.type === 'file' && f.name.toLowerCase().endsWith('.mp3'));
      let audioUrl = undefined;
      if (audioFile) {
        console.log(`   🎵 Audio trouvé: ${audioFile.name}`);
        audioUrl = await getPublicUrl(audioFile.id);
      }

      const photos = folderData.data.filter((f: any) => 
        f.type === 'file' && (f.name.toLowerCase().endsWith('.jpg') || f.name.toLowerCase().endsWith('.jpeg'))
      );

      let rollTags = new Set<string>();
      let rollPalette: string[] = [];

      const imagesData = await Promise.all(photos.map(async (photo: any) => {
        console.log(`   🔗 Traitement de: ${photo.name}...`);
        
        const publicUrl = await getPublicUrl(photo.id);
        const { tags, exif, palette } = await extractMetadataFromRemoteImage(photo.id);
        
        tags.forEach(t => rollTags.add(t));
        
        if (rollPalette.length === 0 && palette.length > 0) {
          rollPalette = palette;
        }

        return {
          url: publicUrl,
          metadata: exif || undefined,
          poem: poetryData.photos[photo.name] || undefined,
          palette: palette,
          dominantColor: palette.length > 0 ? palette[0] : undefined
        };
      }));

      const tagsArray = Array.from(rollTags);
      
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

    // Sauvegarde de l'index de recherche global
    const indexPath = path.join(process.cwd(), 'public', 'search-index.json');
    await fs.writeFile(indexPath, JSON.stringify(searchIndex, null, 2));
    console.log(`\n🔍 Index de recherche généré: ${indexPath}`);

    console.log('\n🎉 Synchronisation terminée !');

  } catch (error) {
    console.error('\n❌ Erreur critique lors de la synchronisation:', error);
  }
}

sync();
