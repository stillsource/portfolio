// Package kdriveadapter bridges the external kdrive client library to the
// portfolio's service.* interfaces. Each adapter translates between the
// portfolio's string IDs / domain types and the library's int64 IDs / kDrive types.
package kdriveadapter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	scerr "github.com/scality/go-errors"
	"github.com/stillsource/kdrive-fuse/kdrive"

	"kdrive-sync/pkg/domain"
)

// FileLister implements service.FileLister by wrapping kdrive.Files.
type FileLister struct {
	client kdrive.Files
}

// NewFileLister builds a FileLister around the given kdrive client.
func NewFileLister(c kdrive.Files) *FileLister {
	return &FileLister{client: c}
}

// ListFiles returns the direct children of folderID.
// folderID is a decimal string; the kdrive library uses int64s.
func (a *FileLister) ListFiles(ctx context.Context, folderID string) ([]domain.DriveFile, error) {
	id, err := parseID(folderID)
	if err != nil {
		return nil, err
	}
	infos, err := a.client.List(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list files %s: %w", folderID, err)
	}
	out := make([]domain.DriveFile, len(infos))
	for i, f := range infos {
		out[i] = domain.DriveFile{
			ID:        strconv.FormatInt(f.ID, 10),
			Name:      f.Name,
			Type:      domain.DriveFileType(f.Type),
			CreatedAt: time.Unix(f.CreatedAt, 0),
		}
	}
	return out, nil
}

// parseID parses a decimal string ID; returns a wrapped validation error on failure.
func parseID(s string) (int64, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, scerr.Wrap(kdrive.ErrValidation,
			scerr.WithDetailf("invalid id %q", s),
			scerr.CausedBy(err),
		)
	}
	return n, nil
}
