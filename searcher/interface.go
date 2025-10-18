package searcher

import (
	"context"
	"fmt"
)

// SearcherType represents the type of searcher to use
type SearcherType string

const (
	SearcherTypeScraper SearcherType = "SearcherTypeScraper"
	SearcherTypeAPI     SearcherType = "SearcherTypeAPI"
)

// SearchResult represents the content of a webpage
type SearchResult struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Searcher defines the interface for different search implementations
type Searcher interface {
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
}

// NewSearchService creates a new search service based on the provided type.
// Returns an error if the type is not recognized.
func NewSearchService(t SearcherType, url string) (Searcher, error) {
	// url: there must be a better way
	switch t {
	case SearcherTypeScraper:
		if url == "" {
			url = "https://html.duckduckgo.com/html/?q="
		}
		return NewWebScraper(url), nil
	case SearcherTypeAPI:
		if url == "" {
			url = "https://searx.grailfinder.net/"
		}
		return NewSearXNGAPISearcher(url), nil // Use config.toml for API endpoint
	default:
		return nil, fmt.Errorf("unknown searcher type: %s", t)
	}
}
