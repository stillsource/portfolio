package metadata

import (
	"image"
	"image/color"
	"testing"

	"github.com/lucasb-eyer/go-colorful"
)

func TestClusterColors(t *testing.T) {
	pixels := []colorful.Color{
		{R: 1, G: 0, B: 0}, // Red
		{R: 1, G: 0, B: 0},
		{R: 0, G: 1, B: 0}, // Green
		{R: 0, G: 1, B: 0},
		{R: 0, G: 0, B: 1}, // Blue
	}

	palette, dominant := ClusterColors(pixels, 3)
	if len(palette) != 3 {
		t.Errorf("Expected 3 colors, got %d", len(palette))
	}
	if dominant == "" {
		t.Error("Expected a dominant color, got empty string")
	}
}

func TestGeneratePalette(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.White)
		}
	}

	palette := GeneratePalette(img, 5)
	if len(palette) != 5 {
		t.Errorf("Expected 5 colors, got %d", len(palette))
	}
}

func TestAggregatePalette(t *testing.T) {
	img1 := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img2 := image.NewRGBA(image.Rect(0, 0, 10, 10))
	
	palette, dominant := AggregatePalette([]image.Image{img1, img2}, 5)
	if len(palette) != 5 {
		t.Errorf("Expected 5 colors, got %d", len(palette))
	}
	if dominant == "" {
		t.Error("Expected a dominant color, got empty string")
	}
}

func BenchmarkClusterColors(b *testing.B) {
	pixels := make([]colorful.Color, 1024)
	for i := range pixels {
		pixels[i] = colorful.FastWarmColor()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ClusterColors(pixels, 5)
	}
}
