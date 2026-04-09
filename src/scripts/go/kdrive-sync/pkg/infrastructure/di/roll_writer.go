package di

import "kdrive-sync/pkg/infrastructure/rollwriter"

func (c *Container) getRollWriter() *rollwriter.Markdown {
	if c.rollWriter == nil {
		c.rollWriter = rollwriter.NewMarkdown(c.cfg.OutDir)
	}
	return c.rollWriter
}
