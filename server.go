package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"searchagent/config"
)

type SearchRequest struct {
	Query      string `json:"query"`
	SearchType string `json:"search_type"`
	NumResults int    `json:"num_results"`
}

type ServerSearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

type SearchResponse struct {
	Query      string            `json:"query"`
	Results    []ServerSearchResult `json:"results"`
	Timestamp  time.Time         `json:"timestamp"`
	TotalCount int               `json:"total_count"`
}

// searchHandler handles incoming search requests
func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SearchRequest

	if r.Method == http.MethodPost {
		// Parse JSON request body
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&req); err != nil {
			http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
			return
		}
	} else if r.Method == http.MethodGet {
		// Parse query parameters from URL
		req.Query = r.URL.Query().Get("q")
		req.SearchType = r.URL.Query().Get("type")
		numResultsStr := r.URL.Query().Get("num")
		if numResultsStr != "" {
			numResults, err := strconv.Atoi(numResultsStr)
			if err != nil {
				http.Error(w, "Invalid num_results parameter", http.StatusBadRequest)
				return
			}
			req.NumResults = numResults
		}
	}

	// Set defaults if not provided
	if req.SearchType == "" {
		req.SearchType = "general" // Default to general search
	}
	if req.NumResults <= 0 {
		req.NumResults = 10 // Default number of results
	}

	if req.Query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	// Perform the search using the existing functionality
	results, err := s.Search(req.Query, req.SearchType, req.NumResults)
	if err != nil {
		slog.Error("Search failed", "error", err)
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := SearchResponse{
		Query:      req.Query,
		Results:    make([]ServerSearchResult, len(results)),
		Timestamp:  time.Now(),
		TotalCount: len(results),
	}

	for i, result := range results {
		response.Results[i] = ServerSearchResult{
			Title:   result.Title,
			URL:     result.URL,
			Content: result.Content,
		}
	}

	// Set content type and encode response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}
}

// Server represents the HTTP server
type Server struct {
	config *config.Config
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config) *Server {
	return &Server{
		config: cfg,
	}
}

// Search performs a search with the given parameters
func (s *Server) Search(query string, searchType string, numResults int) ([]SearchResult, error) {
	var searcher Searcher
	switch searchType {
	case "api":
		searcher = NewSearXNGAPISearcher("config.toml") // Use the API searcher directly
	case "scraper":
		fallthrough
	default:
		searcher = NewWebScraper()
	}

	ctx := context.Background()
	results, err := searcher.Search(ctx, query, numResults)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// Start starts the HTTP server
func (s *Server) Start(port int) error {
	http.HandleFunc("/search", s.searchHandler)

	addr := fmt.Sprintf(":%d", port)
	slog.Info("Starting server", "address", addr)
	
	return http.ListenAndServe(addr, nil)
}