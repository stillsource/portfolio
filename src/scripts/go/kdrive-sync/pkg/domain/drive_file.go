// Package domain defines the core business types: DriveFile, Roll, Image,
// ExifData, Poetry, and the helper types that compose them.
package domain

import "time"

// DriveFileType distinguishes folders from regular files on the drive.
type DriveFileType string

// Drive file type constants.
const (
	DriveFileTypeDir  DriveFileType = "dir"
	DriveFileTypeFile DriveFileType = "file"
)

// DriveFile represents a single entry returned by the drive API.
type DriveFile struct {
	ID        string
	Name      string
	Type      DriveFileType
	CreatedAt time.Time
}

// IsDir reports whether the entry is a directory.
func (f DriveFile) IsDir() bool {
	return f.Type == DriveFileTypeDir
}
