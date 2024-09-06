package main

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// getURLsFromHTML extracts all URLs from <a> tags in the provided HTML body
// and converts relative URLs to absolute URLs using the base URL.
func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
	var urls []string

	// Parse the base URL
	base, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, err
	}

	// Parse the HTML body
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil, err
	}

	// Traverse the HTML nodes
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					// Resolve relative URL to absolute URL
					parsedURL, parseErr := url.Parse(attr.Val)
					if parseErr == nil {
						absoluteURL := base.ResolveReference(parsedURL).String()
						urls = append(urls, absoluteURL)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return urls, nil
}
