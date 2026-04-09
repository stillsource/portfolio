package service

import (
	"context"
	"kdrive-sync/pkg/domain"
)

// FileLister enumerates the entries of a drive folder.
type FileLister interface {
	ListFiles(ctx context.Context, folderID string) ([]domain.DriveFile, error)
}
