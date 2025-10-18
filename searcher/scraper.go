package searcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// WebScraper implements the Searcher interface using web scraping
type WebScraper struct {
	client *http.Client
}

func NewWebScraper() *WebScraper {
	return &WebScraper{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (ws *WebScraper) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Attempt to perform a real search using Google Custom Search or similar
	// Since we don't have an API key in this implementation, let's use a basic technique
	// that searches and extracts results from HTML
	results := make([]SearchResult, 0, limit)
	// For now, let's implement a basic search that uses DuckDuckGo HTML search
	// which doesn't require an API key but is subject to rate limits and may break
	// if DuckDuckGo changes their HTML structure
	searchResults, err := ws.searchDuckDuckGo(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	// If no results found, return empty results
	if len(searchResults) == 0 {
		return []SearchResult{}, nil
	}
	for i, result := range searchResults {
		if i >= limit {
			break
		}
		results = append(results, result)
	}
	return results, nil
}

// searchDuckDuckGo performs a real search on DuckDuckGo and extracts results
func (ws *WebScraper) searchDuckDuckGo(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Encode the query for URL
	encodedQuery := strings.ReplaceAll(query, " ", "+")
	searchURL := "https://html.duckduckgo.com/html/?q=" + encodedQuery
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	// Add user agent and referer headers to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "https://duckduckgo.com/")
	resp, err := ws.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code error: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Parse the HTML to extract search results
	results := ws.parseDuckDuckGoResults(string(body), limit)
	// Extract content for each URL
	for i := range results {
		content, err := ws.extractContentFromURL(ctx, results[i].URL)
		if err != nil {
			// If we can't fetch content, keep the existing content
			continue
		}
		results[i].Content = content
	}
	return results, nil
}

// parseDuckDuckGoResults parses DuckDuckGo HTML results to extract search snippets
func (ws *WebScraper) parseDuckDuckGoResults(htmlContent string, limit int) []SearchResult {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return []SearchResult{} // Return empty results if parsing fails
	}
	var results []SearchResult
	var parse func(*html.Node)
	parse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			// Check if this is a search result container
			if ws.hasClass(n, "result") {
				result := ws.extractResultFromNode(n)
				if result.URL != "" && result.Title != "" { // Only add if both URL and Title are present
					results = append(results, result)
					if len(results) >= limit {
						return
					}
				}
			}
		}
		// Continue traversing children and siblings
		for c := n.FirstChild; c != nil && len(results) < limit; c = c.NextSibling {
			parse(c)
		}
	}
	parse(doc)
	return results
}

// extractResultFromNode extracts the title, URL and content from a search result node
func (ws *WebScraper) extractResultFromNode(n *html.Node) SearchResult {
	var result SearchResult
	var find func(*html.Node)
	find = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Look for the title link (usually has class 'result__a')
			if n.Data == "a" && ws.hasClass(n, "result__a") {
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						result.URL = attr.Val
						break
					}
				}
				// Get title text
				result.Title = ws.getTextContent(n)
			}
			// Look for the snippet (usually has class 'result__snippet')
			if n.Data == "a" && ws.hasClass(n, "result__snippet") {
				result.Content = ws.getTextContent(n)
			}
			// Look for the URL text (sometimes available in separate element)
			if n.Data == "a" && ws.hasClass(n, "result__url") {
				if result.URL == "" {
					// This is a fallback - in a real implementation we'd extract the URL from the href
					// For now we'll just use the text content as a placeholder
					result.URL = ws.getTextContent(n)
				}
			}
		}
		// Recursively look in children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			find(c)
		}
	}
	find(n)
	// If content is still empty, extract any relevant text from the description
	if result.Content == "" {
		var extractDesc func(*html.Node)
		extractDesc = func(n *html.Node) {
			if n.Type == html.TextNode {
				text := strings.TrimSpace(n.Data)
				if len(text) > len(result.Content) { // Take the longest text as the content
					result.Content = text
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if n.Data != "a" { // Avoid re-extracting from title links
					extractDesc(c)
				}
			}
		}
		extractDesc(n)
	}
	// Clean up the content
	result.Content = strings.TrimSpace(result.Content)
	if len(result.Content) > 300 { // Limit content length
		result.Content = result.Content[:300] + "..."
	}
	return result
}

// hasClass checks if an HTML node has a specific class
func (ws *WebScraper) hasClass(n *html.Node, class string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			classes := strings.Split(attr.Val, " ")
			for _, c := range classes {
				if strings.TrimSpace(c) == class {
					return true
				}
			}
		}
	}
	return false
}

// getTextContent extracts text content from an HTML node
func (ws *WebScraper) getTextContent(n *html.Node) string {
	var text string
	var extractText func(*html.Node)
	extractText = func(n *html.Node) {
		if n.Type == html.TextNode {
			text += n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}
	}
	extractText(n)
	return strings.TrimSpace(text)
}

// extractContentFromURL fetches and extracts meaningful text content from a webpage
func (ws *WebScraper) extractContentFromURL(ctx context.Context, pageURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return "", err
	}
	// Add a user agent to avoid being blocked by some sites
	req.Header.Set("User-Agent", "SearchAgent/1.0")
	resp, err := ws.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code error: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// Parse the HTML and extract text content
	content := extractTextFromHTML(string(body))
	// Limit the content to a reasonable size
	if len(content) > 2000 {
		content = content[:2000]
	}
	return content, nil
}

// extractTextFromHTML removes HTML tags and returns text content
func extractTextFromHTML(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return ""
	}
	var extractText func(*html.Node) string
	extractText = func(n *html.Node) string {
		if n.Type == html.TextNode {
			return n.Data
		}
		var text string
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			text += extractText(c)
		}
		// Add a space if the current node is a block element
		if n.Type == html.ElementNode {
			switch n.Data {
			case "p", "div", "h1", "h2", "h3", "h4", "h5", "h6", "br", "li", "tr", "td":
				text += " "
			}
		}
		return text
	}
	text := extractText(doc)
	// Clean up extra whitespace
	text = strings.Join(strings.Fields(text), " ")
	return text
}
