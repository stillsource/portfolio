import { describe, it, expect } from 'vitest';
import { computeLoupeStyle } from '../../src/scripts/lightbox/loupe';

describe('computeLoupeStyle', () => {
  const rect = { left: 0, top: 0, right: 800, bottom: 600, width: 800, height: 600 };

  it('reports inside=false when the cursor is outside the rect', () => {
    expect(computeLoupeStyle({ rect, clientX: -1, clientY: 300, zoom: 2, diameter: 200 }))
      .toEqual({ inside: false });
    expect(computeLoupeStyle({ rect, clientX: 400, clientY: 601, zoom: 2, diameter: 200 }))
      .toEqual({ inside: false });
  });

  it('reports inside=false when the rect has zero dimensions', () => {
    const zero = { left: 0, top: 0, right: 0, bottom: 0, width: 0, height: 0 };
    expect(computeLoupeStyle({ rect: zero, clientX: 0, clientY: 0, zoom: 2, diameter: 200 }))
      .toEqual({ inside: false });
  });

  it('computes bgSize and bgPosition so the lens focuses on the cursor', () => {
    const out = computeLoupeStyle({ rect, clientX: 400, clientY: 300, zoom: 2, diameter: 200 });
    expect(out.inside).toBe(true);
    expect(out.left).toBe(400);
    expect(out.top).toBe(300);
    expect(out.bgSize).toBe('1600px 1200px');
    // rx=400, ry=300 → -(400*2 - 100)=-700, -(300*2 - 100)=-500
    expect(out.bgPosition).toBe('-700px -500px');
  });

  it('handles non-zero rect offsets', () => {
    const r = { left: 100, top: 50, right: 900, bottom: 650, width: 800, height: 600 };
    const out = computeLoupeStyle({ rect: r, clientX: 500, clientY: 350, zoom: 3, diameter: 300 });
    // rx = 500-100 = 400, ry = 350-50 = 300
    // bgPosition = -(400*3 - 150), -(300*3 - 150) = -1050, -750
    expect(out.bgPosition).toBe('-1050px -750px');
    expect(out.bgSize).toBe('2400px 1800px');
  });
});
