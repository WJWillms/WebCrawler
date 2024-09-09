package main

import (
	"net/url"
	"strings"
)

// normalizeURL normalizes a URL string.
func normalizeURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsedURL.Fragment = ""
	parsedURL.RawQuery = ""
	return strings.ToLower(parsedURL.String()), nil
}
