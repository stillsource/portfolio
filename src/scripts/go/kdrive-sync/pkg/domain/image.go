package domain

// Image is a single photo inside a Roll as persisted in the Astro content file.
type Image struct {
	URL           string    `json:"url" yaml:"url"`
	Alt           string    `json:"alt,omitempty" yaml:"alt,omitempty"`
	Exif          *ExifData `json:"exif,omitempty" yaml:"exif,omitempty"`
	Poem          string    `json:"poem,omitempty" yaml:"poem,omitempty"`
	Palette       []string  `json:"palette,omitempty" yaml:"palette,omitempty"`
	DominantColor string    `json:"dominantColor,omitempty" yaml:"dominantColor,omitempty"`
}
