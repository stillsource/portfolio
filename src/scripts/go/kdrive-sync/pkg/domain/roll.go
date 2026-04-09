package domain

// Roll is a coherent series of photos captured at the same time/place.
//
// It mirrors the Astro content schema defined in src/content.config.ts.
// Field order in YAML is controlled by the struct layout.
type Roll struct {
	Title         string   `yaml:"title"`
	Date          string   `yaml:"date"`
	Tags          []string `yaml:"tags"`
	Poem          string   `yaml:"poem,omitempty"`
	Palette       []string `yaml:"palette,omitempty"`
	DominantColor string   `yaml:"dominantColor,omitempty"`
	AudioURL      string   `yaml:"audioUrl,omitempty"`
	Images        []Image  `yaml:"images"`
}
