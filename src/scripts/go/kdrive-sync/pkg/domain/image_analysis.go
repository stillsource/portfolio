package domain

// ImageAnalysis is the bundle of derived data we extract from raw image bytes.
type ImageAnalysis struct {
	Tags          []string
	Exif          ExifData
	Palette       []string
	DominantColor string
	Width         int
	Height        int
}
