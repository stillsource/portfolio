// Package imageanalyzer contains infrastructure adapters for service.ImageAnalyzer.
package imageanalyzer

import (
	"bytes"
	"fmt"
	"html"
	"image"

	// Side effect: register JPEG decoder with image.Decode.
	_ "image/jpeg"
	"kdrive-sync/pkg/domain"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/rwcarlsen/goexif/exif"
)

// Tuning knobs for the palette k-means routine.
const (
	defaultPaletteSize = 5
	defaultIterations  = 10
	maxPixelSamples    = 4096
)

// ExifKMeans extracts EXIF + XMP tags via goexif and derives a 5-color
// palette through k-means clustering in CIELAB space.
//
// A single struct owns the whole analysis so callers decode the JPEG
// exactly once per image.
type ExifKMeans struct {
	paletteSize int
	iterations  int
}

// NewExifKMeans returns an analyzer with sensible defaults (5 colors, 10 k-means iterations).
func NewExifKMeans() *ExifKMeans {
	return &ExifKMeans{
		paletteSize: defaultPaletteSize,
		iterations:  defaultIterations,
	}
}

// Analyze extracts every piece of derived metadata in a single pass.
//
// Partial failures degrade gracefully: a missing EXIF block or a broken
// JPEG decode produce a best-effort ImageAnalysis and no error unless all
// channels fail.
func (a *ExifKMeans) Analyze(data []byte) (domain.ImageAnalysis, error) {
	result := domain.ImageAnalysis{
		Tags: extractXMPKeywords(data),
		Exif: extractExif(data),
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return result, fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	result.Width = bounds.Dx()
	result.Height = bounds.Dy()

	palette, dominant := a.extractPalette(img)
	result.Palette = palette
	result.DominantColor = dominant
	return result, nil
}

// -----------------------------------------------------------------------------
// EXIF extraction
// -----------------------------------------------------------------------------

func extractExif(data []byte) domain.ExifData {
	x, err := exif.Decode(bytes.NewReader(data))
	if err != nil {
		return domain.ExifData{}
	}

	out := domain.ExifData{
		Body:        cameraBody(x),
		Lens:        tagString(x, exif.LensModel),
		FocalLength: focalLength(x),
		Aperture:    aperture(x),
		ISO:         iso(x),
		Shutter:     shutter(x),
	}
	return out
}

func cameraBody(x *exif.Exif) string {
	cameraMake := strings.TrimSpace(strings.ReplaceAll(tagString(x, exif.Make), "Corporation", ""))
	model := tagString(x, exif.Model)

	switch {
	case cameraMake == "" && model == "":
		return ""
	case cameraMake == "":
		return model
	case model == "":
		return cameraMake
	case strings.HasPrefix(model, cameraMake):
		return model
	default:
		return cameraMake + " " + model
	}
}

func focalLength(x *exif.Exif) string {
	num, den, ok := tagRat(x, exif.FocalLength)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%.0fmm", float64(num)/float64(den))
}

func aperture(x *exif.Exif) string {
	num, den, ok := tagRat(x, exif.FNumber)
	if !ok {
		return ""
	}
	return fmt.Sprintf("f/%.1f", float64(num)/float64(den))
}

func iso(x *exif.Exif) string {
	tag, err := x.Get(exif.ISOSpeedRatings)
	if err != nil {
		return ""
	}
	v, err := tag.Int(0)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("ISO %d", v)
}

func shutter(x *exif.Exif) string {
	num, den, ok := tagRat(x, exif.ExposureTime)
	if !ok {
		return ""
	}
	if num == 0 {
		return ""
	}
	if num < den {
		return fmt.Sprintf("1/%ds", den/num)
	}
	return fmt.Sprintf("%.1fs", float64(num)/float64(den))
}

func tagString(x *exif.Exif, name exif.FieldName) string {
	tag, err := x.Get(name)
	if err != nil {
		return ""
	}
	v, err := tag.StringVal()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(v)
}

func tagRat(x *exif.Exif, name exif.FieldName) (int64, int64, bool) {
	tag, err := x.Get(name)
	if err != nil {
		return 0, 0, false
	}
	num, den, err := tag.Rat2(0)
	if err != nil || den == 0 {
		return 0, 0, false
	}
	return num, den, true
}

// -----------------------------------------------------------------------------
// XMP keyword extraction
// -----------------------------------------------------------------------------

var (
	xmpSubjectBlockRE = regexp.MustCompile(`(?s)<dc:subject[^>]*>(.*?)</dc:subject>`)
	xmpRDFLiRE        = regexp.MustCompile(`<rdf:li[^>]*>([^<]*)</rdf:li>`)
)

// extractXMPKeywords scans the raw JPEG bytes for an embedded XMP packet and
// returns the dc:subject rdf:Bag entries (Lightroom/Darktable keywords).
func extractXMPKeywords(data []byte) []string {
	block := xmpSubjectBlockRE.FindSubmatch(data)
	if block == nil {
		return nil
	}
	matches := xmpRDFLiRE.FindAllSubmatch(block[1], -1)
	if len(matches) == 0 {
		return nil
	}

	tags := make([]string, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, m := range matches {
		tag := strings.TrimSpace(html.UnescapeString(string(m[1])))
		if tag == "" {
			continue
		}
		if _, dup := seen[tag]; dup {
			continue
		}
		seen[tag] = struct{}{}
		tags = append(tags, tag)
	}
	return tags
}

// -----------------------------------------------------------------------------
// Palette extraction (k-means in CIELAB)
// -----------------------------------------------------------------------------

func (a *ExifKMeans) extractPalette(img image.Image) ([]string, string) {
	samples := samplePixels(img)
	if len(samples) == 0 {
		return nil, ""
	}
	return a.kmeans(samples)
}

// samplePixels walks img on a uniform grid capped at maxPixelSamples to keep
// memory and CPU bounded regardless of image resolution.
func samplePixels(img image.Image) []colorful.Color {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return nil
	}

	stride := 1
	for (w/stride)*(h/stride) > maxPixelSamples {
		stride++
	}

	capacity := ((w / stride) + 1) * ((h / stride) + 1)
	samples := make([]colorful.Color, 0, capacity)
	for y := bounds.Min.Y; y < bounds.Max.Y; y += stride {
		for x := bounds.Min.X; x < bounds.Max.X; x += stride {
			c, ok := colorful.MakeColor(img.At(x, y))
			if ok {
				samples = append(samples, c)
			}
		}
	}
	return samples
}

// labPoint is a pre-computed CIELAB coordinate triple.
type labPoint struct{ l, a, b float64 }

// kmeans clusters pixels into a.paletteSize color groups using k-means in
// CIELAB space. It returns a hex palette sorted bright→dark and the dominant
// (largest-cluster) color.
//
// Pixels are converted to Lab once before the loop; distances use squared
// Euclidean (no sqrt needed for nearest-centroid comparison).
func (a *ExifKMeans) kmeans(pixels []colorful.Color) ([]string, string) {
	k := a.paletteSize
	if len(pixels) < k {
		k = len(pixels)
	}
	if k == 0 {
		return nil, ""
	}

	// Pre-convert all pixels to Lab — O(n) conversions instead of O(n·k·iter).
	labs := make([]labPoint, len(pixels))
	for i, p := range pixels {
		labs[i].l, labs[i].a, labs[i].b = p.Lab()
	}

	// Evenly-spaced seed centroids drawn directly from the Lab slice.
	centroids := make([]labPoint, k)
	for i := range centroids {
		centroids[i] = labs[(i*len(labs))/k+len(labs)/(2*k)]
	}

	acc := make([]labPoint, k)
	counts := make([]int, k)

	for range a.iterations {
		clear(acc)
		clear(counts)

		for _, p := range labs {
			best := 0
			bestDist := math.Inf(1)
			for i, c := range centroids {
				dl := p.l - c.l
				da := p.a - c.a
				db := p.b - c.b
				// Squared distance — ordering is preserved, sqrt is unnecessary.
				if d := dl*dl + da*da + db*db; d < bestDist {
					bestDist = d
					best = i
				}
			}
			acc[best].l += p.l
			acc[best].a += p.a
			acc[best].b += p.b
			counts[best]++
		}

		for i := range centroids {
			if n := float64(counts[i]); n > 0 {
				centroids[i] = labPoint{l: acc[i].l / n, a: acc[i].a / n, b: acc[i].b / n}
			}
		}
	}

	// Dominant = largest cluster.
	dominantIdx := 0
	for i, c := range counts {
		if c > counts[dominantIdx] {
			dominantIdx = i
		}
	}
	dominantHex := colorful.Lab(centroids[dominantIdx].l, centroids[dominantIdx].a, centroids[dominantIdx].b).Hex()

	// Sort bright → dark by CIELAB L* (already available, no roundtrip needed).
	sort.SliceStable(centroids, func(i, j int) bool { return centroids[i].l > centroids[j].l })

	palette := make([]string, 0, k)
	for _, c := range centroids {
		palette = append(palette, colorful.Lab(c.l, c.a, c.b).Hex())
	}
	return palette, dominantHex
}
