package rollwriter

import (
	"kdrive-sync/pkg/domain"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMarkdown_WriteRoll(t *testing.T) {
	t.Run("creates file with correct name", func(t *testing.T) {
		dir := t.TempDir()
		w := NewMarkdown(dir)

		roll := minimalRoll()
		if err := w.WriteRoll("mon-roll", roll); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		path := filepath.Join(dir, "mon-roll.md")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist: %v", path, err)
		}
	})

	t.Run("creates output directory if missing", func(t *testing.T) {
		base := t.TempDir()
		dir := filepath.Join(base, "deep", "nested")
		w := NewMarkdown(dir)

		if err := w.WriteRoll("test", minimalRoll()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := os.Stat(filepath.Join(dir, "test.md")); err != nil {
			t.Errorf("expected file inside created dir: %v", err)
		}
	})

	t.Run("output is wrapped in YAML delimiters", func(t *testing.T) {
		dir := t.TempDir()
		w := NewMarkdown(dir)

		if err := w.WriteRoll("delimiters", minimalRoll()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := readFile(t, filepath.Join(dir, "delimiters.md"))
		if !strings.HasPrefix(data, "---\n") {
			t.Errorf("output should start with ---\\n, got: %q", data[:min(20, len(data))])
		}
		if !strings.HasSuffix(data, "---\n") {
			t.Errorf("output should end with ---\\n, got suffix: %q", data[max(0, len(data)-20):])
		}
	})

	t.Run("required fields are present in YAML", func(t *testing.T) {
		dir := t.TempDir()
		w := NewMarkdown(dir)

		roll := &domain.Roll{
			Title: "Nuit à Tokyo",
			Date:  "2024-03-15",
			Tags:  []string{"nuit", "urbain"},
		}
		if err := w.WriteRoll("nuit-a-tokyo", roll); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := readFile(t, filepath.Join(dir, "nuit-a-tokyo.md"))
		for _, want := range []string{"title: Nuit à Tokyo", "2024-03-15", "- nuit", "- urbain"} {
			if !strings.Contains(data, want) {
				t.Errorf("output missing %q\nfull output:\n%s", want, data)
			}
		}
	})

	t.Run("field order matches Astro schema", func(t *testing.T) {
		dir := t.TempDir()
		w := NewMarkdown(dir)

		roll := &domain.Roll{
			Title:         "Titre",
			Date:          "2024-01-01",
			Tags:          []string{"tag"},
			Poem:          "Un poème.",
			Palette:       []string{"#ffffff"},
			DominantColor: "#000000",
			AudioURL:      "https://example.com/audio.mp3",
			Images: []domain.Image{
				{URL: "https://example.com/img.jpg"},
			},
		}
		if err := w.WriteRoll("order-test", roll); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := readFile(t, filepath.Join(dir, "order-test.md"))
		fields := []string{"title:", "date:", "tags:", "poem:", "palette:", "dominantColor:", "audioUrl:", "images:"}
		assertFieldOrder(t, data, fields)
	})

	t.Run("optional fields omitted when empty", func(t *testing.T) {
		dir := t.TempDir()
		w := NewMarkdown(dir)

		if err := w.WriteRoll("minimal", minimalRoll()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := readFile(t, filepath.Join(dir, "minimal.md"))
		for _, absent := range []string{"poem:", "palette:", "dominantColor:", "audioUrl:", "videoUrl:"} {
			if strings.Contains(data, absent) {
				t.Errorf("output should not contain %q when field is empty\nfull output:\n%s", absent, data)
			}
		}
	})

	t.Run("image fields are serialized correctly", func(t *testing.T) {
		dir := t.TempDir()
		w := NewMarkdown(dir)

		roll := &domain.Roll{
			Title: "Avec images",
			Date:  "2024-06-01",
			Tags:  []string{},
			Images: []domain.Image{
				{
					URL: "https://example.com/photo.jpg",
					Alt: "Une photo",
					Exif: &domain.ExifData{
						Body:    "Sony A7",
						Shutter: "1/250s",
					},
				},
			},
		}
		if err := w.WriteRoll("avec-images", roll); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := readFile(t, filepath.Join(dir, "avec-images.md"))
		for _, want := range []string{
			"url: https://example.com/photo.jpg",
			"alt: Une photo",
			"body: Sony A7",
			"shutter: 1/250s",
		} {
			if !strings.Contains(data, want) {
				t.Errorf("output missing %q\nfull output:\n%s", want, data)
			}
		}
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		dir := t.TempDir()
		w := NewMarkdown(dir)

		first := &domain.Roll{Title: "Premier", Date: "2024-01-01", Tags: []string{}}
		if err := w.WriteRoll("overwrite", first); err != nil {
			t.Fatalf("unexpected error on first write: %v", err)
		}

		second := &domain.Roll{Title: "Second", Date: "2024-02-01", Tags: []string{}}
		if err := w.WriteRoll("overwrite", second); err != nil {
			t.Fatalf("unexpected error on second write: %v", err)
		}

		data := readFile(t, filepath.Join(dir, "overwrite.md"))
		if strings.Contains(data, "Premier") {
			t.Error("overwrite failed: first title still present")
		}
		if !strings.Contains(data, "Second") {
			t.Error("overwrite failed: second title not found")
		}
	})

	t.Run("returns error when OutDir is an existing file", func(t *testing.T) {
		// Point outDir at a regular file so os.MkdirAll must fail.
		base := t.TempDir()
		filePath := filepath.Join(base, "blocker")
		if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
			t.Fatalf("setup: %v", err)
		}
		w := NewMarkdown(filePath)
		err := w.WriteRoll("anything", minimalRoll())
		if err == nil {
			t.Fatal("expected error when OutDir is a file, got nil")
		}
		if !strings.Contains(err.Error(), "mkdir") {
			t.Errorf("expected mkdir error, got: %v", err)
		}
	})

	t.Run("returns error when target directory is read-only", func(t *testing.T) {
		// Skip when running as root (root bypasses file permissions on Linux).
		if os.Geteuid() == 0 {
			t.Skip("running as root — permission bits are ignored")
		}
		dir := t.TempDir()
		// Ensure the directory exists before locking it down.
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		if err := os.Chmod(dir, 0o500); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		// Restore perms so t.TempDir cleanup can remove the directory.
		t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

		w := NewMarkdown(dir)
		err := w.WriteRoll("locked", minimalRoll())
		if err == nil {
			t.Fatal("expected error writing to read-only dir, got nil")
		}
		if !strings.Contains(err.Error(), "write roll") {
			t.Errorf("expected 'write roll' wrapping, got: %v", err)
		}
	})
}

// minimalRoll returns a Roll with only mandatory fields set.
func minimalRoll() *domain.Roll {
	return &domain.Roll{
		Title:  "Minimal",
		Date:   "2024-01-01",
		Tags:   []string{},
		Images: []domain.Image{},
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}
	return string(data)
}

// assertFieldOrder verifies that each field in fields appears in the data in
// the given order (earlier index → earlier position in text).
func assertFieldOrder(t *testing.T, data string, fields []string) {
	t.Helper()
	prev := -1
	for _, f := range fields {
		idx := strings.Index(data, f)
		if idx == -1 {
			t.Errorf("field %q not found in output", f)
			continue
		}
		if idx <= prev {
			t.Errorf("field %q appears before expected predecessor (pos %d <= %d)", f, idx, prev)
		}
		prev = idx
	}
}
