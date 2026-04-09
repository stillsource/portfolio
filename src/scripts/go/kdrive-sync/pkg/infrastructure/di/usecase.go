package di

import "kdrive-sync/pkg/usecase"

// GetSyncRolls returns the wired SyncRolls usecase. Calling this method is
// the only thing main.go has to know about the object graph.
func (c *Container) GetSyncRolls() *usecase.SyncRolls {
	if c.syncRolls == nil {
		c.syncRolls = usecase.NewSyncRolls(usecase.SyncRollsDeps{
			Logger:      c.GetLogger(),
			Lister:      c.getFileLister(),
			Downloader:  c.getFileDownloader(),
			Publisher:   c.getSharePublisher(),
			Analyzer:    c.getImageAnalyzer(),
			Aggregator:  c.getPaletteAggregator(),
			Poetry:      c.getPoetryParser(),
			RollWriter:  c.getRollWriter(),
			IndexWriter: c.getSearchIndexWriter(),
			Concurrency: c.cfg.Concurrency,
			PaletteSize: c.cfg.PaletteSize,
		})
	}
	return c.syncRolls
}
