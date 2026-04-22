package di

import "kdrive-sync/pkg/infrastructure/kdriveadapter"

func (c *Container) getFileDownloader() *kdriveadapter.FileDownloader {
	if c.fileDownloader == nil {
		c.fileDownloader = kdriveadapter.NewFileDownloader(c.getKDriveClient().Files)
	}
	return c.fileDownloader
}
