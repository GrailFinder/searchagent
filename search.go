package main

import (
	"net/url"
)

func NewSearchService(t SearcherType) Searcher {
	switch t {
	case SearcherTypeScraper:
		return NewWebScraper()
	case SearcherTypeAPI:
		return NewSearXNGAPISearcher("config.toml") // Use config.toml for API endpoint
	default:
		panic("not known searcher type")
	}
}

// isValidURL checks if a string is a valid URL
func isValidURL(input string) bool {
	_, err := url.ParseRequestURI(input)
	return err == nil
}

