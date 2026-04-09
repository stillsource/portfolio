package di

import "kdrive-sync/pkg/infrastructure/filedownloader"

func (c *Container) getFileDownloader() *filedownloader.KDrive {
	if c.fileDownloader == nil {
		c.fileDownloader = filedownloader.NewKDrive(c.getKDriveAPIClient())
	}
	return c.fileDownloader
}
