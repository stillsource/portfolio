package di

import "kdrive-sync/pkg/infrastructure/poetryparser"

func (c *Container) getPoetryParser() *poetryparser.Frontmatter {
	if c.poetryParser == nil {
		c.poetryParser = poetryparser.NewFrontmatter()
	}
	return c.poetryParser
}
