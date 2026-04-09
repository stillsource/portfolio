// Package service declares the service interfaces that decouple the usecase
// layer from its infrastructure implementations.
package service

import "context"

// FileDownloader fetches the raw bytes of a single drive file.
type FileDownloader interface {
	DownloadFile(ctx context.Context, fileID string) ([]byte, error)
}
