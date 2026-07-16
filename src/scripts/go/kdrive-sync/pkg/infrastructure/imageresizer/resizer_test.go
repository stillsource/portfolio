package imageresizer

import (
	"bytes"
	"image"
	"image/jpeg"
	"testing"
)

func makeJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	return buf.Bytes()
}

func decodeDims(t *testing.T, data []byte) (int, int) {
	t.Helper()
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode config: %v", err)
	}
	return cfg.Width, cfg.Height
}

func TestDerive(t *testing.T) {
	tests := []struct {
		name         string
		w, h         int
		wantW, wantH int
	}{
		{"landscape downscaled", 3000, 2000, 1920, 1280},
		{"portrait downscaled", 2000, 3000, 1280, 1920},
		{"small not upscaled", 800, 600, 800, 600},
		{"exactly at cap unchanged", 1920, 1080, 1920, 1080},
	}
	r := New(1920, 85)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := r.Derive(makeJPEG(t, tc.w, tc.h))
			if err != nil {
				t.Fatalf("Derive: %v", err)
			}
			gotW, gotH := decodeDims(t, out)
			if gotW != tc.wantW || gotH != tc.wantH {
				t.Errorf("dims = %dx%d, want %dx%d", gotW, gotH, tc.wantW, tc.wantH)
			}
		})
	}
}

func TestDeriveStripsMetadata(t *testing.T) {
	// Build a valid JPEG, then inject an APP1 "Exif" segment right after SOI.
	base := makeJPEG(t, 100, 100)
	app1 := []byte{0xFF, 0xE1, 0x00, 0x10, 'E', 'x', 'i', 'f', 0x00, 0x00, 1, 2, 3, 4, 5, 6, 7, 8}
	withExif := append([]byte{0xFF, 0xD8}, append(app1, base[2:]...)...)
	if !bytes.Contains(withExif, []byte("Exif\x00\x00")) {
		t.Fatal("fixture should contain the Exif marker")
	}

	out, err := New(1920, 85).Derive(withExif)
	if err != nil {
		t.Fatalf("Derive: %v", err)
	}
	if bytes.Contains(out, []byte("Exif\x00\x00")) {
		t.Error("derivative still contains Exif metadata")
	}
}

func TestDeriveRejectsGarbage(t *testing.T) {
	if _, err := New(1920, 85).Derive([]byte("not an image")); err == nil {
		t.Error("expected an error on undecodable input")
	}
}
