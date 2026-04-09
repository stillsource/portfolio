package metadata

import (
	"fmt"
	"io"

	"github.com/evanoberholster/imagemeta"
	"github.com/evanoberholster/imagemeta/exif2"
)

// ExifData matches the exifSchema in src/types/content.ts
type ExifData struct {
	Shutter     string `json:"shutter,omitempty" yaml:"shutter,omitempty"`
	Aperture    string `json:"aperture,omitempty" yaml:"aperture,omitempty"`
	ISO         string `json:"iso,omitempty" yaml:"iso,omitempty"`
	Body        string `json:"body,omitempty" yaml:"body,omitempty"`
	Lens        string `json:"lens,omitempty" yaml:"lens,omitempty"`
	FocalLength string `json:"focalLength,omitempty" yaml:"focalLength,omitempty"`
}

// ExtractExif decodes EXIF metadata from the provided reader.
func ExtractExif(r io.ReadSeeker) (*ExifData, error) {
	m, err := imagemeta.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	exif, err := m.Exif()
	if err != nil {
		return nil, fmt.Errorf("failed to get EXIF: %w", err)
	}

	data := &ExifData{}

	if tag, err := exif.GetTag(exif2.ExposureTime); err == nil {
		data.Shutter = tag.String()
	}
	if tag, err := exif.GetTag(exif2.FNumber); err == nil {
		data.Aperture = fmt.Sprintf("f/%s", tag.String())
	}
	if tag, err := exif.GetTag(exif2.ISOSpeedRatings); err == nil {
		data.ISO = tag.String()
	}
	if tag, err := exif.GetTag(exif2.Model); err == nil {
		data.Body = tag.String()
	}
	if tag, err := exif.GetTag(exif2.LensModel); err == nil {
		data.Lens = tag.String()
	}
	if tag, err := exif.GetTag(exif2.FocalLength); err == nil {
		data.FocalLength = fmt.Sprintf("%smm", tag.String())
	}

	return data, nil
}

// HumanReadable returns a formatted string like "Fujifilm X-T4 • 23mm • f/2.0 • 1/500s"
func (e *ExifData) HumanReadable() string {
	res := e.Body
	if e.FocalLength != "" {
		if res != "" { res += " • " }
		res += e.FocalLength
	}
	if e.Aperture != "" {
		if res != "" { res += " • " }
		res += e.Aperture
	}
	if e.Shutter != "" {
		if res != "" { res += " • " }
		res += e.Shutter
	}
	return res
}
