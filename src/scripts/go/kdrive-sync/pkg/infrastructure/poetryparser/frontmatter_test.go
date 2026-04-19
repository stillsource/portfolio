package poetryparser

import (
	"testing"
)

func TestFrontmatter_Parse(t *testing.T) {
	p := NewFrontmatter()

	t.Run("empty input", func(t *testing.T) {
		got, err := p.Parse([]byte{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.GlobalPoem != "" {
			t.Errorf("GlobalPoem = %q, want empty", got.GlobalPoem)
		}
		if len(got.PhotoPoems) != 0 {
			t.Errorf("PhotoPoems = %v, want empty map", got.PhotoPoems)
		}
	})

	t.Run("plain body without frontmatter", func(t *testing.T) {
		input := []byte("La lumière filtre\nentre les volets clos.")
		got, err := p.Parse(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "La lumière filtre\nentre les volets clos."
		if got.GlobalPoem != want {
			t.Errorf("GlobalPoem = %q, want %q", got.GlobalPoem, want)
		}
		if len(got.PhotoPoems) != 0 {
			t.Errorf("PhotoPoems should be empty, got %v", got.PhotoPoems)
		}
	})

	t.Run("frontmatter with photo poems and global body", func(t *testing.T) {
		input := []byte("---\nphotos:\n  IMG_0001.jpg: \"Premier vers\"\n  IMG_0002.jpg: \"Deuxième vers\"\n---\nPoème global du roll.")
		got, err := p.Parse(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.GlobalPoem != "Poème global du roll." {
			t.Errorf("GlobalPoem = %q, want %q", got.GlobalPoem, "Poème global du roll.")
		}
		if got.PhotoPoems["IMG_0001.jpg"] != "Premier vers" {
			t.Errorf("PhotoPoems[IMG_0001.jpg] = %q, want %q", got.PhotoPoems["IMG_0001.jpg"], "Premier vers")
		}
		if got.PhotoPoems["IMG_0002.jpg"] != "Deuxième vers" {
			t.Errorf("PhotoPoems[IMG_0002.jpg] = %q, want %q", got.PhotoPoems["IMG_0002.jpg"], "Deuxième vers")
		}
	})

	t.Run("frontmatter only, no body", func(t *testing.T) {
		input := []byte("---\nphotos:\n  IMG_0001.jpg: \"Seul\"\n---\n")
		got, err := p.Parse(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.GlobalPoem != "" {
			t.Errorf("GlobalPoem = %q, want empty", got.GlobalPoem)
		}
		if got.PhotoPoems["IMG_0001.jpg"] != "Seul" {
			t.Errorf("PhotoPoems[IMG_0001.jpg] = %q, want %q", got.PhotoPoems["IMG_0001.jpg"], "Seul")
		}
	})

	t.Run("missing closing delimiter treated as plain body", func(t *testing.T) {
		input := []byte("---\nphotos:\n  IMG_0001.jpg: \"Orphelin\"\n")
		got, err := p.Parse(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// No closing delimiter → entire content treated as GlobalPoem
		if got.GlobalPoem == "" {
			t.Error("GlobalPoem should not be empty when closing delimiter is missing")
		}
		if len(got.PhotoPoems) != 0 {
			t.Errorf("PhotoPoems should be empty, got %v", got.PhotoPoems)
		}
	})

	t.Run("body whitespace is trimmed", func(t *testing.T) {
		input := []byte("---\nphotos: {}\n---\n\n  \n  Le silence.\n  ")
		got, err := p.Parse(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.GlobalPoem != "Le silence." {
			t.Errorf("GlobalPoem = %q, want %q", got.GlobalPoem, "Le silence.")
		}
	})

	t.Run("empty photos map in frontmatter", func(t *testing.T) {
		input := []byte("---\nphotos: {}\n---\nUn poème sans photos.")
		got, err := p.Parse(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.GlobalPoem != "Un poème sans photos." {
			t.Errorf("GlobalPoem = %q, want %q", got.GlobalPoem, "Un poème sans photos.")
		}
		if len(got.PhotoPoems) != 0 {
			t.Errorf("PhotoPoems should be empty, got %v", got.PhotoPoems)
		}
	})

	t.Run("malformed YAML frontmatter returns error", func(t *testing.T) {
		// A stray tab inside mapping breaks the YAML 1.2 grammar used by
		// go-yaml; any invalid token will trigger the Unmarshal error path.
		input := []byte("---\nphotos:\n  - bad: [unterminated\n---\nbody")
		_, err := p.Parse(input)
		if err == nil {
			t.Fatal("expected error on malformed frontmatter, got nil")
		}
	})

	t.Run("leading whitespace before opening delimiter", func(t *testing.T) {
		// splitFrontmatter trims leading whitespace before checking for "---".
		input := []byte("\n\n---\nphotos:\n  a.jpg: \"v\"\n---\nbody")
		got, err := p.Parse(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.GlobalPoem != "body" {
			t.Errorf("GlobalPoem = %q, want %q", got.GlobalPoem, "body")
		}
		if got.PhotoPoems["a.jpg"] != "v" {
			t.Errorf("PhotoPoems[a.jpg] = %q, want %q", got.PhotoPoems["a.jpg"], "v")
		}
	})

	t.Run("delimiter without trailing newline treated as plain body", func(t *testing.T) {
		// "---" with no newline after it falls back to the plain-body path
		// because splitFrontmatter cannot find a line break.
		input := []byte("---")
		got, err := p.Parse(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.GlobalPoem != "---" {
			t.Errorf("GlobalPoem = %q, want %q", got.GlobalPoem, "---")
		}
	})

	t.Run("body-only input without trailing newline", func(t *testing.T) {
		// No frontmatter, no trailing newline — must still parse clean.
		input := []byte("une ligne unique")
		got, err := p.Parse(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.GlobalPoem != "une ligne unique" {
			t.Errorf("GlobalPoem = %q, want %q", got.GlobalPoem, "une ligne unique")
		}
	})
}
