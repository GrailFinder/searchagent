package searcher

import (
	"context"
	"net/url"
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
// Returns nil if the type is not recognized.
func NewSearchService(t SearcherType) Searcher {
	switch t {
	case SearcherTypeScraper:
		return NewWebScraper()
	case SearcherTypeAPI:
		return NewSearXNGAPISearcher("config.toml") // Use config.toml for API endpoint
	default:
		return nil
	}
}

// isValidURL checks if a string is a valid URL
func isValidURL(input string) bool {
	_, err := url.ParseRequestURI(input)
	return err == nil
}
