import type { ExifData } from '../../types/content';

// Builds the "BODY • LENS • FOCAL • APERTURE • SHUTTER • ISO" caption.
// Prefer `metadata` when present: the Go sync (src/scripts/go/kdrive-sync)
// pre-computes the exact same string via `ExifData.Caption()` so we can skip
// the join in the browser.
export function formatExif(exif?: ExifData, metadata?: string): string {
  if (typeof metadata === 'string' && metadata.length > 0) return metadata;
  if (!exif) return '';
  const parts: string[] = [];
  if (exif.body) parts.push(exif.body);
  if (exif.lens) parts.push(exif.lens);
  if (exif.focalLength) parts.push(exif.focalLength);
  if (exif.aperture) parts.push(exif.aperture);
  if (exif.shutter) parts.push(exif.shutter);
  if (exif.iso) parts.push(exif.iso);
  return parts.join(' • ');
}
