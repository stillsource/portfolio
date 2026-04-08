package kdrive

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// KDriveFile represents a file or directory in kDrive.
type KDriveFile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"` // "dir" or "file"
	CreatedAt int64  `json:"created_at"`
}

// KDriveResponse represents the standard API response for file listing.
type KDriveResponse struct {
	Data []KDriveFile `json:"data"`
}

// Client is a kDrive API client.
type Client struct {
	BaseURL    string
	DriveID    string
	APIToken   string
	HTTPClient *http.Client
}

// NewClient creates a new kDrive API client.
func NewClient(driveID, apiToken string) *Client {
	return &Client{
		BaseURL:  "https://api.infomaniak.com/2/drive",
		DriveID:  driveID,
		APIToken: apiToken,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetFiles retrieves the files and folders within a specific folder.
func (c *Client) GetFiles(folderID string) ([]KDriveFile, error) {
	endpoint := fmt.Sprintf("%s/%s/files/%s/files", c.BaseURL, c.DriveID, folderID)
	
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var kResp KDriveResponse
	if err := json.NewDecoder(resp.Body).Decode(&kResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return kResp.Data, nil
}
