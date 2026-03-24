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

const LOCAL_CONTENT_DIR = path.join(process.cwd(), 'src', 'content', 'rolls');
const POETRY_DIR = path.join(process.cwd(), 'src', 'data', 'poetry');
const API_BASE = 'https://api.infomaniak.com/2/drive';

// ... (fetchKDriveAPI, extractMetadataFromRemoteImage, getPublicUrl remain unchanged)

async function getPoetryForRoll(rollSlug: string) {
  try {
    const poetryPath = path.join(POETRY_DIR, `${rollSlug}.md`);
    const fileContent = await fs.readFile(poetryPath, 'utf-8');
    const { data, content } = matter(fileContent);
    return {
      globalPoem: content.trim() || undefined,
      photos: data.photos || {}
    };
  } catch (e) {
    return { globalPoem: undefined, photos: {} };
  }
}

async function fetchKDriveAPI(endpoint: string) {
  const url = `${API_BASE}/${KDRIVE_DRIVE_ID}${endpoint}`;
  const response = await fetch(url, {
    headers: { 'Authorization': `Bearer ${KDRIVE_API_TOKEN}` },
  });
  if (!response.ok) throw new Error(`Erreur API kDrive: ${response.status} ${response.statusText}`);
  return response.json();
}

// Télécharge l'image uniquement en mémoire pour extraire les tags, EXIF et Palette
async function extractMetadataFromRemoteImage(fileId: string): Promise<{tags: string[], exif: string, palette: string[]}> {
  try {
    const downloadUrl = `${API_BASE}/${KDRIVE_DRIVE_ID}/files/${fileId}/download`;
    const response = await fetch(downloadUrl, {
      headers: { 'Authorization': `Bearer ${KDRIVE_API_TOKEN}` },
    });
    
    if (!response.ok) return { tags: [], exif: '', palette: [] };

    const arrayBuffer = await response.arrayBuffer();
    const buffer = Buffer.from(arrayBuffer);
    
    // Extraction Palette avec node-vibrant
    let hexPalette: string[] = [];
    try {
      const palette = await Vibrant.from(buffer).getPalette();
      // On récupère les couleurs disponibles dans l'ordre de pertinence
      const colors = [palette.Muted, palette.DarkMuted, palette.Vibrant, palette.LightMuted, palette.DarkVibrant];
      hexPalette = colors.filter(c => c !== null).map(c => c!.hex);
    } catch (err) { }
    
    // On parse tout : IPTC/XMP (pour les mots clés) et TIFF (pour les données de l'appareil)
    const metadata = await exifr.parse(buffer, { iptc: true, xmp: true, tiff: true, exif: true });
    
    let tags: string[] = [];
    let exifString = '';

    if (metadata) {
      if (metadata.Keywords) {
        tags = Array.isArray(metadata.Keywords) ? metadata.Keywords : [metadata.Keywords];
      } else if (metadata.subject) {
        tags = Array.isArray(metadata.subject) ? metadata.subject : [metadata.subject];
      }

      // Construction de la chaîne EXIF "Fujifilm X-T4 • 23mm • f/2.0 • 1/500s"
      const parts = [];
      if (metadata.Make && metadata.Model) {
        // Enlève la répétition de la marque si le modèle la contient déjà
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

// Fonction pour obtenir l'URL de partage public direct d'un fichier (dépend de kDrive)
// Dans une implémentation réelle, vous devez créer un lien de partage via l'API,
// ou si votre dossier racine est déjà public, construire l'URL directe.
async function getPublicUrl(fileId: string): Promise<string> {
  // EXEMPLE: Appel API pour générer ou récupérer un lien de partage direct
  /*
  const response = await fetchKDriveAPI(`/files/${fileId}/shares`);
  if (response.data && response.data.length > 0) {
    return response.data[0].url; // URL publique
  }
  */
  // En attendant l'API exacte pour les liens directs de votre compte, on simule une URL
  return `https://votre-domaine-kdrive.com/share/${fileId}.jpg`;
}

async function sync() {
  console.log('🔄 Démarrage de la synchronisation (Mode URLs Publiques + Poésie Markdown)...');

  if (!KDRIVE_API_TOKEN || !KDRIVE_DRIVE_ID || !KDRIVE_FOLDER_ID) {
    console.error('⚠️ Variables d\'environnement manquantes. Vérifiez votre .env');
    return;
  }

  try {
    await fs.mkdir(LOCAL_CONTENT_DIR, { recursive: true });

    const data = await fetchKDriveAPI(`/files/${KDRIVE_FOLDER_ID}/files`);
    const folders = data.data.filter((file: any) => file.type === 'dir');
    console.log(`      console.log(`📁 Trouvé ${folders.length} dossiers (Rolls) sur kDrive.`);

    for (const folder of folders) {
      const rollTitle = folder.name;
      const rollSlug = slugify(rollTitle, { lower: true, strict: true });
      const rollDate = new Date(folder.created_at * 1000).toISOString().split('T')[0];
      
      console.log(`\n📸 Traitement du Roll: ${rollTitle}`);

      // Charger la poésie pour ce Roll
      const poetry = await getPoetryForRoll(rollSlug);

      const folderData = await fetchKDriveAPI(`/files/${folder.id}/files`);
      
      // Chercher un fichier audio (MP3)
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

      // Traitement parallèle des images du roll
      const imagesData = await Promise.all(photos.map(async (photo: any) => {
        console.log(`   🔗 Traitement de: ${photo.name}...`);
        
        const publicUrl = await getPublicUrl(photo.id);
        const { tags, exif, palette } = await extractMetadataFromRemoteImage(photo.id);
        
        tags.forEach(t => rollTags.add(t));
        
        // On garde la palette de la première image rencontrée pour le roll entier si non définie
        if (rollPalette.length === 0 && palette.length > 0) {
          rollPalette = palette;
        }

        return {
          url: publicUrl,
          metadata: exif || undefined,
          poem: poetry.photos[photo.name] || undefined,
          palette: palette
        };
      }));

      // Générer le fichier Markdown
      const rollPoemAttr = poetry.globalPoem ? `poem: ${JSON.stringify(poetry.globalPoem)}\n` : '';
      const rollPaletteAttr = rollPalette.length > 0 ? `palette: ${JSON.stringify(rollPalette)}\n` : '';
      const rollAudioAttr = audioUrl ? `audioUrl: "${audioUrl}"\n` : '';
      
      const mdContent = `---
title: "${rollTitle}"
date: ${rollDate}
tags: ${JSON.stringify(Array.from(rollTags))}
${rollPoemAttr}${rollPaletteAttr}${rollAudioAttr}images: ${JSON.stringify(imagesData)}
---
`;
      const mdPath = path.join(LOCAL_CONTENT_DIR, `${rollSlug}.md`);
      await fs.writeFile(mdPath, mdContent);
      console.log(`   ✅ Fichier généré: ${rollSlug}.md (Tags: ${Array.from(rollTags).join(', ') || 'Aucun'})`);
    }

    console.log('\n🎉 Synchronisation terminée ! (Aucun fichier lourd stocké localement)');

  } catch (error) {
    console.error('\n❌ Erreur critique lors de la synchronisation:', error);
  }
}

sync();
