package service

import "kdrive-sync/pkg/domain"

// ImageAnalyzer derives EXIF, tags, palette and dominant color from the raw
// bytes of a JPEG in one pass to avoid re-decoding.
type ImageAnalyzer interface {
	Analyze(data []byte) (domain.ImageAnalysis, error)
}
