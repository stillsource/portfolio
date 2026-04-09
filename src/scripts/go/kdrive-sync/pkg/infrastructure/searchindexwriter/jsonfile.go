// Package searchindexwriter contains infrastructure adapters for
// service.SearchIndexWriter.
package searchindexwriter

import (
	"encoding/json"
	"fmt"
	"kdrive-sync/pkg/domain"
	"os"
	"path/filepath"
)

const (
	dirPerm  os.FileMode = 0o755
	filePerm os.FileMode = 0o644
)

// JSONFile persists the search index as a pretty-printed JSON file.
type JSONFile struct {
	path string
}

// NewJSONFile returns a writer targeting the given file path.
func NewJSONFile(path string) *JSONFile {
	return &JSONFile{path: path}
}

// WriteIndex encodes items to disk, creating parent directories as needed.
func (j *JSONFile) WriteIndex(items []domain.SearchIndexItem) error {
	if err := os.MkdirAll(filepath.Dir(j.path), dirPerm); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(j.path), err)
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal search index: %w", err)
	}

	if err := os.WriteFile(j.path, data, filePerm); err != nil {
		return fmt.Errorf("write search index %s: %w", j.path, err)
	}
	return nil
}
