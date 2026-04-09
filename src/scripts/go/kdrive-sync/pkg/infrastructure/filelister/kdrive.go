// Package filelister contains infrastructure adapters for service.FileLister.
package filelister

import (
	"context"
	"fmt"
	"time"

	"kdrive-sync/pkg/domain"
	"kdrive-sync/pkg/infrastructure/kdriveapi"
)

// KDrive lists files through the Infomaniak kDrive REST API.
type KDrive struct {
	client *kdriveapi.Client
}

// NewKDrive wires a KDrive lister backed by the given API client.
func NewKDrive(client *kdriveapi.Client) *KDrive {
	return &KDrive{client: client}
}

type listResponse struct {
	Data []struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		Type      string `json:"type"`
		CreatedAt int64  `json:"created_at"`
	} `json:"data"`
}

// ListFiles fetches the direct children of folderID.
func (k *KDrive) ListFiles(ctx context.Context, folderID string) ([]domain.DriveFile, error) {
	endpoint := fmt.Sprintf("/files/%s/files", folderID)

	var resp listResponse
	if err := k.client.DecodeJSON(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("list files %s: %w", folderID, err)
	}

	files := make([]domain.DriveFile, len(resp.Data))
	for i, f := range resp.Data {
		files[i] = domain.DriveFile{
			ID:        fmt.Sprintf("%d", f.ID),
			Name:      f.Name,
			Type:      domain.DriveFileType(f.Type),
			CreatedAt: time.Unix(f.CreatedAt, 0),
		}
	}
	return files, nil
}
