package service

// PaletteAggregator fuses per-image palettes into a single representative
// palette for a Roll.
type PaletteAggregator interface {
	Aggregate(palettes [][]string, size int) []string
}
