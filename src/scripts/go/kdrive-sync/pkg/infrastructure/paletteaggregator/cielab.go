// Package paletteaggregator contains infrastructure adapters for
// service.PaletteAggregator.
package paletteaggregator

import (
	"sort"

	"github.com/lucasb-eyer/go-colorful"
)

// CIELAB averages per-image palettes in the perceptually uniform CIELAB color
// space, mirroring the behaviour of calculateMeanPalette in the original
// TypeScript implementation.
type CIELAB struct{}

// NewCIELAB returns a stateless CIELAB aggregator.
func NewCIELAB() *CIELAB {
	return &CIELAB{}
}

type labAccumulator struct {
	l, a, b float64
	count   int
}

// Aggregate fuses the first `size` colors of every input palette into a single
// representative palette of length `size`.
//
// Colors are averaged component-wise in CIELAB space, then sorted from
// brightest to darkest so the dominant tone lands at index 0.
func (CIELAB) Aggregate(palettes [][]string, size int) []string {
	if size <= 0 || len(palettes) == 0 {
		return nil
	}

	buckets := make([]labAccumulator, size)
	for _, palette := range palettes {
		for i, hex := range palette {
			if i >= size {
				break
			}
			c, err := colorful.Hex(hex)
			if err != nil {
				continue
			}
			l, a, b := c.Lab()
			buckets[i].l += l
			buckets[i].a += a
			buckets[i].b += b
			buckets[i].count++
		}
	}

	averaged := make([]colorful.Color, 0, size)
	for _, bucket := range buckets {
		if bucket.count == 0 {
			continue
		}
		averaged = append(averaged, colorful.Lab(
			bucket.l/float64(bucket.count),
			bucket.a/float64(bucket.count),
			bucket.b/float64(bucket.count),
		))
	}

	sort.SliceStable(averaged, func(i, j int) bool {
		li, _, _ := averaged[i].Lab()
		lj, _, _ := averaged[j].Lab()
		return li > lj
	})

	out := make([]string, 0, len(averaged))
	for _, c := range averaged {
		out = append(out, c.Hex())
	}
	return out
}
