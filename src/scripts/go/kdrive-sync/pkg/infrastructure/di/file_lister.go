package di

import "kdrive-sync/pkg/infrastructure/filelister"

func (c *Container) getFileLister() *filelister.KDrive {
	if c.fileLister == nil {
		c.fileLister = filelister.NewKDrive(c.getKDriveAPIClient())
	}
	return c.fileLister
}
