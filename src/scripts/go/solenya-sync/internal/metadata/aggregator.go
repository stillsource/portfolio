package metadata

import (
	"image"
	"sort"

	"github.com/lucasb-eyer/go-colorful"
)

// ClusterColors performs K-means clustering in CIELAB space.
func ClusterColors(pixels []colorful.Color, numColors int) []string {
	if len(pixels) == 0 {
		return nil
	}
	if len(pixels) < numColors {
		res := make([]string, 0, len(pixels))
		for _, p := range pixels {
			res = append(res, p.Hex())
		}
		return res
	}

	// K-means clustering in Lab space
	centroids := make([]colorful.Color, numColors)
	for i := range centroids {
		centroids[i] = pixels[(i*len(pixels))/numColors]
	}

	for iter := 0; iter < 10; iter++ { // 10 iterations for better precision than the original 5
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

	// Sort centroids by luminance (descending) for a pleasing palette
	sort.Slice(centroids, func(i, j int) bool {
		l1, _, _ := centroids[i].Lab()
		l2, _, _ := centroids[j].Lab()
		return l1 > l2
	})

	res := make([]string, 0, numColors)
	for _, c := range centroids {
		res = append(res, c.Hex())
	}
	return res
}

// GeneratePalette extracts a palette from a single image.
func GeneratePalette(src image.Image, numColors int) []string {
	pixels := ExtractPixels(src, 32)
	return ClusterColors(pixels, numColors)
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
			c, _ := colorful.MakeColor(src.At(x, y))
			pixels = append(pixels, c)
		}
	}
	return pixels
}

// AggregatePalette computes a roll-level palette from multiple images.
func AggregatePalette(images []image.Image, numColors int) []string {
	var allPixels []colorful.Color
	for _, img := range images {
		allPixels = append(allPixels, ExtractPixels(img, 16)...) // Use lower density for aggregation to save memory
	}
	return ClusterColors(allPixels, numColors)
}
