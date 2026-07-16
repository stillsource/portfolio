// Package imageresizer produces web-safe JPEG derivatives: downscaled to a
// bounded long edge and re-encoded so all EXIF/IPTC/XMP metadata is dropped.
package imageresizer

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"

	"golang.org/x/image/draw"
)

// Resizer derives bounded-size, metadata-free JPEGs from arbitrary image bytes.
type Resizer struct {
	MaxLongEdge int
	Quality     int
}

// New returns a Resizer capping the long edge at maxLongEdge px and encoding
// JPEG at the given quality (1-100).
func New(maxLongEdge, quality int) *Resizer {
	return &Resizer{MaxLongEdge: maxLongEdge, Quality: quality}
}

// Derive decodes data, downscales it so its long edge is at most MaxLongEdge
// (aspect ratio preserved, never upscaled), and re-encodes it as JPEG at
// Quality. Re-encoding drops all EXIF/IPTC/XMP metadata.
func (r *Resizer) Derive(data []byte) ([]byte, error) {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	tw, th := targetDims(w, h, r.MaxLongEdge)

	out := src
	if tw != w || th != h {
		dst := image.NewRGBA(image.Rect(0, 0, tw, th))
		draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)
		out = dst
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, out, &jpeg.Options{Quality: r.Quality}); err != nil {
		return nil, fmt.Errorf("encode jpeg: %w", err)
	}
	return buf.Bytes(), nil
}

// targetDims scales (w,h) so the long edge is at most maxLong, preserving the
// aspect ratio. It never upscales: inputs already within the cap are unchanged.
func targetDims(w, h, maxLong int) (int, int) {
	long := w
	if h > w {
		long = h
	}
	if long <= maxLong {
		return w, h
	}
	ratio := float64(maxLong) / float64(long)
	nw := max(int(float64(w)*ratio+0.5), 1)
	nh := max(int(float64(h)*ratio+0.5), 1)
	return nw, nh
}
