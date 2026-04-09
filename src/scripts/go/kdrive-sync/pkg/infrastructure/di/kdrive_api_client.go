package di

import "kdrive-sync/pkg/infrastructure/kdriveapi"

// getKDriveAPIClient returns the shared low-level HTTP client for the
// Infomaniak kDrive API.
func (c *Container) getKDriveAPIClient() *kdriveapi.Client {
	if c.apiClient == nil {
		c.apiClient = kdriveapi.NewClient(
			c.GetHTTPClient(),
			c.GetLogger(),
			c.cfg.DriveID,
			c.cfg.APIToken,
			kdriveapi.Options{BaseURL: c.cfg.KDriveBaseURL},
		)
	}
	return c.apiClient
}
