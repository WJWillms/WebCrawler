package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"sync"
)

type config struct {
	pages              map[string]int
	maxPages           int
	baseURL            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
}

type pageInfo struct {
	URL   string
	Count int
}

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

func NewConfig(baseURL *url.URL, maxConcurrency int, maxPages int) *config {
	return &config{
		pages:              make(map[string]int),
		maxPages:           maxPages,
		baseURL:            baseURL,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxConcurrency),
		wg:                 &sync.WaitGroup{},
	}
}

func (cfg *config) addPageVisit(normalizedURL string) (isFirst bool) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if _, exists := cfg.pages[normalizedURL]; exists {
		return false // URL has already been visited
	}

	cfg.pages[normalizedURL] = 1
	return true // First time visiting this URL
}

// crawlPage recursively crawls the website starting from rawCurrentURL.
func (cfg *config) crawlPage(rawCurrentURL string) {

	cfg.concurrencyControl <- struct{}{}
	defer func() {
		<-cfg.concurrencyControl
		cfg.wg.Done()
	}()

	if cfg.pageCount() >= cfg.maxPages {
		return
	}

	fmt.Printf("Entering crawlPage with URL: %s\n", rawCurrentURL)

	normalizedURL, err := normalizeURL(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error normalizing URL: %v\n", err)
		return
	}

	fmt.Printf("Normalized URL: %s\n", normalizedURL)

	isFirst := cfg.addPageVisit(normalizedURL)
	if !isFirst {
		fmt.Printf("Skipping URL (already visited): %s\n", normalizedURL)
		return
	}

	fmt.Printf("Fetching HTML for URL: %s\n", normalizedURL)
	htmlContent, err := getHTML(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error fetching HTML from %s: %v\n", rawCurrentURL, err)
		return
	}

	fmt.Println("Fetched HTML successfully.")

	urls, err := getURLsFromHTML(htmlContent, cfg.baseURL.String())
	if err != nil {
		fmt.Printf("Error extracting URLs from HTML: %v\n", err)
		return
	}

	fmt.Printf("Extracted %d URLs from %s\n", len(urls), rawCurrentURL)

	for _, u := range urls {
		fmt.Printf("Found URL: %s\n", u)
		if isSameDomain(cfg.baseURL.String(), u) && u != rawCurrentURL {
			fmt.Printf("Recursively crawling URL: %s\n", u)
			cfg.wg.Add(1)
			go cfg.crawlPage(u)
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

func (cfg *config) pageCount() int {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	return len(cfg.pages)
}

func printReport(pages map[string]int, baseURL string) {
	var pageList []pageInfo
	for url, count := range pages {
		pageList = append(pageList, pageInfo{URL: url, Count: count})
	}

	sort.Slice(pageList, func(i, j int) bool {
		if pageList[i].Count == pageList[j].Count {
			return pageList[i].URL < pageList[j].URL
		}
		return pageList[i].Count > pageList[j].Count
	})

	fmt.Println("=============================")
	fmt.Printf("  REPORT for %s\n", baseURL)
	fmt.Println("=============================")

	for _, page := range pageList {
		fmt.Printf("Found %d internal links to %s\n", page.Count, page.URL)
	}
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: ./crawler <baseURL> <maxConcurrency> <maxPages> ")
		return
	}

	baseURLStr := os.Args[1]
	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		fmt.Printf("Invalid URL: %v\n", err)
		return
	}

	maxConcurrencyStr := os.Args[2]
	maxPageStr := os.Args[3]

	maxConcurrency, err := strconv.Atoi(maxConcurrencyStr)
	if err != nil {
		fmt.Println("Error converting string to int:", err)
	}

	maxPages, err := strconv.Atoi(maxPageStr)
	if err != nil {
		fmt.Println("Error converting string to int:", err)
	}

	cfg := NewConfig(baseURL, maxConcurrency, maxPages)

	fmt.Println("starting crawl of:", baseURLStr)
	cfg.wg.Add(1)
	go cfg.crawlPage(baseURLStr)

	cfg.wg.Wait()
	fmt.Println("Crawl complete.")
	//fmt.Println("Crawled Pages:")
	//for page, count := range cfg.pages {
	//fmt.Printf("%s: %d\n", page, count)
	//}
	printReport(cfg.pages, baseURLStr)
}
