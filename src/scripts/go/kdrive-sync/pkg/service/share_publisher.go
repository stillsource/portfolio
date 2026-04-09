package service

import "context"

// SharePublisher returns a stable public URL for a drive file, creating a
// share if none exists yet.
type SharePublisher interface {
	PublishShare(ctx context.Context, fileID string) (string, error)
}
