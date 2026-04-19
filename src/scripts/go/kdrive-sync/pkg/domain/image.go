package domain

// Orientation classifies an image's aspect for front-end layout (lightbox,
// etc.). Pre-computed by the sync so the browser doesn't need to re-probe.
type Orientation string

const (
	OrientationLandscape Orientation = "landscape"
	OrientationPortrait  Orientation = "portrait"
)

// Image is a single photo inside a Roll as persisted in the Astro content file.
type Image struct {
	URL           string      `json:"url" yaml:"url"`
	Alt           string      `json:"alt,omitempty" yaml:"alt,omitempty"`
	Exif          *ExifData   `json:"exif,omitempty" yaml:"exif,omitempty"`
	Metadata      string      `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Poem          string      `json:"poem,omitempty" yaml:"poem,omitempty"`
	Palette       []string    `json:"palette,omitempty" yaml:"palette,omitempty"`
	DominantColor string      `json:"dominantColor,omitempty" yaml:"dominantColor,omitempty"`
	Width         int         `json:"width,omitempty" yaml:"width,omitempty"`
	Height        int         `json:"height,omitempty" yaml:"height,omitempty"`
	Orientation   Orientation `json:"orientation,omitempty" yaml:"orientation,omitempty"`
}
