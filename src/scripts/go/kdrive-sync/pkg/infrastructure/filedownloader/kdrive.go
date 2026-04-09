// Package filedownloader contains infrastructure adapters for service.FileDownloader.
package filedownloader

import (
	"context"
	"fmt"
	"io"

	"kdrive-sync/pkg/infrastructure/kdriveapi"
)

// KDrive downloads raw file bytes through the kDrive REST API.
type KDrive struct {
	client *kdriveapi.Client
}

// NewKDrive wires a downloader backed by the given API client.
func NewKDrive(client *kdriveapi.Client) *KDrive {
	return &KDrive{client: client}
}

// DownloadFile fetches the full body of fileID into memory.
func (k *KDrive) DownloadFile(ctx context.Context, fileID string) ([]byte, error) {
	endpoint := fmt.Sprintf("/files/%s/download", fileID)
	resp, err := k.client.Do(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", fileID, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", fileID, err)
	}
	return data, nil
}
