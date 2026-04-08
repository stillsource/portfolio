package processor

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"

	"github.com/bbrks/go-blurhash"
	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/gen2brain/avif"
	"github.com/lucasb-eyer/go-colorful"
)

type Processor struct {
	WatermarkText  string
	WatermarkImage image.Image
	Opacity        float64
}

func NewProcessor(text string, img image.Image, opacity float64) *Processor {
	return &Processor{text, img, opacity}
}

func (p *Processor) LoadImage(path string) (image.Image, error) {
	return imaging.Open(path)
}

type ProcessResult struct {
	Image    image.Image
	Blurhash string
	Palette  []string
}

func (p *Processor) Process(src image.Image, w, h int) (*ProcessResult, error) {
	// 1. Resize
	img := p.Resize(src, w, h)
	
	// 2. Generate Metadata (Blurhash & Palette) on the resized image for performance
	hash, err := p.GenerateBlurhash(img, 4, 3)
	if err != nil {
		return nil, err
	}
	palette := p.GeneratePalette(img, 5)

	// 3. Apply Watermark (last step before saving)
	final := p.ApplyWatermark(img)

	return &ProcessResult{
		Image:    final,
		Blurhash: hash,
		Palette:  palette,
	}, nil
}

func (p *Processor) Resize(src image.Image, w, h int) image.Image {
	return imaging.Resize(src, w, h, imaging.Lanczos)
}

func (p *Processor) ApplyWatermark(src image.Image) image.Image {
	if p.WatermarkImage == nil {
		return src
	}
	sb, wb := src.Bounds(), p.WatermarkImage.Bounds()
	
	targetW := sb.Dx() / 5
	wm := p.WatermarkImage
	if wb.Dx() > targetW {
		wm = imaging.Resize(wm, targetW, 0, imaging.Lanczos)
		wb = wm.Bounds()
	}

	x, y := sb.Dx()-wb.Dx()-20, sb.Dy()-wb.Dy()-20
	if x < 0 { x = 0 }
	if y < 0 { y = 0 }

	op := p.Opacity
	if op < 0 { op = 0 }
	if op > 1 { op = 1 }

	return imaging.Overlay(src, wm, image.Pt(x, y), op)
}

func (p *Processor) GenerateBlurhash(src image.Image, x, y int) (string, error) {
	if x < 1 { x = 1 }
	if y < 1 { y = 1 }
	h, err := blurhash.Encode(x, y, src)
	if err != nil {
		return "", fmt.Errorf("blurhash failed: %w", err)
	}
	return h, nil
}

func (p *Processor) GeneratePalette(src image.Image, numColors int) []string {
	bounds := src.Bounds()
	stepX, stepY := bounds.Dx()/32, bounds.Dy()/32
	if stepX < 1 { stepX = 1 }
	if stepY < 1 { stepY = 1 }

	var pixels []colorful.Color
	for y := bounds.Min.Y; y < bounds.Max.Y; y += stepY {
		for x := bounds.Min.X; x < bounds.Max.X; x += stepX {
			c, _ := colorful.MakeColor(src.At(x, y))
			pixels = append(pixels, c)
		}
	}

	if len(pixels) == 0 { return nil }

	// K-means clustering in Lab space
	centroids := make([]colorful.Color, numColors)
	for i := range centroids {
		centroids[i] = pixels[(i*len(pixels))/numColors]
	}

	for iter := 0; iter < 5; iter++ { // 5 iterations is usually enough for a good-enough palette
		newCentroids := make([]struct{ l, a, b float64; count int }, numColors)
		for _, p := range pixels {
			bestDist, bestIdx := 1e18, 0
			for i, c := range centroids {
				d := p.DistanceLab(c)
				if d < bestDist {
					bestDist, bestIdx = d, i
				}
			}
			l, a, b := p.Lab()
			newCentroids[bestIdx].l += l
			newCentroids[bestIdx].a += a
			newCentroids[bestIdx].b += b
			newCentroids[bestIdx].count++
		}
		for i := range centroids {
			if newCentroids[i].count > 0 {
				centroids[i] = colorful.Lab(
					newCentroids[i].l/float64(newCentroids[i].count),
					newCentroids[i].a/float64(newCentroids[i].count),
					newCentroids[i].b/float64(newCentroids[i].count),
				)
			}
		}
	}

	res := make([]string, 0, numColors)
	for _, c := range centroids {
		res = append(res, c.Hex())
	}
	return res
}

func (p *Processor) SaveAsJPEG(img image.Image, q int, path string) error {
	if q < 1 { q = 1 }
	if q > 100 { q = 100 }
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{Quality: q})
}

func (p *Processor) SaveAsWebP(img image.Image, q float32, path string) error {
	if q < 0 { q = 0 }
	if q > 100 { q = 100 }
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return webp.Encode(f, img, &webp.Options{Lossless: false, Quality: q})
}

func (p *Processor) SaveAsAVIF(img image.Image, q int, path string) error {
	if q < 0 { q = 0 }
	if q > 100 { q = 100 }
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return avif.Encode(f, img, avif.Options{Quality: q})
}
