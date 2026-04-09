package markdown

import (
	"os"
	"path/filepath"
	"solenya-sync/internal/metadata"
	"strings"
	"testing"
)

func TestWriteRoll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "solenya-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	data := &RollData{
		Title: "Test Roll",
		Date:  "2026-04-09",
		Tags:  []string{"test", "photography"},
		Poem:  "A test poem\nWith multiple lines.",
		Palette: []string{"#ffffff", "#000000"},
		DominantColor: "#ffffff",
		Images: []ImageData{
			{
				URL: "https://example.com/img1.jpg",
				Exif: &metadata.ExifData{
					Body: "Fujifilm X-T4",
					ISO:  "400",
				},
				DominantColor: "#ffffff",
			},
		},
		Content: "This is the markdown content.",
	}

	err = WriteRoll(tmpDir, "test-roll", data)
	if err != nil {
		t.Fatalf("WriteRoll failed: %v", err)
	}

	// Verify file content
	content, err := os.ReadFile(filepath.Join(tmpDir, "test-roll.md"))
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	strContent := string(content)
	if !strings.Contains(strContent, "title: \"Test Roll\"") {
		t.Error("Generated YAML missing title")
	}
	if !strings.Contains(strContent, "date: 2026-04-09") {
		t.Error("Generated YAML missing or incorrect date")
	}
	if !strings.Contains(strContent, "Fujifilm X-T4") {
		t.Error("Generated YAML missing EXIF data")
	}
	if !strings.Contains(strContent, "This is the markdown content.") {
		t.Error("Generated file missing markdown content")
	}
	if !strings.HasPrefix(strContent, "---") || !strings.Contains(strContent, "---\nThis is the markdown content.") {
		t.Error("Generated file has incorrect format/delimiters")
	}
}
