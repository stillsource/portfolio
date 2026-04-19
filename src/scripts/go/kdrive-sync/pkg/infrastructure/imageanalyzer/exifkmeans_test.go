package imageanalyzer

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/rwcarlsen/goexif/exif"
)

// fixtureEXIF is decoded once in TestMain so every table-driven test reuses it.
var fixtureEXIF *exif.Exif

// fixtureBytes is the raw content of testdata/with_exif.jpg.
var fixtureBytes []byte

func TestMain(m *testing.M) {
	path := filepath.Join("testdata", "with_exif.jpg")
	data, err := os.ReadFile(path)
	if err != nil {
		// Let tests that need the fixture surface their own failure.
		os.Exit(m.Run())
	}
	fixtureBytes = data
	if x, err := exif.Decode(bytes.NewReader(data)); err == nil {
		fixtureEXIF = x
	}
	os.Exit(m.Run())
}

// -----------------------------------------------------------------------------
// extractXMPKeywords
// -----------------------------------------------------------------------------

func TestExtractXMPKeywords(t *testing.T) {
	t.Run("returns nil when no XMP block", func(t *testing.T) {
		got := extractXMPKeywords([]byte("no xmp here"))
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("extracts single keyword", func(t *testing.T) {
		data := xmpBytes(`<rdf:li>urbain</rdf:li>`)
		got := extractXMPKeywords(data)
		if len(got) != 1 || got[0] != "urbain" {
			t.Errorf("got %v, want [urbain]", got)
		}
	})

	t.Run("extracts multiple keywords", func(t *testing.T) {
		data := xmpBytes(`<rdf:li>nuit</rdf:li><rdf:li>tokyo</rdf:li><rdf:li>pluie</rdf:li>`)
		got := extractXMPKeywords(data)
		if len(got) != 3 {
			t.Fatalf("got %d keywords, want 3: %v", len(got), got)
		}
		want := []string{"nuit", "tokyo", "pluie"}
		for i, w := range want {
			if got[i] != w {
				t.Errorf("got[%d] = %q, want %q", i, got[i], w)
			}
		}
	})

	t.Run("deduplicates keywords", func(t *testing.T) {
		data := xmpBytes(`<rdf:li>nuit</rdf:li><rdf:li>nuit</rdf:li>`)
		got := extractXMPKeywords(data)
		if len(got) != 1 {
			t.Errorf("got %v, want deduped [nuit]", got)
		}
	})

	t.Run("trims whitespace from keywords", func(t *testing.T) {
		data := xmpBytes(`<rdf:li>  brume  </rdf:li>`)
		got := extractXMPKeywords(data)
		if len(got) != 1 || got[0] != "brume" {
			t.Errorf("got %v, want [brume]", got)
		}
	})

	t.Run("skips empty rdf:li entries", func(t *testing.T) {
		data := xmpBytes(`<rdf:li>nuit</rdf:li><rdf:li></rdf:li><rdf:li>   </rdf:li>`)
		got := extractXMPKeywords(data)
		if len(got) != 1 || got[0] != "nuit" {
			t.Errorf("got %v, want [nuit]", got)
		}
	})

	t.Run("returns nil when dc:subject block is empty", func(t *testing.T) {
		data := xmpBytes(``)
		got := extractXMPKeywords(data)
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("handles XML-escaped characters", func(t *testing.T) {
		data := xmpBytes(`<rdf:li>caf&amp;bar</rdf:li>`)
		got := extractXMPKeywords(data)
		if len(got) != 1 || got[0] != "caf&bar" {
			t.Errorf("got %v, want [caf&bar]", got)
		}
	})
}

// xmpBytes wraps inner content in a minimal XMP dc:subject block.
func xmpBytes(inner string) []byte {
	return []byte(`<?xpacket begin=""><x:xmpmeta><rdf:RDF><rdf:Description>` +
		`<dc:subject><rdf:Bag>` + inner + `</rdf:Bag></dc:subject>` +
		`</rdf:Description></rdf:RDF></x:xmpmeta>`)
}

// -----------------------------------------------------------------------------
// samplePixels
// -----------------------------------------------------------------------------

func TestSamplePixels(t *testing.T) {
	t.Run("zero-size image returns nil", func(t *testing.T) {
		img := image.NewNRGBA(image.Rect(0, 0, 0, 0))
		got := samplePixels(img)
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("small image returns all pixels", func(t *testing.T) {
		// 4×4 = 16 pixels, well under maxPixelSamples → stride=1, all sampled
		img := solidImage(4, 4, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		got := samplePixels(img)
		if len(got) != 16 {
			t.Errorf("got %d samples, want 16", len(got))
		}
	})

	t.Run("large image is downsampled", func(t *testing.T) {
		// 200×200 = 40 000 pixels > maxPixelSamples(4096) → stride > 1
		img := solidImage(200, 200, color.RGBA{R: 128, G: 64, B: 32, A: 255})
		got := samplePixels(img)
		if len(got) > maxPixelSamples {
			t.Errorf("got %d samples, want ≤ %d", len(got), maxPixelSamples)
		}
		if len(got) == 0 {
			t.Error("got 0 samples, expected at least some pixels")
		}
	})

	t.Run("1×1 image returns one sample", func(t *testing.T) {
		img := solidImage(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		got := samplePixels(img)
		if len(got) != 1 {
			t.Errorf("got %d samples, want 1", len(got))
		}
	})
}

// solidImage creates a w×h NRGBA image filled with a single color.
func solidImage(w, h int, c color.RGBA) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, c)
		}
	}
	return img
}

// -----------------------------------------------------------------------------
// kmeans (palette extraction)
// -----------------------------------------------------------------------------

func TestExifKMeans_ExtractPalette(t *testing.T) {
	a := NewExifKMeans()

	t.Run("empty image returns empty palette", func(t *testing.T) {
		img := image.NewNRGBA(image.Rect(0, 0, 0, 0))
		palette, dominant := a.extractPalette(img)
		if palette != nil {
			t.Errorf("got palette %v, want nil", palette)
		}
		if dominant != "" {
			t.Errorf("got dominant %q, want empty", dominant)
		}
	})

	t.Run("returns exactly paletteSize colors for large image", func(t *testing.T) {
		img := gradientImage(100, 100)
		palette, dominant := a.extractPalette(img)
		if len(palette) != defaultPaletteSize {
			t.Errorf("got %d colors, want %d", len(palette), defaultPaletteSize)
		}
		if dominant == "" {
			t.Error("dominant color should not be empty")
		}
	})

	t.Run("palette colors are valid hex", func(t *testing.T) {
		img := gradientImage(50, 50)
		palette, _ := a.extractPalette(img)
		for _, hex := range palette {
			if len(hex) != 7 || hex[0] != '#' {
				t.Errorf("invalid hex color %q", hex)
			}
			for _, ch := range hex[1:] {
				if !isHexChar(ch) {
					t.Errorf("non-hex character in %q", hex)
					break
				}
			}
		}
	})

	t.Run("image with fewer pixels than paletteSize", func(t *testing.T) {
		// 2×1 image → 2 pixels < paletteSize(5) → palette has ≤ 2 entries
		img := solidImage(2, 1, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		palette, _ := a.extractPalette(img)
		if len(palette) > 2 {
			t.Errorf("got %d colors for 2-pixel image, want ≤ 2", len(palette))
		}
	})
}

// gradientImage creates a w×h image with a smooth color gradient.
func gradientImage(w, h int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{
				R: uint8(x * 255 / w),
				G: uint8(y * 255 / h),
				B: 128,
				A: 255,
			})
		}
	}
	return img
}

func isHexChar(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

// -----------------------------------------------------------------------------
// Analyze — end-to-end with the JPEG fixture
// -----------------------------------------------------------------------------

func TestAnalyze(t *testing.T) {
	if len(fixtureBytes) == 0 {
		t.Skip("testdata/with_exif.jpg missing — generator did not run")
	}

	a := NewExifKMeans()
	got, err := a.Analyze(fixtureBytes)
	if err != nil {
		t.Fatalf("Analyze returned err: %v", err)
	}

	if len(got.Palette) == 0 {
		t.Error("expected non-empty palette")
	}
	if got.DominantColor == "" {
		t.Error("expected non-empty dominant color")
	}

	// EXIF fields populated via exiftool. cameraBody should collapse
	// "SONY Corporation" → "SONY" and prefix it to the model.
	if got.Exif.Body == "" {
		t.Error("expected Exif.Body to be populated")
	}
	if got.Exif.Lens == "" {
		t.Error("expected Exif.Lens to be populated")
	}
	if got.Exif.FocalLength == "" {
		t.Error("expected Exif.FocalLength to be populated")
	}
	if got.Exif.Aperture == "" {
		t.Error("expected Exif.Aperture to be populated")
	}
	if got.Exif.ISO == "" {
		t.Error("expected Exif.ISO to be populated")
	}
	if got.Exif.Shutter == "" {
		t.Error("expected Exif.Shutter to be populated")
	}

	// XMP keywords injected by exiftool.
	wantTags := map[string]bool{"nuit": true, "tokyo": true, "pluie": true}
	for _, tag := range got.Tags {
		delete(wantTags, tag)
	}
	if len(wantTags) > 0 {
		t.Errorf("missing XMP tags %v in %v", wantTags, got.Tags)
	}
}

func TestAnalyze_NoEXIF(t *testing.T) {
	// Encode a solid image in-memory — no EXIF block present.
	var buf bytes.Buffer
	img := solidImage(8, 8, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	got, err := NewExifKMeans().Analyze(buf.Bytes())
	if err != nil {
		t.Fatalf("Analyze returned err: %v", err)
	}
	// Palette must still be produced from the decoded pixels.
	if len(got.Palette) == 0 {
		t.Error("expected palette even without EXIF")
	}
	// EXIF fields must degrade to zero values.
	if !got.Exif.IsZero() {
		t.Errorf("expected zero EXIF, got %#v", got.Exif)
	}
	if len(got.Tags) != 0 {
		t.Errorf("expected no XMP tags, got %v", got.Tags)
	}
}

func TestAnalyze_InvalidBytes(t *testing.T) {
	_, err := NewExifKMeans().Analyze([]byte("not a jpeg"))
	if err == nil {
		t.Fatal("expected decode error for non-JPEG input")
	}
}

// -----------------------------------------------------------------------------
// EXIF helpers — use the fixture decoded once in TestMain.
// -----------------------------------------------------------------------------

func TestCameraBody(t *testing.T) {
	if fixtureEXIF == nil {
		t.Skip("fixture EXIF unavailable")
	}
	got := cameraBody(fixtureEXIF)
	if got == "" {
		t.Fatal("expected non-empty cameraBody from fixture")
	}
	// "SONY Corporation" → trimmed to "SONY" which is NOT a prefix of
	// "ILCE-7M3", so we expect the concatenated "SONY ILCE-7M3".
	if got != "SONY ILCE-7M3" {
		t.Errorf("cameraBody = %q, want %q", got, "SONY ILCE-7M3")
	}
}

// TestCameraBody_Branches exercises the remaining cameraBody switch arms via
// fixtures that carry a subset of Make/Model values.
func TestCameraBody_Branches(t *testing.T) {
	cases := []struct {
		name    string
		fixture string
		want    string
	}{
		// Only Make present → cameraBody returns Make alone.
		{name: "make only", fixture: "only_make.jpg", want: "SONY"},
		// Only Model present → cameraBody returns Model alone.
		{name: "model only", fixture: "only_model.jpg", want: "ILCE-7M3"},
		// Model starts with Make ("Canon" + "Canon EOS R5") → Model alone.
		{name: "model has make prefix", fixture: "prefix_model.jpg", want: "Canon EOS R5"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", tc.fixture))
			if err != nil {
				t.Skipf("fixture %s missing: %v", tc.fixture, err)
			}
			x, err := exif.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("decode fixture: %v", err)
			}
			if got := cameraBody(x); got != tc.want {
				t.Errorf("cameraBody = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestShutter_LongExposure covers the "num >= den" branch (e.g. 2s exposure)
// using the only_make.jpg fixture which has ExposureTime=2.
func TestShutter_LongExposure(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "only_make.jpg"))
	if err != nil {
		t.Skipf("fixture missing: %v", err)
	}
	x, err := exif.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	// The exact formatting depends on the stored rational — we just verify
	// it contains "s" (seconds suffix) and is non-empty.
	got := shutter(x)
	if got == "" || !bytes.HasSuffix([]byte(got), []byte("s")) {
		t.Errorf("shutter = %q, expected non-empty long-exposure string", got)
	}
}

// TestFocalLength_MissingTag and TestAperture_MissingTag assert the zero-return
// branches when the tag is absent (fixture `only_model.jpg` has neither).
func TestFocalLength_MissingTag(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "only_model.jpg"))
	if err != nil {
		t.Skipf("fixture missing: %v", err)
	}
	x, err := exif.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got := focalLength(x); got != "" {
		t.Errorf("focalLength(missing) = %q, want empty", got)
	}
	if got := aperture(x); got != "" {
		t.Errorf("aperture(missing) = %q, want empty", got)
	}
	if got := iso(x); got != "" {
		t.Errorf("iso(missing) = %q, want empty", got)
	}
	if got := shutter(x); got != "" {
		t.Errorf("shutter(missing) = %q, want empty", got)
	}
}

func TestFocalLength(t *testing.T) {
	if fixtureEXIF == nil {
		t.Skip("fixture EXIF unavailable")
	}
	got := focalLength(fixtureEXIF)
	if got != "35mm" {
		t.Errorf("focalLength = %q, want %q", got, "35mm")
	}
}

func TestAperture(t *testing.T) {
	if fixtureEXIF == nil {
		t.Skip("fixture EXIF unavailable")
	}
	got := aperture(fixtureEXIF)
	if got != "f/1.8" {
		t.Errorf("aperture = %q, want %q", got, "f/1.8")
	}
}

func TestISO(t *testing.T) {
	if fixtureEXIF == nil {
		t.Skip("fixture EXIF unavailable")
	}
	got := iso(fixtureEXIF)
	if got != "ISO 400" {
		t.Errorf("iso = %q, want %q", got, "ISO 400")
	}
}

func TestShutter(t *testing.T) {
	if fixtureEXIF == nil {
		t.Skip("fixture EXIF unavailable")
	}
	got := shutter(fixtureEXIF)
	if got != "1/250s" {
		t.Errorf("shutter = %q, want %q", got, "1/250s")
	}
}

func TestTagString(t *testing.T) {
	if fixtureEXIF == nil {
		t.Skip("fixture EXIF unavailable")
	}
	got := tagString(fixtureEXIF, exif.LensModel)
	if got != "FE 35mm F1.8" {
		t.Errorf("tagString(LensModel) = %q, want %q", got, "FE 35mm F1.8")
	}

	// Missing tag → empty string.
	if got := tagString(fixtureEXIF, exif.FieldName("NonExistentTag")); got != "" {
		t.Errorf("tagString(missing) = %q, want empty", got)
	}
}

func TestTagRat(t *testing.T) {
	if fixtureEXIF == nil {
		t.Skip("fixture EXIF unavailable")
	}
	// Valid rational tag.
	num, den, ok := tagRat(fixtureEXIF, exif.FocalLength)
	if !ok {
		t.Fatal("tagRat(FocalLength) should succeed")
	}
	if num == 0 || den == 0 {
		t.Errorf("tagRat returned zero values: num=%d den=%d", num, den)
	}

	// Missing tag → ok=false.
	if _, _, ok := tagRat(fixtureEXIF, exif.FieldName("NonExistentTag")); ok {
		t.Error("tagRat(missing) should return ok=false")
	}
}

// -----------------------------------------------------------------------------
// cameraBody — table-driven using synthetic EXIF via extractExif on encoded bytes.
//
// We still exercise the pure switch branches with a helper that builds a fake
// *exif.Exif is impractical (the type is opaque); instead we cover those
// branches indirectly by constructing minimal strings through the existing
// extractExif path is not necessary — the switch logic is simple. Skipping.
// -----------------------------------------------------------------------------

func TestExtractExif_NoEXIF(t *testing.T) {
	// Pure-JPEG bytes with no EXIF APP1 marker → zero struct.
	var buf bytes.Buffer
	img := solidImage(4, 4, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 70}); err != nil {
		t.Fatalf("encode: %v", err)
	}
	got := extractExif(buf.Bytes())
	if !got.IsZero() {
		t.Errorf("extractExif on EXIF-less JPEG = %#v, want zero", got)
	}
}
