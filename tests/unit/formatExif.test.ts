import { describe, it, expect } from 'vitest';
import { formatExif } from '../../src/scripts/utils/formatExif';

describe('formatExif', () => {
  it('returns the pre-computed metadata string when present', () => {
    const out = formatExif({ body: 'X-T4' }, 'PRECOMPUTED');
    expect(out).toBe('PRECOMPUTED');
  });

  it('ignores an empty metadata string and falls back to formatting exif', () => {
    const out = formatExif({ body: 'X-T4', aperture: 'f/1.4' }, '');
    expect(out).toBe('X-T4 • f/1.4');
  });

  it('returns an empty string when no exif and no metadata', () => {
    expect(formatExif(undefined, undefined)).toBe('');
  });

  it('joins fields in a stable order (body, lens, focal, aperture, shutter, iso)', () => {
    const out = formatExif({
      body: 'FUJIFILM X-T4',
      lens: 'XF 35mm f/1.4 R',
      focalLength: '35mm',
      aperture: 'f/1.4',
      shutter: '1/60',
      iso: '6400',
    });
    expect(out).toBe('FUJIFILM X-T4 • XF 35mm f/1.4 R • 35mm • f/1.4 • 1/60 • 6400');
  });

  it('skips missing fields', () => {
    const out = formatExif({ body: 'X-T4', shutter: '1/200' });
    expect(out).toBe('X-T4 • 1/200');
  });
});
