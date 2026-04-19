package domain

import "strings"

// ExifData holds the camera metadata surfaced in the UI.
//
// Field names match the Astro content schema (src/types/content.ts) so the
// generated YAML frontmatter is consumed as-is.
type ExifData struct {
	Body        string `json:"body,omitempty" yaml:"body,omitempty"`
	Lens        string `json:"lens,omitempty" yaml:"lens,omitempty"`
	FocalLength string `json:"focalLength,omitempty" yaml:"focalLength,omitempty"`
	Aperture    string `json:"aperture,omitempty" yaml:"aperture,omitempty"`
	ISO         string `json:"iso,omitempty" yaml:"iso,omitempty"`
	Shutter     string `json:"shutter,omitempty" yaml:"shutter,omitempty"`
}

// IsZero reports whether all fields are empty.
func (e ExifData) IsZero() bool {
	return e == (ExifData{})
}

// Caption produces the pre-formatted "BODY • LENS • FOCAL • APERTURE • SHUTTER • ISO"
// string used as the lightbox/journal caption. Computed once at sync time so
// the front-end doesn't need a formatter.
func (e ExifData) Caption() string {
	parts := make([]string, 0, 6)
	for _, p := range []string{e.Body, e.Lens, e.FocalLength, e.Aperture, e.Shutter, e.ISO} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, " • ")
}
