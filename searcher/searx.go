package searcher

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SearXNGAPISearcher implements the Searcher interface using the SearXNG API
type SearXNGAPISearcher struct {
	client  *http.Client
	baseURL string
}

// SearXNGResult represents a single search result from the SearXNG API
type SearXNGResult struct {
	Title         string `json:"title"`
	URL           string `json:"url"`
	Content       string `json:"content"`
	Engine        string `json:"engine"`
	PublishedDate string `json:"publishedDate"`
}

// SearXNGResponse represents the response structure from the SearXNG API
type SearXNGResponse struct {
	Results []SearXNGResult `json:"results"`
}

// NewSearXNGAPISearcher creates a new instance of SearXNGAPISearcher
// Uses the configuration from config.toml for the API endpoint
func NewSearXNGAPISearcher(baseURL string) *SearXNGAPISearcher {
	// Load the configuration
	// Ensure the base URL ends with a slash
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return &SearXNGAPISearcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

func (s *SearXNGAPISearcher) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Try the API endpoint first, then fall back to /search if needed
	endpoints := []string{"/api/v1/search", "/search"}
	var apiResponse SearXNGResponse

	for _, endpoint := range endpoints {
		// Build the API URL
		apiURL := fmt.Sprintf("%s%s", s.baseURL, strings.TrimPrefix(endpoint, "/"))

		// Create URL parameters
		params := url.Values{}
		params.Set("q", query)
		params.Set("format", "json")

		// Note: SearXNG API doesn't have a direct limit parameter in URL by default,
		// so we'll fetch results and limit them after parsing

		// Construct the full URL with parameters
		fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

		// Create the HTTP request
		req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
		if err != nil {
			continue // Try next endpoint
		}

		// Add headers to avoid being blocked - using more realistic browser-like headers
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
		req.Header.Set("Accept", "application/json, */*;q=0.1") // Prioritize JSON response
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req.Header.Set("Accept-Encoding", "gzip, deflate")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Upgrade-Insecure-Requests", "1")

		// Execute the request
		resp, err := s.client.Do(req)
		if err != nil {
			continue // Try next endpoint
		}

		// Read the response body
		var body []byte
		if resp.Header.Get("Content-Encoding") == "gzip" {
			// Handle gzipped response
			gzipReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				resp.Body.Close()
				continue // Try next endpoint
			}
			body, err = io.ReadAll(gzipReader)
			gzipReader.Close()
			if err != nil {
				resp.Body.Close()
				continue // Try next endpoint
			}
		} else {
			body, err = io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue // Try next endpoint
			}
		}

		// Try to parse the JSON response
		if err := json.Unmarshal(body, &apiResponse); err != nil {
			continue // Try next endpoint
		} else {
			// Successfully parsed JSON, break the loop
			break
		}
	}

	if len(apiResponse.Results) == 0 {
		return nil, errors.New("no valid JSON response from any endpoint")
	}

	// Convert the API results to our SearchResult format
	// Limit results after fetching from API
	results := make([]SearchResult, 0, limit)
	for _, result := range apiResponse.Results {
		if len(results) >= limit {
			break
		}

		// Skip results with empty title or URL
		if result.Title == "" || result.URL == "" {
			continue
		}

		results = append(results, SearchResult{
			URL:     result.URL,
			Title:   result.Title,
			Content: result.Content,
		})
	}

	return results, nil
}
