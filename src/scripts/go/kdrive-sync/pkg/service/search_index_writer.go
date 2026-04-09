package service

import "kdrive-sync/pkg/domain"

// SearchIndexWriter persists the global search index consumed by the UI.
type SearchIndexWriter interface {
	WriteIndex(items []domain.SearchIndexItem) error
}
