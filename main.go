package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
)

// isSameDomain checks if the current URL is on the same domain as the base URL.
func isSameDomain(baseURL, currentURL string) bool {
	base, err := url.Parse(baseURL)
	if err != nil {
		return false
	}
	current, err := url.Parse(currentURL)
	if err != nil {
		return false
	}
	return base.Host == current.Host
}

// crawlPage recursively crawls the website starting from rawCurrentURL.
func crawlPage(rawBaseURL, rawCurrentURL string, pages map[string]int, mu *sync.Mutex, depth int) {
	if depth > 5 { // Set a reasonable depth limit
		fmt.Printf("Skipping URL due to depth limit: %s\n", rawCurrentURL)
		return
	}

	fmt.Printf("Entering crawlPage with URL: %s (depth: %d)\n", rawCurrentURL, depth)

	normalizedURL, err := normalizeURL(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error normalizing URL: %v\n", err)
		return
	}

	mu.Lock()
	visited := false
	if count, exists := pages[normalizedURL]; exists {
		pages[normalizedURL] = count + 1
		visited = true
	}
	if !visited {
		pages[normalizedURL] = 1
	}
	mu.Unlock()

	if visited {
		fmt.Printf("Skipping URL (already visited): %s\n", normalizedURL)
		return
	}

	fmt.Printf("Visiting URL: %s\n", normalizedURL)

	htmlContent, err := getHTML(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error fetching HTML from %s: %v\n", rawCurrentURL, err)
		return
	}
	fmt.Println("Fetched HTML successfully.")

	urls, err := getURLsFromHTML(htmlContent, rawBaseURL)
	if err != nil {
		fmt.Printf("Error extracting URLs from HTML: %v\n", err)
		return
	}
	fmt.Printf("Extracted %d URLs from %s\n", len(urls), rawCurrentURL)

	// Release lock before making recursive calls
	fmt.Println("Unlocked, beginning recursive calls")

	for _, u := range urls {
		fmt.Printf("Found URL: %s\n", u)
		if isSameDomain(rawBaseURL, u) && u != rawCurrentURL {
			fmt.Printf("Recursively crawling URL: %s\n", u)
			go crawlPage(rawBaseURL, u, pages, mu, depth+1) // Use goroutine to avoid blocking
		} else {
			fmt.Printf("Skipping URL: %s\n", u)
		}
	}
	fmt.Printf("Exiting crawlPage with URL: %s\n", rawCurrentURL)
}

func getHTML(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch %s: status code %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body from %s: %v", url, err)
	}
	return string(body), nil
}

func main() {
	baseURL := ""

	// Define and parse command-line flags
	flag.Parse()
	args := flag.Args()

	// Check the number of arguments
	switch len(args) {
	case 0:
		fmt.Println("no website provided")
		os.Exit(1)
	case 1:
		baseURL = args[0]
		fmt.Printf("starting crawl of: %s\n", baseURL)
	default:
		fmt.Println("too many arguments provided")
		os.Exit(1)
	}

	pages := make(map[string]int)
	var mu sync.Mutex

	fmt.Println("Starting crawl...")
	crawlPage(baseURL, baseURL, pages, &mu, 0) // Start with depth 0

	fmt.Println("Crawl complete.")
	fmt.Println("Crawled Pages:")
	for url, count := range pages {
		fmt.Printf("%s: %d\n", url, count)
	}
}
