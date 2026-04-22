package kdriveadapter

import (
	"context"
	"fmt"

	"github.com/stillsource/kdrive-fuse/kdrive"
)

// SharePublisher implements service.SharePublisher by wrapping kdrive.Shares.
type SharePublisher struct {
	client kdrive.Shares
}

// NewSharePublisher builds a SharePublisher around the given kdrive client.
func NewSharePublisher(c kdrive.Shares) *SharePublisher {
	return &SharePublisher{client: c}
}

// PublishShare returns the public share URL for fileID (creates one on demand).
func (a *SharePublisher) PublishShare(ctx context.Context, fileID string) (string, error) {
	id, err := parseID(fileID)
	if err != nil {
		return "", err
	}
	info, err := a.client.Publish(ctx, id)
	if err != nil {
		return "", fmt.Errorf("publish share %s: %w", fileID, err)
	}
	return info.ShareURL, nil
}
