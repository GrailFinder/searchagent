# Search Agent

A CLI tool that searches the web and returns content from top results.

## Features

- Search the web using a command-line interface
- Extract content from top search results
- Output results as JSON to stdout or file
- Configurable number of results to return

## Installation

```bash
go install .
```

## Usage

Basic usage:
```bash
searchagent "weather in london"
```

With options:
```bash
# Limit results to 5
searchagent -limit 5 "weather in london"

# Save results to a file
searchagent -output results.json "weather in london"
```

## Architecture

The tool uses an interface-based design that allows different search implementations:

- `Searcher` interface: Defines how to perform a search
- `WebScraper`: Implements search by scraping web results
- `SearchService`: Orchestrates the search process

## Limitations

Currently, the scraper implementation uses mock data as direct scraping of search engines
violates their Terms of Service. In a production implementation, you would want to
integrate with a legitimate search API.