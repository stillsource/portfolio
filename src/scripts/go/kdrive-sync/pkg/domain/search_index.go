package domain

// SearchIndexItem is the compact JSON projection consumed by the client-side
// search UI at public/search-index.json.
type SearchIndexItem struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Date    string   `json:"date"`
	Tags    []string `json:"tags"`
	Poem    string   `json:"poem,omitempty"`
	Cover   string   `json:"cover,omitempty"`
	Palette []string `json:"palette"`
}
