package kdriveadapter

import (
	"context"
	"fmt"

	"github.com/stillsource/kdrive-fuse/kdrive"
)

// FileDownloader implements service.FileDownloader by wrapping kdrive.Files.
type FileDownloader struct {
	client kdrive.Files
}

// NewFileDownloader builds a FileDownloader around the given kdrive client.
func NewFileDownloader(c kdrive.Files) *FileDownloader {
	return &FileDownloader{client: c}
}

// DownloadFile fetches the full content of fileID into memory.
func (a *FileDownloader) DownloadFile(ctx context.Context, fileID string) ([]byte, error) {
	id, err := parseID(fileID)
	if err != nil {
		return nil, err
	}
	data, err := a.client.Download(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("download file %s: %w", fileID, err)
	}
	return data, nil
}
