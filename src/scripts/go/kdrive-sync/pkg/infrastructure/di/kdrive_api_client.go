package di

import "github.com/stillsource/kdrive-fuse/kdrive"

// getKDriveClient returns the shared kdrive library client for this drive.
func (c *Container) getKDriveClient() *kdrive.Client {
	if c.kdriveClient == nil {
		opts := []kdrive.Option{
			kdrive.WithHTTPClient(c.GetHTTPClient()),
			kdrive.WithLogger(c.GetLogger()),
		}
		if c.cfg.KDriveBaseURL != "" {
			opts = append(opts, kdrive.WithBaseURL(c.cfg.KDriveBaseURL))
		}
		c.kdriveClient = kdrive.New(c.cfg.APIToken, c.cfg.DriveID, opts...)
	}
	return c.kdriveClient
}
