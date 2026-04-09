package di

import "kdrive-sync/pkg/infrastructure/sharepublisher"

func (c *Container) getSharePublisher() *sharepublisher.KDrive {
	if c.sharePublisher == nil {
		c.sharePublisher = sharepublisher.NewKDrive(c.getKDriveAPIClient())
	}
	return c.sharePublisher
}
