package di

import "kdrive-sync/pkg/infrastructure/kdriveadapter"

func (c *Container) getSharePublisher() *kdriveadapter.SharePublisher {
	if c.sharePublisher == nil {
		c.sharePublisher = kdriveadapter.NewSharePublisher(c.getKDriveClient().Shares)
	}
	return c.sharePublisher
}
