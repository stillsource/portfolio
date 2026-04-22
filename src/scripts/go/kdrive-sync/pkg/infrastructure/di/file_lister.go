package di

import "kdrive-sync/pkg/infrastructure/kdriveadapter"

func (c *Container) getFileLister() *kdriveadapter.FileLister {
	if c.fileLister == nil {
		c.fileLister = kdriveadapter.NewFileLister(c.getKDriveClient().Files)
	}
	return c.fileLister
}
