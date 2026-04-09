package di

import "kdrive-sync/pkg/infrastructure/searchindexwriter"

func (c *Container) getSearchIndexWriter() *searchindexwriter.JSONFile {
	if c.indexWriter == nil {
		c.indexWriter = searchindexwriter.NewJSONFile(c.cfg.IndexFile)
	}
	return c.indexWriter
}
