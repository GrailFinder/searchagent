package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"log/slog"
	"os"
	"strings"

	"searchagent/config"
)

func main() {
	// Define command line flags
	outputFile := flag.String("output", "", "Output file to save results (default: stdout)")
	limit := flag.Int("limit", 3, "Maximum number of results to return")
	searchType := flag.String("type", "scraper", "Search type: scraper or api")
	serverMode := flag.Bool("server", false, "Run in server mode")
	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	if *serverMode {
		// Load configuration
		cfg := config.LoadConfigOrDefault(*configPath)

		// Create and start the server
		server := NewServer(cfg)
		if err := server.Start(cfg.ServerPort); err != nil {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	} else {
		// Get the search query from command line arguments
		if len(flag.Args()) == 0 {
			log.Fatal("Usage: searchagent [options] <search query>")
		}

		query := strings.Join(flag.Args(), " ")

		// Initialize the searcher based on type
		var searcher Searcher
		switch *searchType {
		case "api":
			searcher = NewSearchService(SearcherTypeAPI)
		case "scraper":
			fallthrough
		default:
			searcher = NewWebScraper()
		}

		// Perform the search
		ctx := context.Background()
		results, err := searcher.Search(ctx, query, *limit)
		if err != nil {
			log.Fatalf("Search error: %v", err)
		}

		// Format results as a map [page_link: content]
		resultsMap := make(map[string]string)
		for _, result := range results {
			resultsMap[result.URL] = result.Content
		}

		// Output the results
		if *outputFile != "" {
			// Save to file
			file, err := os.Create(*outputFile)
			if err != nil {
				log.Fatalf("Error creating output file: %v", err)
			}
			defer file.Close()

			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(resultsMap); err != nil {
				log.Fatalf("Error encoding JSON: %v", err)
			}
		} else {
			// Output to stdout
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(resultsMap); err != nil {
				log.Fatalf("Error encoding JSON: %v", err)
			}
		}
	}
}
