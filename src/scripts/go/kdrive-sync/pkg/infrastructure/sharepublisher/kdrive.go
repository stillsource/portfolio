// Package sharepublisher contains infrastructure adapters for service.SharePublisher.
package sharepublisher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"kdrive-sync/pkg/infrastructure/kdriveapi"
)

// KDrive returns a public share URL for a kDrive file, creating one on
// demand if none exists yet.
type KDrive struct {
	client *kdriveapi.Client
}

// NewKDrive wires a publisher backed by the given API client.
func NewKDrive(client *kdriveapi.Client) *KDrive {
	return &KDrive{client: client}
}

type shareEntry struct {
	ShareURL string `json:"share_url"`
}

type shareListResponse struct {
	Data []shareEntry `json:"data"`
}

type shareCreateResponse struct {
	Data shareEntry `json:"data"`
}

// PublishShare returns the first existing public link for fileID, or creates
// a non-password-protected, non-expiring share if there is none.
func (k *KDrive) PublishShare(ctx context.Context, fileID string) (string, error) {
	endpoint := fmt.Sprintf("/files/%s/shares", fileID)

	var existing shareListResponse
	if err := k.client.DecodeJSON(ctx, "GET", endpoint, nil, &existing); err == nil {
		if len(existing.Data) > 0 && existing.Data[0].ShareURL != "" {
			return existing.Data[0].ShareURL, nil
		}
	}

	payload, err := json.Marshal(map[string]any{
		"type":               "public",
		"password_protected": false,
		"expiration_date":    0,
	})
	if err != nil {
		return "", fmt.Errorf("marshal share payload: %w", err)
	}

	var created shareCreateResponse
	if err := k.client.DecodeJSON(ctx, "POST", endpoint, payload, &created); err != nil {
		return "", fmt.Errorf("create share %s: %w", fileID, err)
	}
	if created.Data.ShareURL == "" {
		return "", errors.New("kdrive: empty share url")
	}
	return created.Data.ShareURL, nil
}
