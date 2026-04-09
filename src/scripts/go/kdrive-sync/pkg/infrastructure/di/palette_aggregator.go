package di

import "kdrive-sync/pkg/infrastructure/paletteaggregator"

func (c *Container) getPaletteAggregator() *paletteaggregator.CIELAB {
	if c.paletteAgg == nil {
		c.paletteAgg = paletteaggregator.NewCIELAB()
	}
	return c.paletteAgg
}
