package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"solenya-sync/internal/metadata"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// RollData matches the rollSchema in src/types/content.ts
type RollData struct {
	Title         string      `yaml:"title"`
	Date          time.Time   `yaml:"date"`
	Tags          []string    `yaml:"tags,omitempty"`
	Poem          string      `yaml:"poem,omitempty"`
	Palette       []string    `yaml:"palette,omitempty"`
	DominantColor string      `yaml:"dominantColor,omitempty"`
	AudioURL      string      `yaml:"audioUrl,omitempty"`
	Images        []ImageData `yaml:"images"`
	Content       string      `yaml:"-"` // Not in frontmatter
}

// ImageData matches the image object in rollSchema
type ImageData struct {
	URL           string             `yaml:"url"`
	Exif          *metadata.ExifData `yaml:"exif,omitempty"`
	Metadata      string             `yaml:"metadata,omitempty"`
	Poem          string             `yaml:"poem,omitempty"`
	Palette       []string           `yaml:"palette,omitempty"`
	DominantColor string             `yaml:"dominantColor,omitempty"`
}

// WriteRoll serializes the roll data into a Markdown file with YAML frontmatter.
func WriteRoll(outDir string, slug string, data *RollData) error {
	// Ensure directory exists
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Prepare YAML frontmatter
	frontmatter, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	// Construct file content
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(frontmatter)
	sb.WriteString("---\n")
	if data.Content != "" {
		sb.WriteString(data.Content)
		if !strings.HasSuffix(data.Content, "\n") {
			sb.WriteString("\n")
		}
	}

	// Write to file
	filePath := filepath.Join(outDir, slug+".md")
	if err := os.WriteFile(filePath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
