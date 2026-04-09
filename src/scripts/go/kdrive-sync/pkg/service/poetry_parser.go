package service

import "kdrive-sync/pkg/domain"

// PoetryParser decodes a poetry markdown file into its global body and
// per-photo fragments.
type PoetryParser interface {
	Parse(data []byte) (domain.Poetry, error)
}
