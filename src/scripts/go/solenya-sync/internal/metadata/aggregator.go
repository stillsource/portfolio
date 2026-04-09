package metadata

import (
	"image"
	"sort"

	"github.com/lucasb-eyer/go-colorful"
)

// ClusterColors performs K-means clustering in CIELAB space.
// Returns the palette (sorted by luminance) and the dominant color (the largest cluster).
func ClusterColors(pixels []colorful.Color, numColors int) ([]string, string) {
	if len(pixels) == 0 {
		return nil, ""
	}
	if len(pixels) < numColors {
		res := make([]string, 0, len(pixels))
		for _, p := range pixels {
			res = append(res, p.Hex())
		}
		return res, res[0]
	}

	// K-means clustering in Lab space. 
	// Initialize centroids by sampling across the pixel range for better diversity.
	centroids := make([]colorful.Color, numColors)
	for i := range centroids {
		centroids[i] = pixels[(i*len(pixels))/numColors + (len(pixels)/(numColors*2))]
	}

	var counts []int
	for iter := 0; iter < 15; iter++ { // Increased to 15 iterations for Rick-level precision
		newCentroids := make([]struct{ l, a, b float64; count int }, numColors)
		for _, p := range pixels {
			bestDist, bestIdx := 1e18, 0
			for i, c := range centroids {
				// CIE76 distance in Lab space is perceptually uniform enough for this slop.
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
		counts = make([]int, numColors)
		for i := range centroids {
			if newCentroids[i].count > 0 {
				centroids[i] = colorful.Lab(
					newCentroids[i].l/float64(newCentroids[i].count),
					newCentroids[i].a/float64(newCentroids[i].count),
					newCentroids[i].b/float64(newCentroids[i].count),
				)
				counts[i] = newCentroids[i].count
			}
		}
	}

	// Dominant color is the most populous cluster.
	maxCount, dominantIdx := -1, 0
	for i, c := range counts {
		if c > maxCount {
			maxCount = c
			dominantIdx = i
		}
	}
	dominantHex := centroids[dominantIdx].Hex()

	// Palette is sorted by luminance (descending).
	sort.Slice(centroids, func(i, j int) bool {
		l1, _, _ := centroids[i].Lab()
		l2, _, _ := centroids[j].Lab()
		return l1 > l2
	})

	palette := make([]string, 0, numColors)
	for _, c := range centroids {
		palette = append(palette, c.Hex())
	}
	return palette, dominantHex
}

// GenerateMetadata extracts a palette and dominant color from a single image.
func GenerateMetadata(src image.Image, numColors int) ([]string, string) {
	pixels := ExtractPixels(src, 32)
	return ClusterColors(pixels, numColors)
}

// GeneratePalette extracts only the palette from a single image.
func GeneratePalette(src image.Image, numColors int) []string {
	palette, _ := GenerateMetadata(src, numColors)
	return palette
}

// ExtractPixels subsamples the image to a grid of step x step.
func ExtractPixels(src image.Image, step int) []colorful.Color {
	bounds := src.Bounds()
	stepX, stepY := bounds.Dx()/step, bounds.Dy()/step
	if stepX < 1 { stepX = 1 }
	if stepY < 1 { stepY = 1 }

	var pixels []colorful.Color
	for y := bounds.Min.Y; y < bounds.Max.Y; y += stepY {
		for x := bounds.Min.X; x < bounds.Max.X; x += stepX {
			c, ok := colorful.MakeColor(src.At(x, y))
			if ok {
				pixels = append(pixels, c)
			}
		}
	}
	return pixels
}

// AggregatePalette computes a roll-level palette from multiple images.
// Note: This loads all images into memory. For large rolls, use a streaming approach.
func AggregatePalette(images []image.Image, numColors int) ([]string, string) {
	var allPixels []colorful.Color
	for _, img := range images {
		allPixels = append(allPixels, ExtractPixels(img, 16)...)
	}
	return ClusterColors(allPixels, numColors)
}
