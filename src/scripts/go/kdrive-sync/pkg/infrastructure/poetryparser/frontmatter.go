// Package poetryparser contains infrastructure adapters for service.PoetryParser.
package poetryparser

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"kdrive-sync/pkg/domain"
)

// Frontmatter parses poetry markdown files that optionally carry a YAML
// frontmatter block of the form:
//
//	---
//	photos:
//	  IMG_0001.jpg: "First verse"
//	---
//	Global poem body...
//
// This mirrors the gray-matter behaviour of the original TypeScript sync.
type Frontmatter struct{}

// NewFrontmatter returns a stateless frontmatter-aware parser.
func NewFrontmatter() *Frontmatter {
	return &Frontmatter{}
}

var delimiter = []byte("---")

// Parse decodes data into a Poetry value. A missing frontmatter block is
// treated as a plain markdown body (GlobalPoem).
func (Frontmatter) Parse(data []byte) (domain.Poetry, error) {
	body, meta, err := splitFrontmatter(data)
	if err != nil {
		return domain.Poetry{}, err
	}

	poetry := domain.Poetry{
		GlobalPoem: strings.TrimSpace(string(body)),
		PhotoPoems: map[string]string{},
	}

	if len(meta) == 0 {
		return poetry, nil
	}

	var fm struct {
		Photos map[string]string `yaml:"photos"`
	}
	if err := yaml.Unmarshal(meta, &fm); err != nil {
		return poetry, fmt.Errorf("parse poetry frontmatter: %w", err)
	}
	if fm.Photos != nil {
		poetry.PhotoPoems = fm.Photos
	}
	return poetry, nil
}

// splitFrontmatter separates an optional YAML frontmatter block from the
// markdown body. It returns (body, frontmatter, error).
func splitFrontmatter(data []byte) ([]byte, []byte, error) {
	trimmed := bytes.TrimLeft(data, " \t\r\n")
	if !bytes.HasPrefix(trimmed, delimiter) {
		return data, nil, nil
	}

	rest := trimmed[len(delimiter):]
	// Expect a newline right after the opening delimiter.
	nl := bytes.IndexByte(rest, '\n')
	if nl < 0 {
		return data, nil, nil
	}
	rest = rest[nl+1:]

	end := bytes.Index(rest, append([]byte{'\n'}, delimiter...))
	if end < 0 {
		return data, nil, nil
	}
	meta := rest[:end]
	body := rest[end+len(delimiter)+1:]
	return body, meta, nil
}
