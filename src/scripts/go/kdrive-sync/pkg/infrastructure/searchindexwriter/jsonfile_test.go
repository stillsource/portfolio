package searchindexwriter

import (
	"encoding/json"
	"kdrive-sync/pkg/domain"
	"os"
	"path/filepath"
	"testing"
)

func TestJSONFile_WriteIndex(t *testing.T) {
	t.Run("creates file at given path", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "search-index.json")
		w := NewJSONFile(path)

		if err := w.WriteIndex([]domain.SearchIndexItem{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist: %v", path, err)
		}
	})

	t.Run("creates parent directories if missing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "public", "search-index.json")
		w := NewJSONFile(path)

		if err := w.WriteIndex([]domain.SearchIndexItem{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file in nested dir: %v", err)
		}
	})

	t.Run("empty list writes JSON array", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "index.json")
		w := NewJSONFile(path)

		if err := w.WriteIndex([]domain.SearchIndexItem{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := readFile(t, path)
		var items []domain.SearchIndexItem
		if err := json.Unmarshal([]byte(data), &items); err != nil {
			t.Fatalf("output is not valid JSON: %v\n%s", err, data)
		}
		if len(items) != 0 {
			t.Errorf("got %d items, want 0", len(items))
		}
	})

	t.Run("items serialized with correct field names", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "index.json")
		w := NewJSONFile(path)

		items := []domain.SearchIndexItem{
			{
				ID:      "nuit-a-tokyo",
				Title:   "Nuit à Tokyo",
				Date:    "2024-03-15",
				Tags:    []string{"nuit", "urbain"},
				Poem:    "Le néon tremble.",
				Cover:   "https://example.com/cover.jpg",
				Palette: []string{"#1a1a2e", "#16213e"},
			},
		}
		if err := w.WriteIndex(items); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := readFile(t, path)
		var got []map[string]any
		if err := json.Unmarshal([]byte(data), &got); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("got %d items, want 1", len(got))
		}

		item := got[0]
		assertString(t, item, "id", "nuit-a-tokyo")
		assertString(t, item, "title", "Nuit à Tokyo")
		assertString(t, item, "date", "2024-03-15")
		assertString(t, item, "poem", "Le néon tremble.")
		assertString(t, item, "cover", "https://example.com/cover.jpg")
	})

	t.Run("poem omitted when empty", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "index.json")
		w := NewJSONFile(path)

		items := []domain.SearchIndexItem{
			{ID: "slug", Title: "T", Date: "2024-01-01", Tags: []string{}, Palette: []string{}},
		}
		if err := w.WriteIndex(items); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := readFile(t, path)
		var got []map[string]any
		if err := json.Unmarshal([]byte(data), &got); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if _, ok := got[0]["poem"]; ok {
			t.Error("empty poem should be omitted from JSON (omitempty)")
		}
	})

	t.Run("output is pretty-printed", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "index.json")
		w := NewJSONFile(path)

		if err := w.WriteIndex([]domain.SearchIndexItem{
			{ID: "x", Title: "X", Date: "2024-01-01", Tags: []string{}, Palette: []string{}},
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data := readFile(t, path)
		// Pretty-printed JSON has newlines
		if len(data) == 0 {
			t.Fatal("empty output")
		}
		var found bool
		for _, ch := range data {
			if ch == '\n' {
				found = true
				break
			}
		}
		if !found {
			t.Error("output is not pretty-printed (no newlines found)")
		}
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "index.json")
		w := NewJSONFile(path)

		first := []domain.SearchIndexItem{{ID: "first", Title: "First", Date: "2024-01-01", Tags: []string{}, Palette: []string{}}}
		if err := w.WriteIndex(first); err != nil {
			t.Fatalf("first write error: %v", err)
		}

		second := []domain.SearchIndexItem{{ID: "second", Title: "Second", Date: "2024-02-01", Tags: []string{}, Palette: []string{}}}
		if err := w.WriteIndex(second); err != nil {
			t.Fatalf("second write error: %v", err)
		}

		data := readFile(t, path)
		var got []domain.SearchIndexItem
		if err := json.Unmarshal([]byte(data), &got); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if len(got) != 1 || got[0].ID != "second" {
			t.Errorf("overwrite failed: got %v", got)
		}
	})

	t.Run("returns error when parent exists as a file", func(t *testing.T) {
		base := t.TempDir()
		// Create a regular file at the location the writer will treat as the
		// parent directory; MkdirAll must fail.
		blocker := filepath.Join(base, "not-a-dir")
		if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
			t.Fatalf("setup: %v", err)
		}
		w := NewJSONFile(filepath.Join(blocker, "search-index.json"))
		err := w.WriteIndex([]domain.SearchIndexItem{})
		if err == nil {
			t.Fatal("expected error when parent is a file, got nil")
		}
	})

	t.Run("returns error when path is an unwritable directory", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("running as root — permission bits are ignored")
		}
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o500); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
		// Points at dir/index.json where writing must fail.
		w := NewJSONFile(filepath.Join(dir, "index.json"))
		if err := w.WriteIndex([]domain.SearchIndexItem{}); err == nil {
			t.Fatal("expected write error on read-only dir, got nil")
		}
	})
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}
	return string(data)
}

func assertString(t *testing.T, m map[string]any, key, want string) {
	t.Helper()
	v, ok := m[key]
	if !ok {
		t.Errorf("missing key %q in JSON", key)
		return
	}
	got, ok := v.(string)
	if !ok {
		t.Errorf("key %q: got type %T, want string", key, v)
		return
	}
	if got != want {
		t.Errorf("key %q: got %q, want %q", key, got, want)
	}
}
