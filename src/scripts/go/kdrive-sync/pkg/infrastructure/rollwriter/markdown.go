// Package rollwriter contains infrastructure adapters for service.RollWriter.
package rollwriter

import (
	"bytes"
	"fmt"
	"kdrive-sync/pkg/domain"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	dirPerm  os.FileMode = 0o755
	filePerm os.FileMode = 0o644
)

// Markdown writes a Roll as a YAML-frontmatter markdown file under OutDir,
// matching the format produced by the original fetch-kdrive.ts script.
type Markdown struct {
	outDir string
}

// NewMarkdown returns a Markdown writer rooted at outDir.
func NewMarkdown(outDir string) *Markdown {
	return &Markdown{outDir: outDir}
}

// WriteRoll serializes roll as a markdown file named slug.md.
func (m *Markdown) WriteRoll(slug string, roll *domain.Roll) error {
	if err := os.MkdirAll(m.outDir, dirPerm); err != nil {
		return fmt.Errorf("mkdir %s: %w", m.outDir, err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")

	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(roll); err != nil {
		return fmt.Errorf("encode roll %s: %w", slug, err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("close encoder for %s: %w", slug, err)
	}
	buf.WriteString("---\n")

	path := filepath.Join(m.outDir, slug+".md")
	if err := os.WriteFile(path, buf.Bytes(), filePerm); err != nil {
		return fmt.Errorf("write roll %s: %w", path, err)
	}
	return nil
}
