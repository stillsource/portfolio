package di

import "kdrive-sync/pkg/infrastructure/imageanalyzer"

func (c *Container) getImageAnalyzer() *imageanalyzer.ExifKMeans {
	if c.imageAnalyzer == nil {
		c.imageAnalyzer = imageanalyzer.NewExifKMeans()
	}
	return c.imageAnalyzer
}
