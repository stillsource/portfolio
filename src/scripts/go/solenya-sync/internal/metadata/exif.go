package metadata

import (
	"fmt"
	"io"

	"github.com/evanoberholster/imagemeta"
	"github.com/evanoberholster/imagemeta/exif2"
)

// ExifData matches the exifSchema in src/types/content.ts
type ExifData struct {
	Shutter     string `json:"shutter,omitempty"`
	Aperture    string `json:"aperture,omitempty"`
	ISO         string `json:"iso,omitempty"`
	Body        string `json:"body,omitempty"`
	Lens        string `json:"lens,omitempty"`
	FocalLength string `json:"focalLength,omitempty"`
}

// ExtractExif decodes EXIF metadata from the provided reader.
func ExtractExif(r io.ReadSeeker) (*ExifData, error) {
	e, err := imagemeta.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode EXIF: %w", err)
	}

	data := &ExifData{}

	if val, err := e.Get(exif2.ExposureTime); err == nil {
		data.Shutter = val.String()
	}
	if val, err := e.Get(exif2.FNumber); err == nil {
		data.Aperture = fmt.Sprintf("f/%s", val.String())
	}
	if val, err := e.Get(exif2.ISOSpeedRatings); err == nil {
		data.ISO = val.String()
	}
	if val, err := e.Get(exif2.Model); err == nil {
		data.Body = val.String()
	}
	if val, err := e.Get(exif2.LensModel); err == nil {
		data.Lens = val.String()
	}
	if val, err := e.Get(exif2.FocalLength); err == nil {
		data.FocalLength = fmt.Sprintf("%smm", val.String())
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
