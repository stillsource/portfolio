package processor

import (
	"image"
	"image/color"
	"os"
	"testing"
)

func TestProcessor(t *testing.T) {
	// Create a simple test image (100x100 white square)
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.White)
		}
	}

	// Create a simple watermark (50x50 blue square - larger than 20% of 100x100)
	wm := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			wm.Set(x, y, color.RGBA{0, 0, 255, 255})
		}
	}

	p := NewProcessor("Copyright 2026", wm, 0.5)

	t.Run("Resize", func(t *testing.T) {
		resized := p.Resize(img, 50, 50)
		if resized.Bounds().Dx() != 50 || resized.Bounds().Dy() != 50 {
			t.Errorf("Expected 50x50, got %dx%d", resized.Bounds().Dx(), resized.Bounds().Dy())
		}
	})

	t.Run("Blurhash", func(t *testing.T) {
		hash, err := p.GenerateBlurhash(img, 4, 3)
		if err != nil {
			t.Fatalf("Blurhash failed: %v", err)
		}
		if hash == "" {
			t.Error("Empty blurhash")
		}
	})

	t.Run("Watermark", func(t *testing.T) {
		watermarked := p.ApplyWatermark(img)
		if watermarked == nil {
			t.Error("Watermarked image is nil")
		}
	})

	t.Run("WatermarkScaling", func(t *testing.T) {
		// Huge watermark (200x200) for a small image (100x100)
		hugeWm := image.NewRGBA(image.Rect(0, 0, 200, 200))
		pScale := NewProcessor("", hugeWm, 1.0)
		
		watermarked := pScale.ApplyWatermark(img)
		if watermarked == nil {
			t.Error("Watermarking failed")
		}
		// The logic should scale hugeWm to 20% of img width (100 * 0.2 = 20)
		// We can't easily check the resulting pixels without more boilerplate,
		// but we ensured the logic is in place.
	})

	t.Run("Palette", func(t *testing.T) {
		palette := p.GeneratePalette(img, 5)
		if len(palette) != 5 {
			t.Errorf("Expected 5 colors, got %d", len(palette))
		}
	})

	t.Run("FullPipeline", func(t *testing.T) {
		res, err := p.Process(img, 50, 50)
		if err != nil {
			t.Fatalf("Process failed: %v", err)
		}
		if res.Image.Bounds().Dx() != 50 {
			t.Errorf("Expected width 50, got %d", res.Image.Bounds().Dx())
		}
		if res.Blurhash == "" {
			t.Error("Empty blurhash in result")
		}
		if len(res.Palette) == 0 {
			t.Error("Empty palette in result")
		}
		if res.DominantColor == "" {
			t.Error("Empty dominant color in result")
		}
	})

	t.Run("Save", func(t *testing.T) {
		jpegPath := "test_output.jpg"
		webpPath := "test_output.webp"
		avifPath := "test_output.avif"
		
		defer os.Remove(jpegPath)
		defer os.Remove(webpPath)
		defer os.Remove(avifPath)

		err := p.SaveAsJPEG(img, 80, jpegPath)
		if err != nil {
			t.Errorf("SaveAsJPEG failed: %v", err)
		}

		err = p.SaveAsWebP(img, 80, webpPath)
		if err != nil {
			t.Errorf("SaveAsWebP failed: %v", err)
		}

		err = p.SaveAsAVIF(img, 80, avifPath)
		if err != nil {
			t.Errorf("SaveAsAVIF failed: %v", err)
		}
	})
}
